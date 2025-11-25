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
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TestApplicationSet_DeepCopy verifies DeepCopy returns a deep copy
func TestApplicationSet_DeepCopy(t *testing.T) {
	original := &ApplicationSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-appset",
			Namespace: "argocd",
		},
		Spec: ApplicationSetSpec{
			GoTemplate: true,
		},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil for non-nil ApplicationSet")
	}
	if copied == original {
		t.Error("DeepCopy() returned same pointer, expected deep copy")
	}
	if copied.Name != original.Name {
		t.Errorf("DeepCopy() name = %v, want %v", copied.Name, original.Name)
	}
}

// TestApplicationSet_DeepCopy_NilReceiver verifies DeepCopy handles nil receiver
func TestApplicationSet_DeepCopy_NilReceiver(t *testing.T) {
	var appSet *ApplicationSet
	copied := appSet.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestApplicationSet_DeepCopyObject verifies DeepCopyObject returns non-nil
func TestApplicationSet_DeepCopyObject(t *testing.T) {
	original := &ApplicationSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}

	copied := original.DeepCopyObject()

	if copied == nil {
		t.Fatal("DeepCopyObject() returned nil")
	}
}

// TestApplicationSet_DeepCopyObject_NilReceiver verifies nil handling
func TestApplicationSet_DeepCopyObject_NilReceiver(t *testing.T) {
	var appSet *ApplicationSet
	copied := appSet.DeepCopyObject()
	if copied != nil {
		t.Error("DeepCopyObject() on nil receiver should return nil")
	}
}

