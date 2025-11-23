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
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	previewv1alpha1 "github.com/mikelane/previewd/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

//nolint:gosec // Test secret, not a real credential
const testSecret = "test-webhook-secret"

func setupTest(t *testing.T) (*Server, client.Client) {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := previewv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed to add scheme: %v", err)
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	server := NewServer("localhost", 8080, fakeClient, testSecret)
	return server, fakeClient
}

//nolint:unparam // secret is always testSecret in tests, but keeping parameter for clarity
func computeSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestHandleHealth(t *testing.T) {
	server, _ := setupTest(t)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleHealth returns %d, expected %d", w.Code, http.StatusOK)
	}

	if w.Body.String() != "OK" {
		t.Errorf("handleHealth body is %q, expected %q", w.Body.String(), "OK")
	}
}

func TestHandleWebhook_MethodNotAllowed(t *testing.T) {
	server, _ := setupTest(t)

	req := httptest.NewRequest("GET", "/webhook", nil)
	w := httptest.NewRecorder()

	server.handleWebhook(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("handleWebhook with GET returns %d, expected %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleWebhook_InvalidSignature(t *testing.T) {
	server, _ := setupTest(t)

	payload := []byte(`{"action":"opened","number":123}`)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", "sha256=invalid")
	w := httptest.NewRecorder()

	server.handleWebhook(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("handleWebhook with invalid signature returns %d, expected %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleWebhook_NonPREvent(t *testing.T) {
	server, _ := setupTest(t)

	payload := []byte(`{"action":"push"}`)
	signature := computeSignature(payload, testSecret)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", signature)
	w := httptest.NewRecorder()

	server.handleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleWebhook with push event returns %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestHandleWebhook_InvalidJSON(t *testing.T) {
	server, _ := setupTest(t)

	payload := []byte(`{invalid json}`)
	signature := computeSignature(payload, testSecret)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", signature)
	w := httptest.NewRecorder()

	server.handleWebhook(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleWebhook with invalid JSON returns %d, expected %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePROpened(t *testing.T) {
	server, k8sClient := setupTest(t)

	event := PullRequestEvent{
		Action: "opened",
		Number: 123,
		PullRequest: PullRequest{
			Head: Ref{
				Ref: "feature/test",
				SHA: "abc123",
			},
			Base: Ref{
				Ref: "main",
				SHA: "def456",
			},
		},
		Repository: Repository{
			FullName: "company/repo",
		},
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal test event: %v", err)
	}
	signature := computeSignature(payload, testSecret)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", signature)
	w := httptest.NewRecorder()

	server.handleWebhook(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("handleWebhook for PR opened returns %d, expected %d", w.Code, http.StatusCreated)
	}

	// Verify PreviewEnvironment was created
	preview := &previewv1alpha1.PreviewEnvironment{}
	getErr := k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "pr-123",
		Namespace: "previewd-system",
	}, preview)

	if getErr != nil {
		t.Fatalf("Failed to get PreviewEnvironment: %v", getErr)
	}

	if preview.Spec.PRNumber != 123 {
		t.Errorf("PreviewEnvironment PRNumber is %d, expected 123", preview.Spec.PRNumber)
	}

	if preview.Spec.Repository != "company/repo" {
		t.Errorf("PreviewEnvironment Repository is %s, expected company/repo", preview.Spec.Repository)
	}

	if preview.Spec.HeadSHA != "abc123" {
		t.Errorf("PreviewEnvironment HeadSHA is %s, expected abc123", preview.Spec.HeadSHA)
	}
}

func TestHandlePROpened_AlreadyExists(t *testing.T) {
	server, k8sClient := setupTest(t)

	// Create existing PreviewEnvironment first
	existingPreview := &previewv1alpha1.PreviewEnvironment{}
	//nolint:goconst // Test data, not a meaningful constant
	existingPreview.Name = "pr-123"
	//nolint:goconst // Test data, not a meaningful constant
	existingPreview.Namespace = "previewd-system"
	existingPreview.Spec.PRNumber = 123
	existingPreview.Spec.HeadSHA = "oldsha"
	if err := k8sClient.Create(context.Background(), existingPreview); err != nil {
		t.Fatalf("Failed to create existing PreviewEnvironment: %v", err)
	}

	event := PullRequestEvent{
		Action: "opened",
		Number: 123,
		PullRequest: PullRequest{
			Head: Ref{
				Ref: "feature/test",
				SHA: "abc123",
			},
			Base: Ref{
				Ref: "main",
				SHA: "def456",
			},
		},
		Repository: Repository{
			FullName: "company/repo",
		},
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal test event: %v", err)
	}
	signature := computeSignature(payload, testSecret)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", signature)
	w := httptest.NewRecorder()

	server.handleWebhook(w, req)

	// Should return Created even though PreviewEnvironment already exists
	if w.Code != http.StatusCreated {
		t.Errorf("handleWebhook for PR opened (already exists) returns %d, expected %d", w.Code, http.StatusCreated)
	}
}

func TestHandlePRClosed(t *testing.T) {
	server, k8sClient := setupTest(t)

	// Create existing PreviewEnvironment
	preview := &previewv1alpha1.PreviewEnvironment{}
	preview.Name = "pr-123"
	preview.Namespace = "previewd-system"
	preview.Spec.PRNumber = 123
	if err := k8sClient.Create(context.Background(), preview); err != nil {
		t.Fatalf("Failed to create test PreviewEnvironment: %v", err)
	}

	event := PullRequestEvent{
		Action: "closed",
		Number: 123,
		Repository: Repository{
			FullName: "company/repo",
		},
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal test event: %v", err)
	}
	signature := computeSignature(payload, testSecret)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", signature)
	w := httptest.NewRecorder()

	server.handleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleWebhook for PR closed returns %d, expected %d", w.Code, http.StatusOK)
	}

	// Verify PreviewEnvironment was deleted
	getErr := k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "pr-123",
		Namespace: "previewd-system",
	}, preview)

	if getErr == nil {
		t.Error("PreviewEnvironment still exists after PR closed")
	}
}

func TestHandlePRClosed_NotFound(t *testing.T) {
	server, _ := setupTest(t)

	// Don't create a PreviewEnvironment - test deletion of non-existent resource

	event := PullRequestEvent{
		Action: "closed",
		Number: 999, // PR that never had a PreviewEnvironment
		Repository: Repository{
			FullName: "company/repo",
		},
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal test event: %v", err)
	}
	signature := computeSignature(payload, testSecret)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", signature)
	w := httptest.NewRecorder()

	server.handleWebhook(w, req)

	// Should return OK even though PreviewEnvironment doesn't exist
	if w.Code != http.StatusOK {
		t.Errorf("handleWebhook for PR closed (not found) returns %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestHandlePRSynchronized(t *testing.T) {
	server, k8sClient := setupTest(t)

	// Create existing PreviewEnvironment
	preview := &previewv1alpha1.PreviewEnvironment{}
	preview.Name = "pr-123"
	preview.Namespace = "previewd-system"
	preview.Spec.PRNumber = 123
	preview.Spec.HeadSHA = "oldsha"
	if err := k8sClient.Create(context.Background(), preview); err != nil {
		t.Fatalf("Failed to create test PreviewEnvironment: %v", err)
	}

	event := PullRequestEvent{
		Action: "synchronize",
		Number: 123,
		PullRequest: PullRequest{
			Head: Ref{
				SHA: "newsha123",
			},
		},
		Repository: Repository{
			FullName: "company/repo",
		},
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal test event: %v", err)
	}
	signature := computeSignature(payload, testSecret)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", signature)
	w := httptest.NewRecorder()

	server.handleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleWebhook for PR synchronized returns %d, expected %d", w.Code, http.StatusOK)
	}

	// Verify PreviewEnvironment was updated
	updated := &previewv1alpha1.PreviewEnvironment{}
	getErr := k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "pr-123",
		Namespace: "previewd-system",
	}, updated)

	if getErr != nil {
		t.Fatalf("Failed to get updated PreviewEnvironment: %v", getErr)
	}

	if updated.Spec.HeadSHA != "newsha123" {
		t.Errorf("PreviewEnvironment HeadSHA is %s, expected newsha123", updated.Spec.HeadSHA)
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(3, 100*time.Millisecond)

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		if !rl.Allow("test-repo") {
			t.Errorf("Request %d was rate limited, expected to be allowed", i+1)
		}
	}

	// 4th request should be rate limited
	if rl.Allow("test-repo") {
		t.Error("Request 4 was allowed, expected to be rate limited")
	}

	// Wait for window to reset
	time.Sleep(110 * time.Millisecond)

	// Should allow again after reset
	if !rl.Allow("test-repo") {
		t.Error("Request after reset was rate limited, expected to be allowed")
	}
}

// TestRateLimiter_Concurrency verifies thread-safe concurrent access
func TestRateLimiter_Concurrency(t *testing.T) {
	rl := NewRateLimiter(100, time.Second)
	repo := "test/repo"

	// Run 50 goroutines making requests concurrently
	done := make(chan bool)
	for i := 0; i < 50; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				_ = rl.Allow(repo)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to finish
	for i := 0; i < 50; i++ {
		<-done
	}

	// Verify rate limiter still works correctly
	// After 500 requests (50*10), we should be at limit (100 per second)
	if rl.Allow(repo) {
		t.Error("Rate limiter allowed request after exceeding limit")
	}
}

func TestRateLimiter_DifferentRepos(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)

	// Repo A: 2 requests (at limit)
	if !rl.Allow("repo-a") {
		t.Error("repo-a request 1 was rate limited")
	}
	if !rl.Allow("repo-a") {
		t.Error("repo-a request 2 was rate limited")
	}

	// Repo B: should still be allowed (different bucket)
	if !rl.Allow("repo-b") {
		t.Error("repo-b request 1 was rate limited")
	}

	// Repo A: should be rate limited
	if rl.Allow("repo-a") {
		t.Error("repo-a request 3 was allowed, expected rate limit")
	}
}

