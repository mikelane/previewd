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

//go:build integration
// +build integration

package namespace

import (
	"context"
	"testing"

	previewv1alpha1 "github.com/mikelane/previewd/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestIntegration_FullWorkflow tests the complete namespace management workflow
func TestIntegration_FullWorkflow(t *testing.T) {
	// Setup
	scheme := runtime.NewScheme()
	_ = previewv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	manager := NewManager(c, scheme)
	ctx := context.Background()

	// Create a PreviewEnvironment
	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-42",
			Namespace: "previewd-system",
			UID:       "test-integration-uid",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   42,
			Repository: "myorg/myapp",
			HeadSHA:    "abcdef1234567890abcdef1234567890abcdef12",
			TTL:        "4h",
		},
	}

	// Step 1: Ensure namespace is created
	err := manager.EnsureNamespace(ctx, preview)
	if err != nil {
		t.Fatalf("failed to ensure namespace: %v", err)
	}

	// Verify namespace exists
	nsName := manager.GetNamespaceName(preview)
	ns := &corev1.Namespace{}
	err = c.Get(ctx, types.NamespacedName{Name: nsName}, ns)
	if err != nil {
		t.Errorf("namespace should exist: %v", err)
	}

	// Verify namespace labels
	if ns.Labels["preview.previewd.io/pr"] != "42" {
		t.Errorf("PR label should be '42', got %s", ns.Labels["preview.previewd.io/pr"])
	}

	// Step 2: Ensure resource quota is created
	err = manager.EnsureResourceQuota(ctx, preview, nsName)
	if err != nil {
		t.Fatalf("failed to ensure resource quota: %v", err)
	}

	// Verify quota exists
	quota := &corev1.ResourceQuota{}
	err = c.Get(ctx, types.NamespacedName{Name: "preview-quota", Namespace: nsName}, quota)
	if err != nil {
		t.Errorf("resource quota should exist: %v", err)
	}

	// Step 3: Ensure network policies are created
	err = manager.EnsureNetworkPolicies(ctx, preview, nsName)
	if err != nil {
		t.Fatalf("failed to ensure network policies: %v", err)
	}

	// Verify policies exist
	policies := []string{"default-deny-all", "allow-ingress", "allow-egress"}
	for _, policyName := range policies {
		policy := &networkingv1.NetworkPolicy{}
		err = c.Get(ctx, types.NamespacedName{Name: policyName, Namespace: nsName}, policy)
		if err != nil {
			t.Errorf("network policy %s should exist: %v", policyName, err)
		}
	}

	// Step 4: Update preview (simulate reconciliation)
	err = manager.EnsureNamespace(ctx, preview)
	if err != nil {
		t.Fatalf("idempotent namespace ensure should not fail: %v", err)
	}

	err = manager.EnsureResourceQuota(ctx, preview, nsName)
	if err != nil {
		t.Fatalf("idempotent quota ensure should not fail: %v", err)
	}

	err = manager.EnsureNetworkPolicies(ctx, preview, nsName)
	if err != nil {
		t.Fatalf("idempotent policies ensure should not fail: %v", err)
	}

	// Step 5: Cleanup when preview is deleted
	preview.DeletionTimestamp = &metav1.Time{}
	err = manager.Cleanup(ctx, preview)
	if err != nil {
		t.Fatalf("cleanup should not fail: %v", err)
	}

	// In a real cluster, resources would be garbage collected
	// Here we just verify the delete was called
	t.Log("Integration test completed successfully")
}
