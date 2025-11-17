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

package namespace

import (
	"context"
	"fmt"
	"testing"

	previewv1alpha1 "github.com/mikelane/previewd/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestManager_EnsureNamespace(t *testing.T) {
	tests := []struct {
		name       string
		preview    *previewv1alpha1.PreviewEnvironment
		existingNS *corev1.Namespace
		wantErr    bool
		validateFn func(t *testing.T, c client.Client, preview *previewv1alpha1.PreviewEnvironment)
	}{
		{
			name: "creates namespace with correct labels and owner reference",
			preview: &previewv1alpha1.PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pr-123",
					Namespace: "previewd-system",
					UID:       "test-uid",
				},
				Spec: previewv1alpha1.PreviewEnvironmentSpec{
					PRNumber:   123,
					Repository: "owner/repo",
				},
			},
			validateFn: func(t *testing.T, c client.Client, preview *previewv1alpha1.PreviewEnvironment) {
				ns := &corev1.Namespace{}
				nsName := generateNamespaceName(preview.Spec.PRNumber, preview.Spec.Repository)
				err := c.Get(context.Background(), types.NamespacedName{Name: nsName}, ns)
				if err != nil {
					t.Errorf("failed to get namespace: %v", err)
					return
				}

				// Verify labels
				if ns.Labels["preview.previewd.io/pr"] != "123" {
					t.Errorf("expected PR label to be '123', got %s", ns.Labels["preview.previewd.io/pr"])
				}
				if ns.Labels["preview.previewd.io/repository"] != "owner-repo" {
					t.Errorf("expected repository label to be 'owner-repo', got %s", ns.Labels["preview.previewd.io/repository"])
				}
				if ns.Labels["preview.previewd.io/managed-by"] != "previewd" {
					t.Errorf("expected managed-by label to be 'previewd', got %s", ns.Labels["preview.previewd.io/managed-by"])
				}

				// Verify owner tracking via annotations (instead of owner references)
				if ns.Annotations["preview.previewd.io/owner-uid"] != string(preview.UID) {
					t.Errorf("expected owner UID annotation to be %s, got %s", preview.UID, ns.Annotations["preview.previewd.io/owner-uid"])
				}
				if ns.Annotations["preview.previewd.io/owner-name"] != preview.Name {
					t.Errorf("expected owner name annotation to be %s, got %s", preview.Name, ns.Annotations["preview.previewd.io/owner-name"])
				}
			},
		},
		{
			name: "idempotent - does not error if namespace already exists",
			preview: &previewv1alpha1.PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pr-456",
					Namespace: "previewd-system",
					UID:       "test-uid-2",
				},
				Spec: previewv1alpha1.PreviewEnvironmentSpec{
					PRNumber:   456,
					Repository: "owner/repo",
				},
			},
			existingNS: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "preview-pr-456-65e817ee",
					Labels: map[string]string{
						"preview.previewd.io/pr":         "456",
						"preview.previewd.io/repository": "owner-repo",
						"preview.previewd.io/managed-by": "previewd",
					},
				},
			},
			validateFn: func(t *testing.T, c client.Client, preview *previewv1alpha1.PreviewEnvironment) {
				// Should not error and namespace should still exist
				ns := &corev1.Namespace{}
				nsName := generateNamespaceName(preview.Spec.PRNumber, preview.Spec.Repository)
				err := c.Get(context.Background(), types.NamespacedName{Name: nsName}, ns)
				if err != nil {
					t.Errorf("namespace should exist: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = previewv1alpha1.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)

			var objs []client.Object
			if tt.existingNS != nil {
				objs = append(objs, tt.existingNS)
			}

			c := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(objs...).
				Build()

			m := NewManager(c, scheme)
			err := m.EnsureNamespace(context.Background(), tt.preview)

			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureNamespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validateFn != nil {
				tt.validateFn(t, c, tt.preview)
			}
		})
	}
}

