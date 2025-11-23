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

// Package ingress provides functionality for managing Kubernetes Ingress resources
// for preview environments, including TLS configuration via cert-manager and DNS
// management via external-dns.
package ingress

import (
	"context"
	"fmt"

	previewv1alpha1 "github.com/mikelane/previewd/api/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// IngressName is the name of the Ingress resource
	IngressName = "preview-ingress"

	// DefaultServicePort is the default port for services
	DefaultServicePort = 8080

	// CertManagerIssuerAnnotation is the annotation key for cert-manager cluster issuer
	CertManagerIssuerAnnotation = "cert-manager.io/cluster-issuer"

	// ExternalDNSHostnameAnnotation is the annotation key for external-dns hostname
	ExternalDNSHostnameAnnotation = "external-dns.alpha.kubernetes.io/hostname"

	// SSLRedirectAnnotation is the annotation key for nginx SSL redirect
	SSLRedirectAnnotation = "nginx.ingress.kubernetes.io/ssl-redirect"

	managedByLabel = "previewd"
)

// Manager handles Ingress lifecycle for preview environments
type Manager struct {
	client     client.Client
	scheme     *runtime.Scheme
	baseDomain string
	certIssuer string
}

// NewManager creates a new Ingress manager
func NewManager(c client.Client, scheme *runtime.Scheme, baseDomain, certIssuer string) *Manager {
	return &Manager{
		client:     c,
		scheme:     scheme,
		baseDomain: baseDomain,
		certIssuer: certIssuer,
	}
}

// EnsureIngress creates or updates an Ingress resource for the preview environment
// with TLS certificates (cert-manager) and DNS routing (external-dns).
func (m *Manager) EnsureIngress(ctx context.Context, preview *previewv1alpha1.PreviewEnvironment, namespace string) error {
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      IngressName,
			Namespace: namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, m.client, ingress, func() error {
		// Set labels
		if ingress.Labels == nil {
			ingress.Labels = make(map[string]string)
		}
		ingress.Labels["preview.previewd.io/pr"] = fmt.Sprintf("%d", preview.Spec.PRNumber)
		ingress.Labels["preview.previewd.io/repository"] = preview.Spec.Repository
		ingress.Labels["preview.previewd.io/managed-by"] = managedByLabel

		// Set annotations for cert-manager, external-dns, and nginx
		if ingress.Annotations == nil {
			ingress.Annotations = make(map[string]string)
		}
		host := m.GetIngressHost(preview)
		ingress.Annotations[CertManagerIssuerAnnotation] = m.certIssuer
		ingress.Annotations[ExternalDNSHostnameAnnotation] = host
		ingress.Annotations[SSLRedirectAnnotation] = "true"

		// Add annotations to track the owner (informational only, since cross-namespace owner references are not allowed)
		ingress.Annotations["preview.previewd.io/owner-name"] = preview.Name
		ingress.Annotations["preview.previewd.io/owner-namespace"] = preview.Namespace
		ingress.Annotations["preview.previewd.io/owner-uid"] = string(preview.UID)

		// Build TLS configuration
		tlsSecretName := fmt.Sprintf("pr-%d-tls", preview.Spec.PRNumber)
		ingress.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts:      []string{host},
				SecretName: tlsSecretName,
			},
		}

		// Build HTTP rules with path-based routing
		pathType := networkingv1.PathTypePrefix
		var paths []networkingv1.HTTPIngressPath

		for _, service := range preview.Spec.Services {
			serviceName := generateServiceName(preview.Spec.PRNumber, service)
			path := generatePathForService(service)

			paths = append(paths, networkingv1.HTTPIngressPath{
				Path:     path,
				PathType: &pathType,
				Backend: networkingv1.IngressBackend{
					Service: &networkingv1.IngressServiceBackend{
						Name: serviceName,
						Port: networkingv1.ServiceBackendPort{
							Number: DefaultServicePort,
						},
					},
				},
			})
		}

		ingress.Spec.Rules = []networkingv1.IngressRule{
			{
				Host: host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: paths,
					},
				},
			},
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to ensure ingress: %w", err)
	}

	return nil
}

// GetIngressHost returns the hostname for the preview environment ingress
func (m *Manager) GetIngressHost(preview *previewv1alpha1.PreviewEnvironment) string {
	return fmt.Sprintf("pr-%d.%s", preview.Spec.PRNumber, m.baseDomain)
}

// generateServiceName generates the service name for a given PR number and service
func generateServiceName(prNumber int, service string) string {
	return fmt.Sprintf("preview-pr-%d-%s", prNumber, service)
}

// generatePathForService generates the path for a service
// frontend gets "/", other services get "/<service-name>"
func generatePathForService(service string) string {
	if service == "frontend" {
		return "/"
	}
	return fmt.Sprintf("/%s", service)
}
