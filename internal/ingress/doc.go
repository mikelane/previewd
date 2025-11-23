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

// Package ingress provides management of Kubernetes Ingress resources for preview environments.
//
// # Overview
//
// The ingress package handles HTTP(S) routing to preview services through Kubernetes Ingress
// resources. It integrates with cert-manager for automatic TLS certificate provisioning and
// external-dns for automatic DNS record creation.
//
// # Key Features
//
//   - Automatic host generation (pr-{number}.{baseDomain})
//   - TLS certificate management via cert-manager
//   - DNS record creation via external-dns
//   - Path-based routing to multiple services
//   - Owner references for cascade deletion
//   - Idempotent operations
//
// # Usage
//
// Create a manager with your cluster client and configuration:
//
//	mgr := ingress.NewManager(k8sClient, scheme, "preview.example.com", "letsencrypt-prod")
//
// Ensure an ingress exists for a preview environment:
//
//	err := mgr.EnsureIngress(ctx, previewEnv, namespace)
//	if err != nil {
//	    // Handle error
//	}
//
// Get the public URL for a preview environment:
//
//	host := mgr.GetIngressHost(previewEnv)
//	url := fmt.Sprintf("https://%s", host)
//
// # Integration with cert-manager
//
// The manager creates Ingress resources with the cert-manager.io/cluster-issuer annotation,
// which triggers cert-manager to automatically provision TLS certificates. The certificate
// is stored in a Kubernetes Secret referenced by the Ingress TLS configuration.
//
// # Integration with external-dns
//
// The manager adds the external-dns.alpha.kubernetes.io/hostname annotation to Ingress
// resources, which triggers external-dns to create DNS A records pointing to the Ingress
// controller's LoadBalancer IP.
//
// # Path-based Routing
//
// Services are exposed through path-based routing:
//   - frontend service: / (root path)
//   - other services: /{service-name}
//
// For example, with services ["auth", "api", "frontend"], the Ingress routes:
//   - /auth → preview-pr-123-auth:8080
//   - /api → preview-pr-123-api:8080
//   - / → preview-pr-123-frontend:8080
//
// # Owner References
//
// Ingress resources are created with owner references to their PreviewEnvironment.
// When the PreviewEnvironment is deleted, Kubernetes automatically garbage collects
// the associated Ingress resource.
package ingress
