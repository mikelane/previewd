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

// Package namespace provides functionality for managing Kubernetes namespaces
// for preview environments, including creation, configuration, and cleanup.
package namespace

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	previewv1alpha1 "github.com/mikelane/previewd/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const managedByLabel = "previewd"

// Manager handles namespace lifecycle for preview environments
type Manager struct {
	client client.Client
	scheme *runtime.Scheme
}

// NewManager creates a new namespace manager
func NewManager(c client.Client, scheme *runtime.Scheme) *Manager {
	return &Manager{
		client: c,
		scheme: scheme,
	}
}

// EnsureNamespace creates or updates a namespace for the preview environment
// with appropriate labels. Note: We don't set owner references on namespaces
// as cross-namespace owner references are not allowed in Kubernetes.
func (m *Manager) EnsureNamespace(ctx context.Context, preview *previewv1alpha1.PreviewEnvironment) error {
	nsName := generateNamespaceName(preview.Spec.PRNumber, preview.Spec.Repository)

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, m.client, ns, func() error {
		// Set labels
		if ns.Labels == nil {
			ns.Labels = make(map[string]string)
		}
		ns.Labels["preview.previewd.io/pr"] = fmt.Sprintf("%d", preview.Spec.PRNumber)
		ns.Labels["preview.previewd.io/repository"] = strings.ReplaceAll(preview.Spec.Repository, "/", "-")
		ns.Labels["preview.previewd.io/managed-by"] = managedByLabel

		// Add annotations to track the owner (informational only)
		if ns.Annotations == nil {
			ns.Annotations = make(map[string]string)
		}
		ns.Annotations["preview.previewd.io/owner-name"] = preview.Name
		ns.Annotations["preview.previewd.io/owner-namespace"] = preview.Namespace
		ns.Annotations["preview.previewd.io/owner-uid"] = string(preview.UID)

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to ensure namespace: %w", err)
	}

	return nil
}

// EnsureResourceQuota creates or updates resource quotas in the namespace
// to limit resource consumption by the preview environment.
func (m *Manager) EnsureResourceQuota(ctx context.Context, preview *previewv1alpha1.PreviewEnvironment, namespace string) error {
	quota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "preview-quota",
			Namespace: namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, m.client, quota, func() error {
		// Set resource limits
		quota.Spec.Hard = corev1.ResourceList{
			corev1.ResourceRequestsCPU:            resource.MustParse("2"),
			corev1.ResourceRequestsMemory:         resource.MustParse("4Gi"),
			corev1.ResourceLimitsCPU:              resource.MustParse("4"),
			corev1.ResourceLimitsMemory:           resource.MustParse("8Gi"),
			corev1.ResourcePersistentVolumeClaims: resource.MustParse("0"),
			"services.loadbalancers":              resource.MustParse("0"),
		}

		// Add labels to associate with preview environment
		if quota.Labels == nil {
			quota.Labels = make(map[string]string)
		}
		quota.Labels["preview.previewd.io/pr"] = fmt.Sprintf("%d", preview.Spec.PRNumber)
		quota.Labels["preview.previewd.io/managed-by"] = managedByLabel

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to ensure resource quota: %w", err)
	}

	return nil
}

// EnsureNetworkPolicies creates network policies to isolate the preview environment
// and control ingress/egress traffic.
func (m *Manager) EnsureNetworkPolicies(ctx context.Context, preview *previewv1alpha1.PreviewEnvironment, namespace string) error {
	// Create default deny all policy
	if err := m.ensureDefaultDenyPolicy(ctx, preview, namespace); err != nil {
		return fmt.Errorf("failed to ensure default deny policy: %w", err)
	}

	// Create allow ingress from ingress controller
	if err := m.ensureAllowIngressPolicy(ctx, preview, namespace); err != nil {
		return fmt.Errorf("failed to ensure allow ingress policy: %w", err)
	}

	// Create allow egress for DNS and HTTPS
	if err := m.ensureAllowEgressPolicy(ctx, preview, namespace); err != nil {
		return fmt.Errorf("failed to ensure allow egress policy: %w", err)
	}

	return nil
}

// ensureDefaultDenyPolicy creates a NetworkPolicy that denies all ingress and egress by default
func (m *Manager) ensureDefaultDenyPolicy(ctx context.Context, preview *previewv1alpha1.PreviewEnvironment, namespace string) error {
	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-deny-all",
			Namespace: namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, m.client, policy, func() error {
		// Apply to all pods in namespace
		policy.Spec.PodSelector = metav1.LabelSelector{}

		// Set policy types to both Ingress and Egress
		policy.Spec.PolicyTypes = []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
			networkingv1.PolicyTypeEgress,
		}

		// Empty ingress and egress rules mean deny all
		policy.Spec.Ingress = []networkingv1.NetworkPolicyIngressRule{}
		policy.Spec.Egress = []networkingv1.NetworkPolicyEgressRule{}

		// Add labels to associate with preview environment
		if policy.Labels == nil {
			policy.Labels = make(map[string]string)
		}
		policy.Labels["preview.previewd.io/pr"] = fmt.Sprintf("%d", preview.Spec.PRNumber)
		policy.Labels["preview.previewd.io/managed-by"] = managedByLabel

		return nil
	})

	return err
}