// TestApplicationSetList_DeepCopy verifies DeepCopy for list types
func TestApplicationSetList_DeepCopy(t *testing.T) {
	original := &ApplicationSetList{
		Items: []ApplicationSet{
			{ObjectMeta: metav1.ObjectMeta{Name: "item1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "item2"}},
		},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil")
	}
	if len(copied.Items) != len(original.Items) {
		t.Errorf("DeepCopy() items length = %v, want %v", len(copied.Items), len(original.Items))
	}
}

// TestApplicationSetList_DeepCopy_NilReceiver verifies nil handling
func TestApplicationSetList_DeepCopy_NilReceiver(t *testing.T) {
	var list *ApplicationSetList
	copied := list.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestApplicationSetList_DeepCopyObject verifies DeepCopyObject
func TestApplicationSetList_DeepCopyObject(t *testing.T) {
	original := &ApplicationSetList{}
	copied := original.DeepCopyObject()
	if copied == nil {
		t.Fatal("DeepCopyObject() returned nil")
	}
}

// TestApplicationSetList_DeepCopyObject_NilReceiver verifies nil handling
func TestApplicationSetList_DeepCopyObject_NilReceiver(t *testing.T) {
	var list *ApplicationSetList
	copied := list.DeepCopyObject()
	if copied != nil {
		t.Error("DeepCopyObject() on nil receiver should return nil")
	}
}

// TestApplicationSetSpec_DeepCopy verifies DeepCopy for spec
func TestApplicationSetSpec_DeepCopy(t *testing.T) {
	original := &ApplicationSetSpec{
		GoTemplate:        true,
		GoTemplateOptions: []string{"missingkey=error"},
		Generators: []ApplicationSetGenerator{
			{List: &ListGenerator{}},
		},
		SyncPolicy: &ApplicationSetSyncPolicy{PreserveResourcesOnDeletion: true},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil")
	}
	if copied.SyncPolicy == original.SyncPolicy {
		t.Error("DeepCopy() SyncPolicy should be a new pointer")
	}
}

// TestApplicationSetSpec_DeepCopy_NilReceiver verifies nil handling
func TestApplicationSetSpec_DeepCopy_NilReceiver(t *testing.T) {
	var spec *ApplicationSetSpec
	copied := spec.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestApplicationSetStatus_DeepCopy verifies DeepCopy for status
func TestApplicationSetStatus_DeepCopy(t *testing.T) {
	original := &ApplicationSetStatus{
		Conditions: []ApplicationSetCondition{
			{Type: "Ready", Status: "True"},
		},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil")
	}
	if len(copied.Conditions) != len(original.Conditions) {
		t.Errorf("DeepCopy() conditions length = %v, want %v", len(copied.Conditions), len(original.Conditions))
	}
}

// TestApplicationSetStatus_DeepCopy_NilReceiver verifies nil handling
func TestApplicationSetStatus_DeepCopy_NilReceiver(t *testing.T) {
	var status *ApplicationSetStatus
	copied := status.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestApplicationSetGenerator_DeepCopy verifies DeepCopy for generator
func TestApplicationSetGenerator_DeepCopy(t *testing.T) {
	original := &ApplicationSetGenerator{
		List: &ListGenerator{
			Elements: []apiextensionsv1.JSON{
				{Raw: []byte(`{"service":"api"}`)},
			},
		},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil")
	}
	if copied.List == original.List {
		t.Error("DeepCopy() List should be a new pointer")
	}
}

// TestApplicationSetGenerator_DeepCopy_NilReceiver verifies nil handling
func TestApplicationSetGenerator_DeepCopy_NilReceiver(t *testing.T) {
	var gen *ApplicationSetGenerator
	copied := gen.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestListGenerator_DeepCopy verifies DeepCopy for list generator
func TestListGenerator_DeepCopy(t *testing.T) {
	original := &ListGenerator{
		Elements: []apiextensionsv1.JSON{
			{Raw: []byte(`{"key":"value"}`)},
		},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil")
	}
}

// TestListGenerator_DeepCopy_NilReceiver verifies nil handling
func TestListGenerator_DeepCopy_NilReceiver(t *testing.T) {
	var gen *ListGenerator
	copied := gen.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestApplicationSetTemplate_DeepCopy verifies DeepCopy for template
func TestApplicationSetTemplate_DeepCopy(t *testing.T) {
	original := &ApplicationSetTemplate{
		ApplicationSetTemplateMeta: ApplicationSetTemplateMeta{
			Name: "test-{{service}}",
		},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil")
	}
}

// TestApplicationSetTemplate_DeepCopy_NilReceiver verifies nil handling
func TestApplicationSetTemplate_DeepCopy_NilReceiver(t *testing.T) {
	var template *ApplicationSetTemplate
	copied := template.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestApplicationSetTemplateMeta_DeepCopy verifies DeepCopy for template meta
func TestApplicationSetTemplateMeta_DeepCopy(t *testing.T) {
	original := &ApplicationSetTemplateMeta{
		Name:        "test",
		Namespace:   "argocd",
		Labels:      map[string]string{"app": "test"},
		Annotations: map[string]string{"note": "value"},
		Finalizers:  []string{"foregroundDeletion"},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil")
	}
	if copied.Labels == nil {
		t.Error("DeepCopy() Labels should not be nil")
	}
	// Verify maps are deep copied
	copied.Labels["new"] = "value"
	if _, exists := original.Labels["new"]; exists {
		t.Error("DeepCopy() Labels should be a new map")
	}
}

// TestApplicationSetTemplateMeta_DeepCopy_NilReceiver verifies nil handling
func TestApplicationSetTemplateMeta_DeepCopy_NilReceiver(t *testing.T) {
	var meta *ApplicationSetTemplateMeta
	copied := meta.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestApplication_DeepCopy verifies DeepCopy for Application
func TestApplication_DeepCopy(t *testing.T) {
	original := &Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "argocd",
		},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil")
	}
	if copied == original {
		t.Error("DeepCopy() should return new pointer")
	}
}

// TestApplication_DeepCopy_NilReceiver verifies nil handling
func TestApplication_DeepCopy_NilReceiver(t *testing.T) {
	var app *Application
	copied := app.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestApplication_DeepCopyObject verifies DeepCopyObject
func TestApplication_DeepCopyObject(t *testing.T) {
	original := &Application{}
	copied := original.DeepCopyObject()
	if copied == nil {
		t.Fatal("DeepCopyObject() returned nil")
	}
}

// TestApplication_DeepCopyObject_NilReceiver verifies nil handling
func TestApplication_DeepCopyObject_NilReceiver(t *testing.T) {
	var app *Application
	copied := app.DeepCopyObject()
	if copied != nil {
		t.Error("DeepCopyObject() on nil receiver should return nil")
	}
}

// TestApplicationList_DeepCopy verifies DeepCopy for list
func TestApplicationList_DeepCopy(t *testing.T) {
	original := &ApplicationList{
		Items: []Application{
			{ObjectMeta: metav1.ObjectMeta{Name: "app1"}},
		},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil")
	}
}

// TestApplicationList_DeepCopy_NilReceiver verifies nil handling
func TestApplicationList_DeepCopy_NilReceiver(t *testing.T) {
	var list *ApplicationList
	copied := list.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestApplicationList_DeepCopyObject verifies DeepCopyObject
func TestApplicationList_DeepCopyObject(t *testing.T) {
	original := &ApplicationList{}
	copied := original.DeepCopyObject()
	if copied == nil {
		t.Fatal("DeepCopyObject() returned nil")
	}
}

// TestApplicationList_DeepCopyObject_NilReceiver verifies nil handling
func TestApplicationList_DeepCopyObject_NilReceiver(t *testing.T) {
	var list *ApplicationList
	copied := list.DeepCopyObject()
	if copied != nil {
		t.Error("DeepCopyObject() on nil receiver should return nil")
	}
}

// TestApplicationSpec_DeepCopy verifies DeepCopy for spec
func TestApplicationSpec_DeepCopy(t *testing.T) {
	original := &ApplicationSpec{
		Project: "default",
		Source: &ApplicationSource{
			RepoURL: "https://github.com/example/app",
		},
		SyncPolicy: &SyncPolicy{
			Automated: &SyncPolicyAutomated{Prune: true},
		},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil")
	}
	if copied.Source == original.Source {
		t.Error("DeepCopy() Source should be new pointer")
	}
}

// TestApplicationSpec_DeepCopy_NilReceiver verifies nil handling
func TestApplicationSpec_DeepCopy_NilReceiver(t *testing.T) {
	var spec *ApplicationSpec
	copied := spec.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestApplicationStatus_DeepCopy verifies DeepCopy for status
func TestApplicationStatus_DeepCopy(t *testing.T) {
	original := &ApplicationStatus{
		Health: HealthStatus{Status: "Healthy"},
		Sync:   SyncStatus{Status: "Synced"},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil")
	}
}

// TestApplicationStatus_DeepCopy_NilReceiver verifies nil handling
func TestApplicationStatus_DeepCopy_NilReceiver(t *testing.T) {
	var status *ApplicationStatus
	copied := status.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestApplicationSource_DeepCopy verifies DeepCopy for source
func TestApplicationSource_DeepCopy(t *testing.T) {
	original := &ApplicationSource{
		RepoURL: "https://github.com/example/app",
		Kustomize: &ApplicationSourceKustomize{
			NamePrefix: "pr-123-",
		},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil")
	}
	if copied.Kustomize == original.Kustomize {
		t.Error("DeepCopy() Kustomize should be new pointer")
	}
}

// TestApplicationSource_DeepCopy_NilReceiver verifies nil handling
func TestApplicationSource_DeepCopy_NilReceiver(t *testing.T) {
	var source *ApplicationSource
	copied := source.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestApplicationSourceKustomize_DeepCopy verifies DeepCopy for Kustomize
func TestApplicationSourceKustomize_DeepCopy(t *testing.T) {
	original := &ApplicationSourceKustomize{
		NamePrefix:        "pr-123-",
		Namespace:         "preview",
		CommonLabels:      map[string]string{"pr": "123"},
		CommonAnnotations: map[string]string{"note": "value"},
		Images:            []string{"nginx:1.19"},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil")
	}
	// Verify maps are deep copied
	copied.CommonLabels["new"] = "value"
	if _, exists := original.CommonLabels["new"]; exists {
		t.Error("DeepCopy() CommonLabels should be a new map")
	}
}

// TestApplicationSourceKustomize_DeepCopy_NilReceiver verifies nil handling
func TestApplicationSourceKustomize_DeepCopy_NilReceiver(t *testing.T) {
	var kustomize *ApplicationSourceKustomize
	copied := kustomize.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestSyncPolicy_DeepCopy verifies DeepCopy for sync policy
func TestSyncPolicy_DeepCopy(t *testing.T) {
	factor := int64(2)
	original := &SyncPolicy{
		Automated:   &SyncPolicyAutomated{Prune: true, SelfHeal: true},
		SyncOptions: []string{"CreateNamespace=true"},
		Retry: &RetryStrategy{
			Limit: 5,
			Backoff: &Backoff{
				Duration:    "5s",
				Factor:      &factor,
				MaxDuration: "3m",
			},
		},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil")
	}
	if copied.Automated == original.Automated {
		t.Error("DeepCopy() Automated should be new pointer")
	}
	if copied.Retry == original.Retry {
		t.Error("DeepCopy() Retry should be new pointer")
	}
}

// TestSyncPolicy_DeepCopy_NilReceiver verifies nil handling
func TestSyncPolicy_DeepCopy_NilReceiver(t *testing.T) {
	var policy *SyncPolicy
	copied := policy.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestRetryStrategy_DeepCopy verifies DeepCopy for retry strategy
func TestRetryStrategy_DeepCopy(t *testing.T) {
	factor := int64(2)
	original := &RetryStrategy{
		Limit: 5,
		Backoff: &Backoff{
			Duration:    "5s",
			Factor:      &factor,
			MaxDuration: "3m",
		},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Fatal("DeepCopy() returned nil")
	}
	if copied.Backoff == original.Backoff {
		t.Error("DeepCopy() Backoff should be new pointer")
	}
}

// TestRetryStrategy_DeepCopy_NilReceiver verifies nil handling
func TestRetryStrategy_DeepCopy_NilReceiver(t *testing.T) {
	var strategy *RetryStrategy
	copied := strategy.DeepCopy()
	if copied != nil {
		t.Error("DeepCopy() on nil receiver should return nil")
	}
}

// TestAddKnownTypes verifies schema registration
func TestAddKnownTypes(t *testing.T) {
	scheme := runtime.NewScheme()

	err := AddToScheme(scheme)

	if err != nil {
		t.Fatalf("AddToScheme() error = %v", err)
	}

	// Verify types are registered
	gvk := GroupVersion.WithKind("ApplicationSet")
	if !scheme.Recognizes(gvk) {
		t.Error("Scheme should recognize ApplicationSet")
	}

	gvk = GroupVersion.WithKind("Application")
	if !scheme.Recognizes(gvk) {
		t.Error("Scheme should recognize Application")
	}
}
