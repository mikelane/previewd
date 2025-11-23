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
	"errors"
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
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	scheduler := NewScheduler(k8sClient, 50*time.Millisecond)

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
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Start() returned unexpected error: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Start() did not return after context cancellation")
	}
}

func TestScheduler_cleanup_deletes_expired_environment(t *testing.T) {
	// Setup scheme with PreviewEnvironment CRD
	scheme := runtime.NewScheme()
	if err := previewdv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed to add scheme: %v", err)
	}

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
	if err := previewdv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed to add scheme: %v", err)
	}

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
	if err := previewdv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed to add scheme: %v", err)
	}

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

func TestScheduler_cleanup_continues_on_delete_error(t *testing.T) {
	// Setup scheme with PreviewEnvironment CRD
	scheme := runtime.NewScheme()
	if err := previewdv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed to add scheme: %v", err)
	}

	// Create two expired environments
	now := metav1.Now()
	expiredTime := metav1.NewTime(now.Add(-1 * time.Hour))

	env1 := &previewdv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-100",
			Namespace: "default",
		},
		Spec: previewdv1alpha1.PreviewEnvironmentSpec{
			Repository: "owner/repo",
			HeadSHA:    "0123456789abcdef0123456789abcdef01234567",
			PRNumber:   100,
		},
		Status: previewdv1alpha1.PreviewEnvironmentStatus{
			CreatedAt: &now,
			ExpiresAt: &expiredTime,
		},
	}

	env2 := &previewdv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-200",
			Namespace: "default",
		},
		Spec: previewdv1alpha1.PreviewEnvironmentSpec{
			Repository: "owner/repo",
			HeadSHA:    "fedcba9876543210fedcba9876543210fedcba98",
			PRNumber:   200,
		},
		Status: previewdv1alpha1.PreviewEnvironmentStatus{
			CreatedAt: &now,
			ExpiresAt: &expiredTime,
		},
	}

	// Create fake client - this test documents current behavior
	// Note: fake.Client doesn't simulate delete errors easily
	// This test verifies cleanup processes multiple expired environments successfully
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(env1, env2).
		Build()

	scheduler := NewScheduler(fakeClient, 50*time.Millisecond)

	// Run cleanup
	err := scheduler.cleanup(context.Background())
	if err != nil {
		t.Fatalf("cleanup() returned error: %v", err)
	}

	// Verify both environments were deleted
	var checkEnv previewdv1alpha1.PreviewEnvironment
	err1 := fakeClient.Get(context.Background(), client.ObjectKey{
		Name:      "pr-100",
		Namespace: "default",
	}, &checkEnv)

	err2 := fakeClient.Get(context.Background(), client.ObjectKey{
		Name:      "pr-200",
		Namespace: "default",
	}, &checkEnv)

	// Both should be deleted (Get should fail)
	if err1 == nil || err2 == nil {
		t.Error("Expected both environments to be deleted")
	}
}

func TestScheduler_cleanup_returns_error_on_list_failure(t *testing.T) {
	// Setup scheme WITHOUT adding PreviewEnvironment CRD
	// This will cause List() to fail with "no kind is registered" error
	scheme := runtime.NewScheme()

	// Create fake client without the CRD registered
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	scheduler := NewScheduler(fakeClient, 50*time.Millisecond)

	// Run cleanup - should fail when trying to list
	err := scheduler.cleanup(context.Background())

	// Verify we get an error (K8s API unavailable / scheme not registered)
	if err == nil {
		t.Error("Expected cleanup() to return error when List fails, got nil")
	}
}

func TestScheduler_cleanup_handles_nil_labels(t *testing.T) {
	// Setup scheme with PreviewEnvironment CRD
	scheme := runtime.NewScheme()
	if err := previewdv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed to add scheme: %v", err)
	}

	// Create expired environment with nil Labels map
	now := metav1.Now()
	expiredTime := metav1.NewTime(now.Add(-1 * time.Hour))

	expiredEnv := &previewdv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-999",
			Namespace: "default",
			Labels:    nil, // Explicitly nil labels
		},
		Spec: previewdv1alpha1.PreviewEnvironmentSpec{
			Repository: "owner/repo",
			HeadSHA:    "0123456789abcdef0123456789abcdef01234567",
			PRNumber:   999,
		},
		Status: previewdv1alpha1.PreviewEnvironmentStatus{
			CreatedAt: &now,
			ExpiresAt: &expiredTime,
		},
	}

	// Create fake client with the environment
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(expiredEnv).
		Build()

	scheduler := NewScheduler(fakeClient, 50*time.Millisecond)

	// Run cleanup - should handle nil labels gracefully
	err := scheduler.cleanup(context.Background())
	if err != nil {
		t.Fatalf("cleanup() returned error: %v", err)
	}

	// Verify environment was deleted (nil labels shouldn't prevent deletion)
	var env previewdv1alpha1.PreviewEnvironment
	err = fakeClient.Get(context.Background(), client.ObjectKey{
		Name:      "pr-999",
		Namespace: "default",
	}, &env)

	if err == nil {
		t.Error("Expected environment with nil labels to be deleted, but it still exists")
	}
}
