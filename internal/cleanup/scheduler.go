/*
Copyright (c) 2025 Mike Lane

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package cleanup

import (
	"context"
	"time"

	previewdv1alpha1 "github.com/mikelane/previewd/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Scheduler manages automatic cleanup of expired PreviewEnvironment resources.
// It runs periodically to check for environments that have exceeded their TTL
// and deletes them to prevent resource waste.
type Scheduler struct {
	client   client.Client
	interval time.Duration
}

// NewScheduler creates a new cleanup scheduler with the specified interval.
// The scheduler will check for expired environments every interval duration.
//
// Parameters:
//   - k8sClient: Kubernetes client for listing and deleting PreviewEnvironments
//   - interval: Duration between cleanup runs (e.g., 5*time.Minute)
//
// Returns a configured Scheduler ready to start.
func NewScheduler(k8sClient client.Client, interval time.Duration) *Scheduler {
	return &Scheduler{
		client:   k8sClient,
		interval: interval,
	}
}

// Start begins the cleanup scheduler, running periodically until the context is canceled.
// It uses a ticker to trigger cleanup at the configured interval and respects graceful
// shutdown via context cancellation.
//
// Parameters:
//   - ctx: Context for cancellation and deadline control
//
// Returns nil on graceful shutdown, or an error if cleanup operations fail.
func (s *Scheduler) Start(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	logger := log.FromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.cleanup(ctx); err != nil {
				logger.Error(err, "cleanup pass failed")
				// Continue to next tick - don't stop scheduler on transient errors
			}
		}
	}
}

// cleanup performs a single cleanup pass, deleting all expired PreviewEnvironments.
// It lists all PreviewEnvironment resources, checks their ExpiresAt timestamps,
// and deletes any that have exceeded their TTL.
//
// The following rules apply:
//   - Environments without ExpiresAt are skipped
//   - Environments with label "preview.previewd.io/do-not-expire=true" are skipped
//   - Only environments where ExpiresAt is before current time are deleted
//
// Returns an error if listing or deletion operations fail.
func (s *Scheduler) cleanup(ctx context.Context) error {
	// List all PreviewEnvironments
	var envList previewdv1alpha1.PreviewEnvironmentList
	if err := s.client.List(ctx, &envList); err != nil {
		return err
	}

	// Check each environment for expiration
	now := metav1.Now()
	for i := range envList.Items {
		env := &envList.Items[i]

		// Skip if no expiration time set
		if env.Status.ExpiresAt == nil {
			continue
		}

		// Skip if environment has do-not-expire label
		if env.Labels != nil {
			if doNotExpire, exists := env.Labels["preview.previewd.io/do-not-expire"]; exists && doNotExpire == "true" {
				continue
			}
		}

		// Check if expired
		if env.Status.ExpiresAt.Before(&now) {
			// Delete expired environment
			if err := s.client.Delete(ctx, env); err != nil {
				return err
			}
		}
	}

	return nil
}
