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

package argocd

import (
	"context"
	"encoding/json"
	"testing"

	previewv1alpha1 "github.com/mikelane/previewd/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// setupTestClient creates a fake Kubernetes client with necessary schemes
func setupTestClient(t *testing.T) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := previewv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add preview scheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add core scheme: %v", err)
	}
	if err := AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add argocd scheme: %v", err)
	}

	return fake.NewClientBuilder().
		WithScheme(scheme).
		Build()
}

func TestNewManager(t *testing.T) {
	c := setupTestClient(t)
	repoURL := "https://github.com/example/app"
	argocdNamespace := "argocd"
	project := "default"

	m := NewManager(c, c.Scheme(), repoURL, argocdNamespace, project)

	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	if m.client == nil {
		t.Error("Manager client is nil")
	}
	if m.scheme == nil {
		t.Error("Manager scheme is nil")
	}
	if m.repoURL != repoURL {
		t.Errorf("repoURL = %v, want %v", m.repoURL, repoURL)
	}
	if m.argocdNamespace != argocdNamespace {
		t.Errorf("argocdNamespace = %v, want %v", m.argocdNamespace, argocdNamespace)
	}
	if m.project != project {
		t.Errorf("project = %v, want %v", m.project, project)
	}
}

// TestBuildApplicationSet_Name verifies the ApplicationSet name follows the pattern "preview-{prNumber}"
func TestBuildApplicationSet_Name(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth", "api", "frontend"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-123-abc12345")

	expectedName := "preview-123"
	if appSet.Name != expectedName {
		t.Errorf("ApplicationSet name = %v, want %v", appSet.Name, expectedName)
	}
}

// TestBuildApplicationSet_Namespace verifies the ApplicationSet is created in the argocd namespace
func TestBuildApplicationSet_Namespace(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth", "api", "frontend"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-123-abc12345")

	if appSet.Namespace != "argocd" {
		t.Errorf("ApplicationSet namespace = %v, want %v", appSet.Namespace, "argocd")
	}
}

// TestBuildApplicationSet_ListGeneratorElements verifies the list generator has elements for each service
func TestBuildApplicationSet_ListGeneratorElements(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth", "api", "frontend"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-123-abc12345")

	if len(appSet.Spec.Generators) == 0 {
		t.Fatal("ApplicationSet has no generators")
	}

	listGen := appSet.Spec.Generators[0].List
	if listGen == nil {
		t.Fatal("ApplicationSet does not have a list generator")
	}

	if len(listGen.Elements) != 3 {
		t.Errorf("List generator elements count = %v, want %v", len(listGen.Elements), 3)
	}

	// Verify each service is represented in the elements
	expectedServices := map[string]bool{"auth": false, "api": false, "frontend": false}
	for _, element := range listGen.Elements {
		var data map[string]interface{}
		if err := json.Unmarshal(element.Raw, &data); err != nil {
			t.Fatalf("failed to unmarshal element: %v", err)
		}
		if service, ok := data["service"].(string); ok {
			expectedServices[service] = true
		}
	}

	for service, found := range expectedServices {
		if !found {
			t.Errorf("service %q not found in list generator elements", service)
		}
	}
}

