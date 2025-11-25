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

// Package argocd provides functionality for managing ArgoCD ApplicationSets
// for preview environments, enabling GitOps-based deployment of services.
package argocd

import (
	"context"
	"encoding/json"
	"fmt"

	previewv1alpha1 "github.com/mikelane/previewd/api/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	managedByLabel = "previewd"

	// InClusterServer is the default in-cluster Kubernetes API server URL
	InClusterServer = "https://kubernetes.default.svc"
)

// ApplicationStatusInfo contains the health and sync status of an ArgoCD Application
type ApplicationStatusInfo struct {
	// Health is the health status (e.g., "Healthy", "Progressing", "Degraded")
	Health string
	// Sync is the sync status (e.g., "Synced", "OutOfSync")
	Sync string
	// Message is an optional message with more details
	Message string
}

// Manager handles ArgoCD ApplicationSet lifecycle for preview environments
type Manager struct {
	client          client.Client
	scheme          *runtime.Scheme
	repoURL         string
	argocdNamespace string
	project         string
}

// NewManager creates a new ArgoCD manager
func NewManager(c client.Client, scheme *runtime.Scheme, repoURL, argocdNamespace, project string) *Manager {
	return &Manager{
		client:          c,
		scheme:          scheme,
		repoURL:         repoURL,
		argocdNamespace: argocdNamespace,
		project:         project,
	}
}

// BuildApplicationSet creates an ApplicationSet resource for a preview environment.
// It generates one Application per service using a list generator.
func (m *Manager) BuildApplicationSet(preview *previewv1alpha1.PreviewEnvironment, namespace string) *ApplicationSet {
	prNumber := preview.Spec.PRNumber
	appSetName := m.GetApplicationSetName(prNumber)

	// Build list generator elements - one per service
	elements := make([]apiextensionsv1.JSON, len(preview.Spec.Services))
	for i, service := range preview.Spec.Services {
		elementData := map[string]string{
			"service": service,
		}
		// json.Marshal on map[string]string never fails
		raw, _ := json.Marshal(elementData) //nolint:errcheck
		elements[i] = apiextensionsv1.JSON{Raw: raw}
	}

	appSet := &ApplicationSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "argoproj.io/v1alpha1",
			Kind:       "ApplicationSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appSetName,
			Namespace: m.argocdNamespace,
			Labels: map[string]string{
				"preview.previewd.io/pr":         fmt.Sprintf("%d", prNumber),
				"preview.previewd.io/managed-by": managedByLabel,
			},
			Annotations: map[string]string{
				"preview.previewd.io/owner-name":      preview.Name,
				"preview.previewd.io/owner-namespace": preview.Namespace,
				"preview.previewd.io/owner-uid":       string(preview.UID),
			},
		},
		Spec: ApplicationSetSpec{
			GoTemplate: true,
			GoTemplateOptions: []string{
				"missingkey=error",
			},
			Generators: []ApplicationSetGenerator{
				{
					List: &ListGenerator{
						Elements: elements,
					},
				},
			},
			Template: ApplicationSetTemplate{
				ApplicationSetTemplateMeta: ApplicationSetTemplateMeta{
					Name: fmt.Sprintf("preview-%d-{{service}}", prNumber),
					Labels: map[string]string{
						"preview.previewd.io/pr":         fmt.Sprintf("%d", prNumber),
						"preview.previewd.io/service":    "{{service}}",
						"preview.previewd.io/managed-by": managedByLabel,
					},
				},
				Spec: ApplicationSpec{
					Project: m.project,
					Source: &ApplicationSource{
						RepoURL:        m.repoURL,
						Path:           "services/{{service}}",
						TargetRevision: preview.Spec.HeadSHA,
						Kustomize: &ApplicationSourceKustomize{
							NamePrefix: fmt.Sprintf("pr-%d-", prNumber),
							Namespace:  namespace,
							CommonLabels: map[string]string{
								"preview.previewd.io/pr":         fmt.Sprintf("%d", prNumber),
								"preview.previewd.io/service":    "{{service}}",
								"preview.previewd.io/managed-by": managedByLabel,
							},
						},
					},
					Destination: ApplicationDestination{
						Server:    InClusterServer,
						Namespace: namespace,
					},
					SyncPolicy: &SyncPolicy{
						Automated: &SyncPolicyAutomated{
							Prune:    true,
							SelfHeal: true,
						},
						SyncOptions: []string{
							"CreateNamespace=false",
							"PruneLast=true",
						},
						Retry: &RetryStrategy{
							Limit: 5,
							Backoff: &Backoff{
								Duration:    "5s",
								MaxDuration: "3m",
							},
						},
					},
				},
			},
		},
	}

	return appSet
}

// EnsureApplicationSet creates or updates an ApplicationSet for the preview environment.
func (m *Manager) EnsureApplicationSet(ctx context.Context, preview *previewv1alpha1.PreviewEnvironment, namespace string) error {
	// Input validation
	if preview == nil {
		return fmt.Errorf("preview environment cannot be nil")
	}
	if namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}

	appSetName := m.GetApplicationSetName(preview.Spec.PRNumber)

	appSet := &ApplicationSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appSetName,
			Namespace: m.argocdNamespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, m.client, appSet, func() error {
		// Build the desired state
		desired := m.BuildApplicationSet(preview, namespace)

		// Copy spec from desired to actual
		appSet.Spec = desired.Spec

		// Set labels
		if appSet.Labels == nil {
			appSet.Labels = make(map[string]string)
		}
		for k, v := range desired.Labels {
			appSet.Labels[k] = v
		}

		// Set annotations for owner tracking (cross-namespace owner refs not allowed)
		if appSet.Annotations == nil {
			appSet.Annotations = make(map[string]string)
		}
		for k, v := range desired.Annotations {
			appSet.Annotations[k] = v
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to ensure ApplicationSet for preview %s/%s (PR #%d): %w",
			preview.Namespace, preview.Name, preview.Spec.PRNumber, err)
	}

	return nil
}

// DeleteApplicationSet removes an ApplicationSet by name.
// Returns nil if the ApplicationSet doesn't exist (idempotent).
func (m *Manager) DeleteApplicationSet(ctx context.Context, name, namespace string) error {
	appSet := &ApplicationSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err := m.client.Delete(ctx, appSet)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete ApplicationSet %s/%s: %w", namespace, name, err)
	}

	return nil
}

// GetApplicationStatus retrieves the health and sync status of an ArgoCD Application.
func (m *Manager) GetApplicationStatus(ctx context.Context, name, namespace string) (*ApplicationStatusInfo, error) {
	app := &Application{}
	err := m.client.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, app)
	if err != nil {
		return nil, err
	}

	return &ApplicationStatusInfo{
		Health:  app.Status.Health.Status,
		Sync:    app.Status.Sync.Status,
		Message: app.Status.Health.Message,
	}, nil
}

// GetApplicationSetName generates the ApplicationSet name for a PR number.
func (m *Manager) GetApplicationSetName(prNumber int) string {
	return fmt.Sprintf("preview-%d", prNumber)
}

// GetArgocdNamespace returns the ArgoCD namespace configured for this manager.
func (m *Manager) GetArgocdNamespace() string {
	return m.argocdNamespace
}
