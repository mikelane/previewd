// Copyright 2025 The Previewd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	previewv1alpha1 "github.com/mikelane/previewd/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Server handles GitHub webhook requests
type Server struct {
	client        client.Client
	server        *http.Server
	rateLimiter   *RateLimiter
	addr          string
	webhookSecret string
	port          int
}

// RateLimiter provides per-repository rate limiting
type RateLimiter struct {
	limiters map[string]*bucket
	window   time.Duration
	maxAge   time.Duration // Max age before cleaning up inactive buckets
	mu       sync.Mutex
	limit    int
}

type bucket struct {
	lastReset time.Time
	lastUsed  time.Time
	tokens    int
}

// NewServer creates a new webhook server
func NewServer(addr string, port int, k8sClient client.Client, webhookSecret string) *Server {
	return &Server{
		addr:          addr,
		port:          port,
		client:        k8sClient,
		webhookSecret: webhookSecret,
		rateLimiter:   NewRateLimiter(10, time.Second), // 10 requests per second per repo
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*bucket),
		limit:    limit,
		window:   window,
		maxAge:   1 * time.Hour, // Clean up buckets inactive for 1 hour
	}
}

// Allow checks if a request from the given repository should be allowed
func (rl *RateLimiter) Allow(repo string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Cleanup inactive buckets to prevent memory leak
	rl.cleanup()

	b, exists := rl.limiters[repo]
	if !exists {
		b = &bucket{
			tokens:    rl.limit,
			lastReset: time.Now(),
			lastUsed:  time.Now(),
		}
		rl.limiters[repo] = b
	}

	// Reset bucket if window has passed
	if time.Since(b.lastReset) >= rl.window {
		b.tokens = rl.limit
		b.lastReset = time.Now()
	}

	// Update last used time
	b.lastUsed = time.Now()

	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// cleanup removes buckets that haven't been used recently
// Must be called with mu locked
func (rl *RateLimiter) cleanup() {
	now := time.Now()
	for repo, b := range rl.limiters {
		if now.Sub(b.lastUsed) > rl.maxAge {
			delete(rl.limiters, repo)
		}
	}
}

// Start starts the webhook server
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", s.handleWebhook)
	mux.HandleFunc("/healthz", s.handleHealth)

	s.server = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", s.addr, s.port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Log.Info("Starting webhook server", "addr", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		return s.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	log.Log.Info("Shutting down webhook server")
	return s.server.Shutdown(ctx)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Log.Error(err, "Failed to write health check response")
	}
}

// handleWebhook handles GitHub webhook requests
func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())

	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error(err, "Failed to read request body")
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			logger.Error(closeErr, "Failed to close request body")
		}
	}()

	// Validate signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if !ValidateSignature(payload, signature, s.webhookSecret) {
		logger.Info("Invalid webhook signature")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Check event type
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType != "pull_request" {
		logger.V(1).Info("Ignoring non-PR event", "event", eventType)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse event
	var event PullRequestEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.Error(err, "Failed to parse JSON payload")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Rate limiting check
	if !s.rateLimiter.Allow(event.Repository.FullName) {
		logger.Info("Rate limit exceeded", "repository", event.Repository.FullName)
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	// Handle event
	ctx := r.Context()
	switch strings.ToLower(event.Action) {
	case "opened", "reopened":
		if err := s.handlePROpened(ctx, &event); err != nil {
			logger.Error(err, "Failed to handle PR opened")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)

	case "closed":
		if err := s.handlePRClosed(ctx, &event); err != nil {
			logger.Error(err, "Failed to handle PR closed")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "synchronize":
		if err := s.handlePRSynchronized(ctx, &event); err != nil {
			logger.Error(err, "Failed to handle PR synchronized")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	default:
		logger.V(1).Info("Ignoring PR action", "action", event.Action)
		w.WriteHeader(http.StatusOK)
	}
}

// handlePROpened creates a PreviewEnvironment CR when a PR is opened
func (s *Server) handlePROpened(ctx context.Context, event *PullRequestEvent) error {
	logger := log.FromContext(ctx)

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("pr-%d", event.Number),
			Namespace: "previewd-system",
			Labels: map[string]string{
				"previewd.io/pr":         fmt.Sprintf("%d", event.Number),
				"previewd.io/repository": sanitizeLabel(event.Repository.FullName),
			},
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			Repository: event.Repository.FullName,
			PRNumber:   event.Number,
			HeadSHA:    event.PullRequest.Head.SHA,
		},
	}

	if err := s.client.Create(ctx, preview); err != nil {
		if client.IgnoreAlreadyExists(err) == nil {
			logger.Info("PreviewEnvironment already exists", "name", preview.Name)
			return nil // Already exists, not an error
		}
		return fmt.Errorf("failed to create PreviewEnvironment: %w", err)
	}

	logger.Info("Created PreviewEnvironment", "name", preview.Name, "pr", event.Number)
	return nil
}

// handlePRClosed deletes a PreviewEnvironment CR when a PR is closed
func (s *Server) handlePRClosed(ctx context.Context, event *PullRequestEvent) error {
	logger := log.FromContext(ctx)

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("pr-%d", event.Number),
			Namespace: "previewd-system",
		},
	}

	if err := s.client.Delete(ctx, preview); err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("PreviewEnvironment not found (already deleted)", "name", preview.Name)
			return nil // Not found, not an error
		}
		return fmt.Errorf("failed to delete PreviewEnvironment: %w", err)
	}

	logger.Info("Deleted PreviewEnvironment", "name", preview.Name, "pr", event.Number)
	return nil
}

// handlePRSynchronized updates a PreviewEnvironment CR when new commits are pushed
func (s *Server) handlePRSynchronized(ctx context.Context, event *PullRequestEvent) error {
	logger := log.FromContext(ctx)

	preview := &previewv1alpha1.PreviewEnvironment{}
	name := fmt.Sprintf("pr-%d", event.Number)
	if err := s.client.Get(ctx, client.ObjectKey{Name: name, Namespace: "previewd-system"}, preview); err != nil {
		return fmt.Errorf("failed to get PreviewEnvironment: %w", err)
	}

	// Update HeadSHA
	preview.Spec.HeadSHA = event.PullRequest.Head.SHA

	if err := s.client.Update(ctx, preview); err != nil {
		return fmt.Errorf("failed to update PreviewEnvironment: %w", err)
	}

	logger.Info("Updated PreviewEnvironment", "name", preview.Name, "newSHA", event.PullRequest.Head.SHA)
	return nil
}

// sanitizeLabel converts a repository name to a valid Kubernetes label value
// Labels must be 63 characters or less and match [a-z0-9]([-a-z0-9]*[a-z0-9])?
func sanitizeLabel(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "_", "-")
	if len(s) > 63 {
		s = s[:63]
	}
	return s
}