// TestBuildApplicationSet_AutomatedSyncPolicy verifies automated sync with prune and selfHeal
func TestBuildApplicationSet_AutomatedSyncPolicy(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-456",
			Namespace: "previewd-system",
			UID:       "def-456",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   456,
			Repository: "example/app",
			HeadSHA:    "def456abc789012345678901234567890abcdef",
			Services:   []string{"api"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-456-def45678")

	syncPolicy := appSet.Spec.Template.Spec.SyncPolicy
	if syncPolicy == nil {
		t.Fatal("ApplicationSet template has no sync policy")
	}

	if syncPolicy.Automated == nil {
		t.Fatal("Sync policy has no automated configuration")
	}

	if !syncPolicy.Automated.Prune {
		t.Error("Automated sync policy prune = false, want true")
	}

	if !syncPolicy.Automated.SelfHeal {
		t.Error("Automated sync policy selfHeal = false, want true")
	}
}

// TestBuildApplicationSet_SyncOptions verifies CreateNamespace=false is in sync options
func TestBuildApplicationSet_SyncOptions(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-789",
			Namespace: "previewd-system",
			UID:       "ghi-789",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   789,
			Repository: "example/app",
			HeadSHA:    "ghi789jkl012345678901234567890abcdefgh",
			Services:   []string{"frontend"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-789-ghi78901")

	syncPolicy := appSet.Spec.Template.Spec.SyncPolicy
	if syncPolicy == nil {
		t.Fatal("ApplicationSet template has no sync policy")
	}

	found := false
	for _, opt := range syncPolicy.SyncOptions {
		if opt == "CreateNamespace=false" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Sync options does not include CreateNamespace=false")
	}
}

// TestBuildApplicationSet_OwnerReference verifies owner reference is set correctly via annotations
func TestBuildApplicationSet_OwnerReference(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123-uid",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-123-abc12345")

	// Note: Cross-namespace owner references are not allowed in Kubernetes,
	// so we use annotations to track the owner instead
	if appSet.Annotations == nil {
		t.Fatal("ApplicationSet has no annotations")
	}

	if appSet.Annotations["preview.previewd.io/owner-uid"] != string(preview.UID) {
		t.Errorf("owner UID annotation = %v, want %v",
			appSet.Annotations["preview.previewd.io/owner-uid"], preview.UID)
	}

	if appSet.Annotations["preview.previewd.io/owner-name"] != preview.Name {
		t.Errorf("owner name annotation = %v, want %v",
			appSet.Annotations["preview.previewd.io/owner-name"], preview.Name)
	}

	if appSet.Annotations["preview.previewd.io/owner-namespace"] != preview.Namespace {
		t.Errorf("owner namespace annotation = %v, want %v",
			appSet.Annotations["preview.previewd.io/owner-namespace"], preview.Namespace)
	}
}

// TestBuildApplicationSet_KustomizeNamespace verifies Kustomize namespace settings
func TestBuildApplicationSet_KustomizeNamespace(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	namespace := "preview-pr-123-abc12345"
	appSet := m.BuildApplicationSet(preview, namespace)

	source := appSet.Spec.Template.Spec.Source
	if source == nil {
		t.Fatal("ApplicationSet template has no source")
	}

	if source.Kustomize == nil {
		t.Fatal("ApplicationSet template source has no Kustomize config")
	}

	// Verify namespace is set
	if source.Kustomize.Namespace != namespace {
		t.Errorf("Kustomize namespace = %v, want %v", source.Kustomize.Namespace, namespace)
	}
}

// TestBuildApplicationSet_KustomizeNamePrefix verifies Kustomize name prefix
func TestBuildApplicationSet_KustomizeNamePrefix(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-123-abc12345")

	source := appSet.Spec.Template.Spec.Source
	if source == nil {
		t.Fatal("ApplicationSet template has no source")
	}

	if source.Kustomize == nil {
		t.Fatal("ApplicationSet template source has no Kustomize config")
	}

	expectedPrefix := "pr-123-"
	if source.Kustomize.NamePrefix != expectedPrefix {
		t.Errorf("Kustomize namePrefix = %v, want %v", source.Kustomize.NamePrefix, expectedPrefix)
	}
}

// TestBuildApplicationSet_KustomizeCommonLabels verifies Kustomize common labels
func TestBuildApplicationSet_KustomizeCommonLabels(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-123-abc12345")

	source := appSet.Spec.Template.Spec.Source
	if source == nil {
		t.Fatal("ApplicationSet template has no source")
	}

	if source.Kustomize == nil {
		t.Fatal("ApplicationSet template source has no Kustomize config")
	}

	if source.Kustomize.CommonLabels == nil {
		t.Fatal("Kustomize has no common labels")
	}

	expectedPRLabel := "123"
	if source.Kustomize.CommonLabels["preview.previewd.io/pr"] != expectedPRLabel {
		t.Errorf("Kustomize commonLabels[preview.previewd.io/pr] = %v, want %v",
			source.Kustomize.CommonLabels["preview.previewd.io/pr"], expectedPRLabel)
	}
}

// TestBuildApplicationSet_ApplicationNamePattern verifies the generated Application name pattern
func TestBuildApplicationSet_ApplicationNamePattern(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-123-abc12345")

	// The template metadata name should follow the pattern "preview-{prNumber}-{{service}}"
	expectedNamePattern := "preview-123-{{service}}"
	if appSet.Spec.Template.Name != expectedNamePattern {
		t.Errorf("Template name pattern = %v, want %v", appSet.Spec.Template.Name, expectedNamePattern)
	}
}

// TestBuildApplicationSet_DestinationNamespace verifies the destination namespace
func TestBuildApplicationSet_DestinationNamespace(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	namespace := "preview-pr-123-abc12345"
	appSet := m.BuildApplicationSet(preview, namespace)

	if appSet.Spec.Template.Spec.Destination.Namespace != namespace {
		t.Errorf("Destination namespace = %v, want %v",
			appSet.Spec.Template.Spec.Destination.Namespace, namespace)
	}
}

// TestBuildApplicationSet_Labels verifies the ApplicationSet has proper labels
func TestBuildApplicationSet_Labels(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-123-abc12345")

	if appSet.Labels == nil {
		t.Fatal("ApplicationSet has no labels")
	}

	if appSet.Labels["preview.previewd.io/pr"] != "123" {
		t.Errorf("label preview.previewd.io/pr = %v, want %v",
			appSet.Labels["preview.previewd.io/pr"], "123")
	}

	if appSet.Labels["preview.previewd.io/managed-by"] != "previewd" {
		t.Errorf("label preview.previewd.io/managed-by = %v, want %v",
			appSet.Labels["preview.previewd.io/managed-by"], "previewd")
	}
}

// TestBuildApplicationSet_SourcePath verifies the source path includes service templating
func TestBuildApplicationSet_SourcePath(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-123-abc12345")

	source := appSet.Spec.Template.Spec.Source
	if source == nil {
		t.Fatal("ApplicationSet template has no source")
	}

	// The path should use Go templating for the service name
	expectedPath := "services/{{service}}"
	if source.Path != expectedPath {
		t.Errorf("Source path = %v, want %v", source.Path, expectedPath)
	}
}

// TestBuildApplicationSet_GoTemplate verifies Go templating is enabled
func TestBuildApplicationSet_GoTemplate(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-123-abc12345")

	if !appSet.Spec.GoTemplate {
		t.Error("GoTemplate = false, want true")
	}
}

// TestEnsureApplicationSet_Creates creates the ApplicationSet if it doesn't exist
func TestEnsureApplicationSet_Creates(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth", "api"},
		},
	}

	err := m.EnsureApplicationSet(context.Background(), preview, "preview-pr-123-abc12345")
	if err != nil {
		t.Fatalf("EnsureApplicationSet() error = %v", err)
	}

	// Verify the ApplicationSet was created
	appSet := &ApplicationSet{}
	err = c.Get(context.Background(), types.NamespacedName{
		Name:      "preview-123",
		Namespace: "argocd",
	}, appSet)
	if err != nil {
		t.Fatalf("failed to get created ApplicationSet: %v", err)
	}

	if appSet.Name != "preview-123" {
		t.Errorf("ApplicationSet name = %v, want %v", appSet.Name, "preview-123")
	}
}