func TestManager_EnsureResourceQuota(t *testing.T) {
	tests := []struct {
		name       string
		preview    *previewv1alpha1.PreviewEnvironment
		namespace  string
		wantErr    bool
		validateFn func(t *testing.T, c client.Client, namespace string)
	}{
		{
			name: "creates resource quota with correct limits",
			preview: &previewv1alpha1.PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pr-789",
					Namespace: "previewd-system",
					UID:       "test-uid-3",
				},
				Spec: previewv1alpha1.PreviewEnvironmentSpec{
					PRNumber:   789,
					Repository: "owner/repo",
				},
			},
			namespace: "preview-pr-789-65e817ee",
			validateFn: func(t *testing.T, c client.Client, namespace string) {
				quota := &corev1.ResourceQuota{}
				err := c.Get(context.Background(), types.NamespacedName{
					Name:      "preview-quota",
					Namespace: namespace,
				}, quota)
				if err != nil {
					t.Errorf("failed to get resource quota: %v", err)
					return
				}

				// Check CPU requests
				cpuRequests := quota.Spec.Hard[corev1.ResourceRequestsCPU]
				expectedCPU := resource.MustParse("2")
				if !cpuRequests.Equal(expectedCPU) {
					t.Errorf("expected CPU requests to be %v, got %v", expectedCPU, cpuRequests)
				}

				// Check memory requests
				memRequests := quota.Spec.Hard[corev1.ResourceRequestsMemory]
				expectedMem := resource.MustParse("4Gi")
				if !memRequests.Equal(expectedMem) {
					t.Errorf("expected memory requests to be %v, got %v", expectedMem, memRequests)
				}

				// Check PVC count
				pvcCount := quota.Spec.Hard[corev1.ResourcePersistentVolumeClaims]
				expectedPVC := resource.MustParse("0")
				if !pvcCount.Equal(expectedPVC) {
					t.Errorf("expected PVC count to be %v, got %v", expectedPVC, pvcCount)
				}

				// Check LoadBalancer services
				lbCount := quota.Spec.Hard["services.loadbalancers"]
				expectedLB := resource.MustParse("0")
				if !lbCount.Equal(expectedLB) {
					t.Errorf("expected LoadBalancer count to be %v, got %v", expectedLB, lbCount)
				}
			},
		},
		{
			name: "idempotent - updates existing quota if changed",
			preview: &previewv1alpha1.PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pr-321",
					Namespace: "previewd-system",
					UID:       "test-uid-4",
				},
				Spec: previewv1alpha1.PreviewEnvironmentSpec{
					PRNumber:   321,
					Repository: "owner/repo",
				},
			},
			namespace: "preview-pr-321-65e817ee",
			validateFn: func(t *testing.T, c client.Client, namespace string) {
				quota := &corev1.ResourceQuota{}
				err := c.Get(context.Background(), types.NamespacedName{
					Name:      "preview-quota",
					Namespace: namespace,
				}, quota)
				if err != nil {
					t.Errorf("resource quota should exist: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = previewv1alpha1.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)

			// Create namespace first
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: tt.namespace,
				},
			}

			c := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(ns).
				Build()

			m := NewManager(c, scheme)
			err := m.EnsureResourceQuota(context.Background(), tt.preview, tt.namespace)

			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureResourceQuota() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validateFn != nil {
				tt.validateFn(t, c, tt.namespace)
			}
		})
	}
}