// ensureAllowIngressPolicy creates a NetworkPolicy that allows ingress from the ingress controller
func (m *Manager) ensureAllowIngressPolicy(ctx context.Context, preview *previewv1alpha1.PreviewEnvironment, namespace string) error {
	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "allow-ingress",
			Namespace: namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, m.client, policy, func() error {
		// Apply to all pods in namespace
		policy.Spec.PodSelector = metav1.LabelSelector{}

		// Set policy type to Ingress
		policy.Spec.PolicyTypes = []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
		}

		// Allow ingress from ingress-nginx namespace on port 8080
		tcp := corev1.ProtocolTCP
		port8080 := intstr.FromInt(8080)
		policy.Spec.Ingress = []networkingv1.NetworkPolicyIngressRule{
			{
				From: []networkingv1.NetworkPolicyPeer{
					{
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"kubernetes.io/metadata.name": "ingress-nginx",
							},
						},
					},
				},
				Ports: []networkingv1.NetworkPolicyPort{
					{
						Protocol: &tcp,
						Port:     &port8080,
					},
				},
			},
		}

		// Add labels to associate with preview environment
		if policy.Labels == nil {
			policy.Labels = make(map[string]string)
		}
		policy.Labels["preview.previewd.io/pr"] = fmt.Sprintf("%d", preview.Spec.PRNumber)
		policy.Labels["preview.previewd.io/managed-by"] = managedByLabel

		return nil
	})

	return err
}

// ensureAllowEgressPolicy creates a NetworkPolicy that allows necessary egress traffic
func (m *Manager) ensureAllowEgressPolicy(ctx context.Context, preview *previewv1alpha1.PreviewEnvironment, namespace string) error {
	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "allow-egress",
			Namespace: namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, m.client, policy, func() error {
		// Apply to all pods in namespace
		policy.Spec.PodSelector = metav1.LabelSelector{}

		// Set policy type to Egress
		policy.Spec.PolicyTypes = []networkingv1.PolicyType{
			networkingv1.PolicyTypeEgress,
		}

		tcp := corev1.ProtocolTCP
		udp := corev1.ProtocolUDP
		port53 := intstr.FromInt(53)
		port443 := intstr.FromInt(443)
		port8080 := intstr.FromInt(8080)

		policy.Spec.Egress = []networkingv1.NetworkPolicyEgressRule{
			// Allow DNS to kube-system namespace
			{
				To: []networkingv1.NetworkPolicyPeer{
					{
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"kubernetes.io/metadata.name": "kube-system",
							},
						},
					},
				},
				Ports: []networkingv1.NetworkPolicyPort{
					{
						Protocol: &udp,
						Port:     &port53,
					},
				},
			},
			// Allow HTTPS to any destination
			{
				Ports: []networkingv1.NetworkPolicyPort{
					{
						Protocol: &tcp,
						Port:     &port443,
					},
				},
			},
			// Allow intra-namespace communication on port 8080
			{
				To: []networkingv1.NetworkPolicyPeer{
					{
						PodSelector: &metav1.LabelSelector{},
					},
				},
				Ports: []networkingv1.NetworkPolicyPort{
					{
						Protocol: &tcp,
						Port:     &port8080,
					},
				},
			},
		}

		// Add labels to associate with preview environment
		if policy.Labels == nil {
			policy.Labels = make(map[string]string)
		}
		policy.Labels["preview.previewd.io/pr"] = fmt.Sprintf("%d", preview.Spec.PRNumber)
		policy.Labels["preview.previewd.io/managed-by"] = managedByLabel

		return nil
	})

	return err
}

// Cleanup removes namespace and associated resources when a preview environment is deleted.
// The actual deletion is handled by Kubernetes garbage collection through owner references.
func (m *Manager) Cleanup(ctx context.Context, preview *previewv1alpha1.PreviewEnvironment) error {
	nsName := generateNamespaceName(preview.Spec.PRNumber, preview.Spec.Repository)

	ns := &corev1.Namespace{}
	err := m.client.Get(ctx, types.NamespacedName{Name: nsName}, ns)
	if err != nil {
		if errors.IsNotFound(err) {
			// Namespace already deleted
			return nil
		}
		return fmt.Errorf("failed to get namespace: %w", err)
	}

	// Delete the namespace - Kubernetes will handle cascade deletion of resources
	if err := m.client.Delete(ctx, ns); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	return nil
}

// GetNamespaceName returns the namespace name for a preview environment
func (m *Manager) GetNamespaceName(preview *previewv1alpha1.PreviewEnvironment) string {
	return generateNamespaceName(preview.Spec.PRNumber, preview.Spec.Repository)
}

// generateNamespaceName generates a deterministic namespace name from PR number and repository
func generateNamespaceName(prNumber int, repository string) string {
	// Create a hash of the repository to ensure uniqueness
	h := sha256.New()
	h.Write([]byte(repository))
	hash := fmt.Sprintf("%x", h.Sum(nil))[:8]

	// Format: preview-pr-{number}-{hash}
	// This ensures unique namespaces even if multiple repositories use same PR numbers
	return fmt.Sprintf("preview-pr-%d-%s", prNumber, hash)
}