// TestEnsureApplicationSet_Idempotent verifies calling twice doesn't error
func TestEnsureApplicationSet_Idempotent(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	// First call
	err := m.EnsureApplicationSet(context.Background(), preview, "preview-pr-123-abc12345")
	if err != nil {
		t.Fatalf("first EnsureApplicationSet() error = %v", err)
	}

	// Second call should not error
	err = m.EnsureApplicationSet(context.Background(), preview, "preview-pr-123-abc12345")
	if err != nil {
		t.Fatalf("second EnsureApplicationSet() error = %v", err)
	}
}

// TestDeleteApplicationSet_Deletes verifies deletion works
func TestDeleteApplicationSet_Deletes(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	// Create an ApplicationSet first
	appSet := &ApplicationSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "preview-123",
			Namespace: "argocd",
		},
	}
	if err := c.Create(context.Background(), appSet); err != nil {
		t.Fatalf("failed to create ApplicationSet: %v", err)
	}

	// Delete it
	err := m.DeleteApplicationSet(context.Background(), "preview-123", "argocd")
	if err != nil {
		t.Fatalf("DeleteApplicationSet() error = %v", err)
	}

	// Verify it's deleted
	err = c.Get(context.Background(), types.NamespacedName{
		Name:      "preview-123",
		Namespace: "argocd",
	}, &ApplicationSet{})
	if !errors.IsNotFound(err) {
		t.Error("ApplicationSet should be deleted")
	}
}