func TestHandleWebhook_RateLimited(t *testing.T) {
	server, _ := setupTest(t)

	// Exhaust rate limit (10 requests)
	event := PullRequestEvent{
		Action: "opened",
		Number: 999,
		Repository: Repository{
			FullName: "test/repo",
		},
		PullRequest: PullRequest{
			Head: Ref{SHA: "test"},
			Base: Ref{Ref: "main"},
		},
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal test event: %v", err)
	}
	signature := computeSignature(payload, testSecret)

	// Send 11 requests
	for i := 0; i < 11; i++ {
		req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
		req.Header.Set("X-GitHub-Event", "pull_request")
		req.Header.Set("X-Hub-Signature-256", signature)
		w := httptest.NewRecorder()

		server.handleWebhook(w, req)

		if i < 10 {
			// First 10 should succeed (or 201 Created)
			if w.Code != http.StatusCreated {
				t.Errorf("Request %d returned %d, expected %d", i+1, w.Code, http.StatusCreated)
			}
		} else {
			// 11th should be rate limited
			if w.Code != http.StatusTooManyRequests {
				t.Errorf("Request %d returned %d, expected %d (rate limited)", i+1, w.Code, http.StatusTooManyRequests)
			}
		}
	}
}

// TestHandleWebhook_BodyReadFailure verifies that body read errors are handled correctly
func TestHandleWebhook_BodyReadFailure(t *testing.T) {
	server, _ := setupTest(t)

	// Create a request with a body that will error on read
	req := httptest.NewRequest("POST", "/webhook", &errorReader{})
	req.Header.Set("X-GitHub-Event", "pull_request")
	w := httptest.NewRecorder()

	server.handleWebhook(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleWebhook with read error returns %d, expected %d", w.Code, http.StatusBadRequest)
	}
}

// errorReader always returns an error when Read is called
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestSanitizeLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"company/repo", "company-repo"},
		{"Company/Repo", "company-repo"},
		{"company_name/repo_name", "company-name-repo-name"},
		{"a" + string(make([]byte, 70)), "a" + string(make([]byte, 62))}, // Truncate to 63
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeLabel(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeLabel(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
			if len(result) > 63 {
				t.Errorf("sanitizeLabel(%q) = %q (len=%d), exceeds 63 characters", tt.input, result, len(result))
			}
		})
	}
}
