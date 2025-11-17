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

// Package namespace provides namespace management functionality for preview environments.
//
// The package implements a Manager that handles the complete lifecycle of Kubernetes
// namespaces for preview environments, including:
//
//   - Namespace creation with appropriate labels for identification
//   - Resource quotas to limit CPU, memory, and other resources
//   - Network policies for security isolation
//   - Cleanup when preview environments are deleted
//
// # Architecture
//
// The namespace manager follows these design principles:
//
//   - Idempotency: All operations can be safely retried
//   - Isolation: Each preview environment gets its own namespace
//   - Security: Default-deny network policies with selective allow rules
//   - Resource Control: Quotas prevent resource exhaustion
//   - Garbage Collection: Resources are cleaned up with the namespace
//
// # Namespace Naming
//
// Namespaces are named using a deterministic pattern:
//
//	preview-pr-{PR-NUMBER}-{REPO-HASH}
//
// Where REPO-HASH is the first 8 characters of the SHA256 hash of the repository name.
// This ensures unique namespaces even when multiple repositories use the same PR numbers.
//
// # Resource Quotas
//
// Each namespace gets a ResourceQuota with the following defaults:
//
//   - CPU Requests: 2 cores
//   - Memory Requests: 4Gi
//   - CPU Limits: 4 cores
//   - Memory Limits: 8Gi
//   - Persistent Volume Claims: 0 (no persistent storage by default)
//   - LoadBalancer Services: 0 (use Ingress instead)
//
// # Network Policies
//
// Three NetworkPolicies are created for security isolation:
//
//  1. default-deny-all: Denies all ingress and egress by default
//  2. allow-ingress: Allows ingress from the ingress-nginx namespace on port 8080
//  3. allow-egress: Allows DNS (UDP 53), HTTPS (TCP 443), and intra-namespace communication
//
// # Ownership and Deletion
//
// Due to Kubernetes limitations on cross-namespace owner references, namespaces cannot
// have owner references to PreviewEnvironment resources in different namespaces.
// Instead, we use:
//
//   - Labels to associate namespaces with preview environments
//   - Annotations to track the owner information
//   - Explicit cleanup in the Cleanup() method
//
// The controller must call Cleanup() when a PreviewEnvironment is being deleted.
//
// # Usage Example
//
//	// Create a manager
//	mgr := namespace.NewManager(k8sClient, scheme)
//
//	// Create namespace and resources for a preview environment
//	err := mgr.EnsureNamespace(ctx, preview)
//	if err != nil {
//	    return err
//	}
//
//	nsName := mgr.GetNamespaceName(preview)
//	err = mgr.EnsureResourceQuota(ctx, preview, nsName)
//	if err != nil {
//	    return err
//	}
//
//	err = mgr.EnsureNetworkPolicies(ctx, preview, nsName)
//	if err != nil {
//	    return err
//	}
//
//	// Clean up when done
//	err = mgr.Cleanup(ctx, preview)
//	if err != nil {
//	    return err
//	}
package namespace
