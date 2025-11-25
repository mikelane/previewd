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

// Package argocd provides management of ArgoCD ApplicationSets for preview environments.
//
// # Overview
//
// The argocd package handles GitOps-based deployment of preview services through ArgoCD
// ApplicationSets. It generates one Application per service using a list generator, enabling
// automatic deployment and synchronization of preview environment services.
//
// # Key Features
//
//   - ApplicationSet generation with list generator
//   - One Application per service in spec.services
//   - Go templating for dynamic service configuration
//   - Automated sync with prune and self-heal
//   - Kustomize integration for namespace isolation
//   - Owner reference tracking via annotations (cross-namespace)
//   - Idempotent operations
//
// # Usage
//
// Create a manager with your cluster client and configuration:
//
//	mgr := argocd.NewManager(
//	    k8sClient,
//	    scheme,
//	    "https://github.com/example/app",
//	    "argocd",
//	    "default",
//	)
//
// Build an ApplicationSet for a preview environment:
//
//	appSet := mgr.BuildApplicationSet(previewEnv, namespace)
//
// Ensure an ApplicationSet exists (create or update):
//
//	err := mgr.EnsureApplicationSet(ctx, previewEnv, namespace)
//	if err != nil {
//	    // Handle error
//	}
//
// Delete an ApplicationSet:
//
//	err := mgr.DeleteApplicationSet(ctx, "preview-123", "argocd")
//	if err != nil {
//	    // Handle error
//	}
//
// Get Application status:
//
//	status, err := mgr.GetApplicationStatus(ctx, "preview-123-auth", "argocd")
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Printf("Health: %s, Sync: %s\n", status.Health, status.Sync)
//
// # ApplicationSet Structure
//
// The generated ApplicationSet uses the following structure:
//
//	apiVersion: argoproj.io/v1alpha1
//	kind: ApplicationSet
//	metadata:
//	  name: preview-{prNumber}
//	  namespace: argocd
//	spec:
//	  goTemplate: true
//	  generators:
//	  - list:
//	      elements:
//	      - service: auth
//	      - service: api
//	      - service: frontend
//	  template:
//	    metadata:
//	      name: preview-{prNumber}-{{service}}
//	    spec:
//	      project: default
//	      source:
//	        repoURL: https://github.com/example/app
//	        path: services/{{service}}
//	        targetRevision: {headSHA}
//	        kustomize:
//	          namePrefix: pr-{prNumber}-
//	          namespace: {namespace}
//	          commonLabels:
//	            preview.previewd.io/pr: "{prNumber}"
//	      destination:
//	        server: https://kubernetes.default.svc
//	        namespace: {namespace}
//	      syncPolicy:
//	        automated:
//	          prune: true
//	          selfHeal: true
//	        syncOptions:
//	        - CreateNamespace=false
//
// # Sync Policy
//
// The ApplicationSet configures automated sync with:
//   - Prune: true - removes resources not in Git
//   - SelfHeal: true - reverts manual changes in cluster
//   - CreateNamespace: false - namespace created by previewd, not ArgoCD
//
// # Owner Reference Tracking
//
// Since ApplicationSets are created in the argocd namespace and PreviewEnvironments
// are typically in a different namespace, cross-namespace owner references are not
// allowed in Kubernetes. Instead, ownership is tracked via annotations:
//
//   - preview.previewd.io/owner-name: PreviewEnvironment name
//   - preview.previewd.io/owner-namespace: PreviewEnvironment namespace
//   - preview.previewd.io/owner-uid: PreviewEnvironment UID
//
// The previewd controller uses these annotations to clean up ApplicationSets
// when their owning PreviewEnvironment is deleted.
//
// # Types
//
// This package defines minimal ArgoCD types (ApplicationSet, Application) to avoid
// pulling in the full ArgoCD dependency tree which has complex transitive dependencies.
// These types are compatible with the ArgoCD CRDs and can be used with the
// controller-runtime client.
package argocd
