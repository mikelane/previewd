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
	"testing"
	"time"

	previewdv1alpha1 "github.com/mikelane/previewd/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestScheduler_Start_runs_periodically_and_stops_gracefully(t *testing.T) {
	// Setup
	scheme := runtime.NewScheme()
	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	scheduler := NewScheduler(client, 50*time.Millisecond)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start scheduler in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- scheduler.Start(ctx)
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Verify graceful shutdown (no error)
	select {
	case err := <-errCh:
		if err != nil && err != context.DeadlineExceeded {
			t.Errorf("Start() returned unexpected error: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Start() did not return after context cancellation")
	}
}

func TestScheduler_cleanup_deletes_expired_environment(t *testing.T) {
	// Setup scheme with PreviewEnvironment CRD
	scheme := runtime.NewScheme()
	_ = previewdv1alpha1.AddToScheme(scheme)

	// Create expired environment (expires 1 hour ago)
	now := metav1.Now()
	expiredTime := metav1.NewTime(now.Add(-1 * time.Hour))

	expiredEnv := &previewdv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "default",
		},
		Spec: previewdv1alpha1.PreviewEnvironmentSpec{
			Repository: "owner/repo",
			HeadSHA:    "0123456789abcdef0123456789abcdef01234567",
			PRNumber:   123,
		},
		Status: previewdv1alpha1.PreviewEnvironmentStatus{
			CreatedAt: &now,
			ExpiresAt: &expiredTime,
		},
	}

	// Create fake client with the expired environment
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(expiredEnv).
		Build()

	scheduler := NewScheduler(fakeClient, 50*time.Millisecond)

	// Run cleanup
	err := scheduler.cleanup(context.Background())
	if err != nil {
		t.Fatalf("cleanup() returned error: %v", err)
	}

	// Verify environment was deleted
	var env previewdv1alpha1.PreviewEnvironment
	err = fakeClient.Get(context.Background(), client.ObjectKey{
		Name:      "pr-123",
		Namespace: "default",
	}, &env)

	if err == nil {
		t.Error("Expected environment to be deleted, but it still exists")
	}
}

func TestScheduler_cleanup_does_not_delete_non_expired_environment(t *testing.T) {
	// Setup scheme with PreviewEnvironment CRD
	scheme := runtime.NewScheme()
	_ = previewdv1alpha1.AddToScheme(scheme)

	// Create non-expired environment (expires 2 hours from now)
	now := metav1.Now()
	futureTime := metav1.NewTime(now.Add(2 * time.Hour))

	nonExpiredEnv := &previewdv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-456",
			Namespace: "default",
		},
		Spec: previewdv1alpha1.PreviewEnvironmentSpec{
			Repository: "owner/repo",
			HeadSHA:    "0123456789abcdef0123456789abcdef01234567",
			PRNumber:   456,
		},
		Status: previewdv1alpha1.PreviewEnvironmentStatus{
			CreatedAt: &now,
			ExpiresAt: &futureTime,
		},
	}

	// Create fake client with the non-expired environment
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(nonExpiredEnv).
		Build()

	scheduler := NewScheduler(fakeClient, 50*time.Millisecond)

	// Run cleanup
	err := scheduler.cleanup(context.Background())
	if err != nil {
		t.Fatalf("cleanup() returned error: %v", err)
	}

	// Verify environment still exists
	var env previewdv1alpha1.PreviewEnvironment
	err = fakeClient.Get(context.Background(), client.ObjectKey{
		Name:      "pr-456",
		Namespace: "default",
	}, &env)

	if err != nil {
		t.Errorf("Expected environment to still exist, but got error: %v", err)
	}
}

func TestScheduler_cleanup_skips_environment_with_do_not_expire_label(t *testing.T) {
	// Setup scheme with PreviewEnvironment CRD
	scheme := runtime.NewScheme()
	_ = previewdv1alpha1.AddToScheme(scheme)

	// Create expired environment with do-not-expire label
	now := metav1.Now()
	expiredTime := metav1.NewTime(now.Add(-1 * time.Hour))

	expiredEnvWithLabel := &previewdv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-789",
			Namespace: "default",
			Labels: map[string]string{
				"preview.previewd.io/do-not-expire": "true",
			},
		},
		Spec: previewdv1alpha1.PreviewEnvironmentSpec{
			Repository: "owner/repo",
			HeadSHA:    "0123456789abcdef0123456789abcdef01234567",
			PRNumber:   789,
		},
		Status: previewdv1alpha1.PreviewEnvironmentStatus{
			CreatedAt: &now,
			ExpiresAt: &expiredTime,
		},
	}

	// Create fake client with the labeled environment
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(expiredEnvWithLabel).
		Build()

	scheduler := NewScheduler(fakeClient, 50*time.Millisecond)

	// Run cleanup
	err := scheduler.cleanup(context.Background())
	if err != nil {
		t.Fatalf("cleanup() returned error: %v", err)
	}

	// Verify environment still exists (not deleted due to label)
	var env previewdv1alpha1.PreviewEnvironment
	err = fakeClient.Get(context.Background(), client.ObjectKey{
		Name:      "pr-789",
		Namespace: "default",
	}, &env)

	if err != nil {
		t.Errorf("Expected environment with do-not-expire label to still exist, but got error: %v", err)
	}
}