// TestDeleteApplicationSet_NotFoundNoError verifies deleting non-existent ApplicationSet doesn't error
func TestDeleteApplicationSet_NotFoundNoError(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	err := m.DeleteApplicationSet(context.Background(), "non-existent", "argocd")
	if err != nil {
		t.Errorf("DeleteApplicationSet() for non-existent should not error, got %v", err)
	}
}

// TestGetApplicationStatus_Healthy verifies getting status for a healthy application
func TestGetApplicationStatus_Healthy(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	// Create an Application with Healthy status
	app := &Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "preview-123-auth",
			Namespace: "argocd",
		},
		Status: ApplicationStatus{
			Health: HealthStatus{
				Status: "Healthy",
			},
			Sync: SyncStatus{
				Status: "Synced",
			},
		},
	}
	if err := c.Create(context.Background(), app); err != nil {
		t.Fatalf("failed to create Application: %v", err)
	}

	status, err := m.GetApplicationStatus(context.Background(), "preview-123-auth", "argocd")
	if err != nil {
		t.Fatalf("GetApplicationStatus() error = %v", err)
	}

	if status.Health != "Healthy" {
		t.Errorf("Health = %v, want %v", status.Health, "Healthy")
	}

	if status.Sync != "Synced" {
		t.Errorf("Sync = %v, want %v", status.Sync, "Synced")
	}
}

// TestGetApplicationStatus_Progressing verifies getting status for a progressing application
func TestGetApplicationStatus_Progressing(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	// Create an Application with Progressing status
	app := &Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "preview-123-api",
			Namespace: "argocd",
		},
		Status: ApplicationStatus{
			Health: HealthStatus{
				Status: "Progressing",
			},
			Sync: SyncStatus{
				Status: "OutOfSync",
			},
		},
	}
	if err := c.Create(context.Background(), app); err != nil {
		t.Fatalf("failed to create Application: %v", err)
	}

	status, err := m.GetApplicationStatus(context.Background(), "preview-123-api", "argocd")
	if err != nil {
		t.Fatalf("GetApplicationStatus() error = %v", err)
	}

	if status.Health != "Progressing" {
		t.Errorf("Health = %v, want %v", status.Health, "Progressing")
	}
}

// TestGetApplicationStatus_NotFound verifies error for non-existent application
func TestGetApplicationStatus_NotFound(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	_, err := m.GetApplicationStatus(context.Background(), "preview-999-missing", "argocd")
	if err == nil {
		t.Error("GetApplicationStatus() should return error for non-existent application")
	}

	if !errors.IsNotFound(err) {
		t.Errorf("error should be NotFound, got %v", err)
	}
}