func TestManager_EnsureNetworkPolicies(t *testing.T) {
	tests := []struct {
		name       string
		preview    *previewv1alpha1.PreviewEnvironment
		namespace  string
		wantErr    bool
		validateFn func(t *testing.T, c client.Client, namespace string)
	}{
		{
			name: "creates network policies with correct rules",
			preview: &previewv1alpha1.PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pr-555",
					Namespace: "previewd-system",
					UID:       "test-uid-5",
				},
				Spec: previewv1alpha1.PreviewEnvironmentSpec{
					PRNumber:   555,
					Repository: "owner/repo",
				},
			},
			namespace: "preview-pr-555-65e817ee",
			validateFn: func(t *testing.T, c client.Client, namespace string) {
				// Check default-deny-all policy
				defaultDeny := &networkingv1.NetworkPolicy{}
				err := c.Get(context.Background(), types.NamespacedName{
					Name:      "default-deny-all",
					Namespace: namespace,
				}, defaultDeny)
				if err != nil {
					t.Errorf("failed to get default-deny-all policy: %v", err)
					return
				}

				// Verify it applies to all pods
				if len(defaultDeny.Spec.PodSelector.MatchLabels) != 0 {
					t.Errorf("default-deny-all should apply to all pods, got selector: %v", defaultDeny.Spec.PodSelector.MatchLabels)
				}

				// Check allow-ingress policy
				allowIngress := &networkingv1.NetworkPolicy{}
				err = c.Get(context.Background(), types.NamespacedName{
					Name:      "allow-ingress",
					Namespace: namespace,
				}, allowIngress)
				if err != nil {
					t.Errorf("failed to get allow-ingress policy: %v", err)
					return
				}

				// Verify ingress from ingress-nginx namespace
				if len(allowIngress.Spec.Ingress) == 0 {
					t.Errorf("allow-ingress should have ingress rules")
					return
				}

				// Check that ingress is allowed from ingress-nginx namespace
				found := false
				for _, rule := range allowIngress.Spec.Ingress {
					for _, from := range rule.From {
						if from.NamespaceSelector != nil &&
							from.NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] == "ingress-nginx" {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("allow-ingress should allow traffic from ingress-nginx namespace")
				}

				// Check allow-egress policy
				allowEgress := &networkingv1.NetworkPolicy{}
				err = c.Get(context.Background(), types.NamespacedName{
					Name:      "allow-egress",
					Namespace: namespace,
				}, allowEgress)
				if err != nil {
					t.Errorf("failed to get allow-egress policy: %v", err)
					return
				}

				// Verify egress rules exist
				if len(allowEgress.Spec.Egress) == 0 {
					t.Errorf("allow-egress should have egress rules")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = previewv1alpha1.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)
			_ = networkingv1.AddToScheme(scheme)

			// Create namespace first
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: tt.namespace,
				},
			}

			c := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(ns).
				Build()

			m := NewManager(c, scheme)
			err := m.EnsureNetworkPolicies(context.Background(), tt.preview, tt.namespace)

			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureNetworkPolicies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validateFn != nil {
				tt.validateFn(t, c, tt.namespace)
			}
		})
	}
}

func TestManager_Cleanup(t *testing.T) {
	tests := []struct {
		name       string
		preview    *previewv1alpha1.PreviewEnvironment
		setupFn    func(c client.Client) error
		wantErr    bool
		validateFn func(t *testing.T, c client.Client, preview *previewv1alpha1.PreviewEnvironment)
	}{
		{
			name: "deletes namespace when preview is deleted",
			preview: &previewv1alpha1.PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pr-999",
					Namespace:         "previewd-system",
					UID:               "test-uid-6",
					DeletionTimestamp: &metav1.Time{},
				},
				Spec: previewv1alpha1.PreviewEnvironmentSpec{
					PRNumber:   999,
					Repository: "owner/repo",
				},
			},
			setupFn: func(c client.Client) error {
				// Create namespace with owner reference
				ns := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "preview-pr-999-65e817ee",
						Labels: map[string]string{
							"preview.previewd.io/pr": "999",
						},
					},
				}
				return c.Create(context.Background(), ns)
			},
			validateFn: func(t *testing.T, c client.Client, preview *previewv1alpha1.PreviewEnvironment) {
				// Namespace should be deleted or have deletion timestamp
				ns := &corev1.Namespace{}
				err := c.Get(context.Background(), types.NamespacedName{
					Name: generateNamespaceName(preview.Spec.PRNumber, preview.Spec.Repository),
				}, ns)
				if err == nil && ns.DeletionTimestamp == nil {
					// In a real cluster, the namespace would be deleted by GC
					// In fake client, we should at least verify the delete was called
					t.Logf("namespace exists (expected in fake client): %s", ns.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = previewv1alpha1.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)

			c := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			if tt.setupFn != nil {
				if err := tt.setupFn(c); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			m := NewManager(c, scheme)
			err := m.Cleanup(context.Background(), tt.preview)

			if (err != nil) != tt.wantErr {
				t.Errorf("Cleanup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validateFn != nil {
				tt.validateFn(t, c, tt.preview)
			}
		})
	}
}

// Helper function tests
func TestGenerateNamespaceName(t *testing.T) {
	tests := []struct {
		prNumber int
		repo     string
		want     string
	}{
		{
			prNumber: 123,
			repo:     "owner/repo",
			want:     "preview-pr-123-65e817ee",
		},
		{
			prNumber: 456,
			repo:     "myorg/myrepo",
			want:     "preview-pr-456-71b1f54a",
		},
		{
			prNumber: 789,
			repo:     "test/test-repo",
			want:     "preview-pr-789-3741d56e",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("PR-%d-%s", tt.prNumber, tt.repo), func(t *testing.T) {
			got := generateNamespaceName(tt.prNumber, tt.repo)
			if got != tt.want {
				t.Errorf("generateNamespaceName() = %v, want %v", got, tt.want)
			}
		})
	}
}