// TestBuildApplicationSet_Project verifies the project is set correctly
func TestBuildApplicationSet_Project(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "preview-project")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-123-abc12345")

	if appSet.Spec.Template.Spec.Project != "preview-project" {
		t.Errorf("Project = %v, want %v", appSet.Spec.Template.Spec.Project, "preview-project")
	}
}

// TestBuildApplicationSet_TargetRevision verifies the target revision uses HEAD SHA
func TestBuildApplicationSet_TargetRevision(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-123-abc12345")

	source := appSet.Spec.Template.Spec.Source
	if source == nil {
		t.Fatal("ApplicationSet template has no source")
	}

	if source.TargetRevision != preview.Spec.HeadSHA {
		t.Errorf("TargetRevision = %v, want %v", source.TargetRevision, preview.Spec.HeadSHA)
	}
}

// TestBuildApplicationSet_RepoURL verifies the repository URL is set correctly
func TestBuildApplicationSet_RepoURL(t *testing.T) {
	c := setupTestClient(t)
	expectedRepoURL := "https://github.com/example/app"
	m := NewManager(c, c.Scheme(), expectedRepoURL, "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-123-abc12345")

	source := appSet.Spec.Template.Spec.Source
	if source == nil {
		t.Fatal("ApplicationSet template has no source")
	}

	if source.RepoURL != expectedRepoURL {
		t.Errorf("RepoURL = %v, want %v", source.RepoURL, expectedRepoURL)
	}
}

// TestEnsureApplicationSet_NilPreview tests validation
func TestEnsureApplicationSet_NilPreview(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	err := m.EnsureApplicationSet(context.Background(), nil, "preview-ns")
	if err == nil {
		t.Error("EnsureApplicationSet() should return error for nil preview")
	}
}

// TestEnsureApplicationSet_EmptyNamespace tests validation
func TestEnsureApplicationSet_EmptyNamespace(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	err := m.EnsureApplicationSet(context.Background(), preview, "")
	if err == nil {
		t.Error("EnsureApplicationSet() should return error for empty namespace")
	}
}

// TestGetApplicationSetName verifies the ApplicationSet name generation
func TestGetApplicationSetName(t *testing.T) {
	tests := []struct {
		prNumber int
		want     string
	}{
		{prNumber: 1, want: "preview-1"},
		{prNumber: 123, want: "preview-123"},
		{prNumber: 9999, want: "preview-9999"},
	}

	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := m.GetApplicationSetName(tt.prNumber)
			if got != tt.want {
				t.Errorf("GetApplicationSetName(%d) = %v, want %v", tt.prNumber, got, tt.want)
			}
		})
	}
}

// TestBuildApplicationSet_DestinationServer verifies the destination server is set
func TestBuildApplicationSet_DestinationServer(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-123-abc12345")

	// Default to in-cluster
	expectedServer := "https://kubernetes.default.svc"
	if appSet.Spec.Template.Spec.Destination.Server != expectedServer {
		t.Errorf("Destination server = %v, want %v",
			appSet.Spec.Template.Spec.Destination.Server, expectedServer)
	}
}

// TestBuildApplicationSet_ManagedByLabel verifies the managed-by label
func TestBuildApplicationSet_ManagedByLabel(t *testing.T) {
	c := setupTestClient(t)
	m := NewManager(c, c.Scheme(), "https://github.com/example/app", "argocd", "default")

	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "abc-123",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "example/app",
			HeadSHA:    "abc123def456789012345678901234567890abcd",
			Services:   []string{"auth"},
		},
	}

	appSet := m.BuildApplicationSet(preview, "preview-pr-123-abc12345")

	if appSet.Labels["preview.previewd.io/managed-by"] != "previewd" {
		t.Errorf("managed-by label = %v, want previewd",
			appSet.Labels["preview.previewd.io/managed-by"])
	}
}
