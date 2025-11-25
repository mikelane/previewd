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

// Struct field order must match ArgoCD API for JSON serialization compatibility.
//
//nolint:govet // fieldalignment warnings ignored - field order matches ArgoCD CRD API
package argocd

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupVersion is group version used to register these objects
var GroupVersion = schema.GroupVersion{Group: "argoproj.io", Version: "v1alpha1"}

// SchemeBuilder is used to add go types to the GroupVersionKind scheme
var SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

// AddToScheme adds the types in this group-version to the given scheme.
var AddToScheme = SchemeBuilder.AddToScheme

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&ApplicationSet{},
		&ApplicationSetList{},
		&Application{},
		&ApplicationList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}

// ApplicationSet is a set of Application resources
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type ApplicationSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`
	Spec              ApplicationSetSpec   `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	Status            ApplicationSetStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// DeepCopyObject returns a deep copy of the ApplicationSet
func (in *ApplicationSet) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(ApplicationSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ApplicationSet) DeepCopyInto(out *ApplicationSet) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy returns a deep copy of the ApplicationSet
func (in *ApplicationSet) DeepCopy() *ApplicationSet {
	if in == nil {
		return nil
	}
	out := new(ApplicationSet)
	in.DeepCopyInto(out)
	return out
}

// ApplicationSetList is a list of ApplicationSet resources
// +kubebuilder:object:root=true
type ApplicationSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationSet `json:"items"`
}

// DeepCopyObject returns a deep copy of the ApplicationSetList
func (in *ApplicationSetList) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(ApplicationSetList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ApplicationSetList) DeepCopyInto(out *ApplicationSetList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ApplicationSet, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy returns a deep copy of the ApplicationSetList
func (in *ApplicationSetList) DeepCopy() *ApplicationSetList {
	if in == nil {
		return nil
	}
	out := new(ApplicationSetList)
	in.DeepCopyInto(out)
	return out
}

// ApplicationSetSpec represents a class of application set state
type ApplicationSetSpec struct {
	// GoTemplate enables Go templating in the ApplicationSet
	GoTemplate bool `json:"goTemplate,omitempty"`
	// GoTemplateOptions specifies options for Go templating
	GoTemplateOptions []string `json:"goTemplateOptions,omitempty"`
	// Generators is a list of generators to generate Applications
	Generators []ApplicationSetGenerator `json:"generators"`
	// Template is the template for generating Applications
	Template ApplicationSetTemplate `json:"template"`
	// SyncPolicy controls the sync behavior of the ApplicationSet
	SyncPolicy *ApplicationSetSyncPolicy `json:"syncPolicy,omitempty"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ApplicationSetSpec) DeepCopyInto(out *ApplicationSetSpec) {
	*out = *in
	if in.GoTemplateOptions != nil {
		in, out := &in.GoTemplateOptions, &out.GoTemplateOptions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Generators != nil {
		in, out := &in.Generators, &out.Generators
		*out = make([]ApplicationSetGenerator, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Template.DeepCopyInto(&out.Template)
	if in.SyncPolicy != nil {
		in, out := &in.SyncPolicy, &out.SyncPolicy
		*out = new(ApplicationSetSyncPolicy)
		**out = **in
	}
}

// DeepCopy returns a deep copy of the ApplicationSetSpec
func (in *ApplicationSetSpec) DeepCopy() *ApplicationSetSpec {
	if in == nil {
		return nil
	}
	out := new(ApplicationSetSpec)
	in.DeepCopyInto(out)
	return out
}

// ApplicationSetStatus contains status information for the ApplicationSet
type ApplicationSetStatus struct {
	// Conditions contains the conditions of the ApplicationSet
	Conditions []ApplicationSetCondition `json:"conditions,omitempty"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ApplicationSetStatus) DeepCopyInto(out *ApplicationSetStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ApplicationSetCondition, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy returns a deep copy of the ApplicationSetStatus
func (in *ApplicationSetStatus) DeepCopy() *ApplicationSetStatus {
	if in == nil {
		return nil
	}
	out := new(ApplicationSetStatus)
	in.DeepCopyInto(out)
	return out
}

// ApplicationSetCondition contains details about an ApplicationSet condition
type ApplicationSetCondition struct {
	// Type is the type of the condition
	Type string `json:"type"`
	// Status is the status of the condition
	Status string `json:"status"`
	// Message contains human-readable message indicating details about the condition
	Message string `json:"message,omitempty"`
	// Reason is a brief machine-readable explanation for the condition's state
	Reason string `json:"reason,omitempty"`
}

// ApplicationSetGenerator is a generator to generate Applications
type ApplicationSetGenerator struct {
	// List generator
	List *ListGenerator `json:"list,omitempty"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ApplicationSetGenerator) DeepCopyInto(out *ApplicationSetGenerator) {
	*out = *in
	if in.List != nil {
		in, out := &in.List, &out.List
		*out = new(ListGenerator)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy returns a deep copy of the ApplicationSetGenerator
func (in *ApplicationSetGenerator) DeepCopy() *ApplicationSetGenerator {
	if in == nil {
		return nil
	}
	out := new(ApplicationSetGenerator)
	in.DeepCopyInto(out)
	return out
}

// ListGenerator generates Applications from a list of elements
type ListGenerator struct {
	// Elements is a list of element objects
	Elements []apiextensionsv1.JSON `json:"elements"`
	// Template is an optional template to override the default template
	Template ApplicationSetTemplate `json:"template,omitempty"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ListGenerator) DeepCopyInto(out *ListGenerator) {
	*out = *in
	if in.Elements != nil {
		in, out := &in.Elements, &out.Elements
		*out = make([]apiextensionsv1.JSON, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Template.DeepCopyInto(&out.Template)
}

// DeepCopy returns a deep copy of the ListGenerator
func (in *ListGenerator) DeepCopy() *ListGenerator {
	if in == nil {
		return nil
	}
	out := new(ListGenerator)
	in.DeepCopyInto(out)
	return out
}

// ApplicationSetTemplate represents the template for generated Applications
type ApplicationSetTemplate struct {
	// ApplicationSetTemplateMeta is the metadata for the Application template
	ApplicationSetTemplateMeta `json:"metadata"`
	// Spec is the Application spec template
	Spec ApplicationSpec `json:"spec"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ApplicationSetTemplate) DeepCopyInto(out *ApplicationSetTemplate) {
	*out = *in
	in.ApplicationSetTemplateMeta.DeepCopyInto(&out.ApplicationSetTemplateMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy returns a deep copy of the ApplicationSetTemplate
func (in *ApplicationSetTemplate) DeepCopy() *ApplicationSetTemplate {
	if in == nil {
		return nil
	}
	out := new(ApplicationSetTemplate)
	in.DeepCopyInto(out)
	return out
}

// ApplicationSetTemplateMeta is the metadata template for Applications
type ApplicationSetTemplateMeta struct {
	// Name is the template for the Application name
	Name string `json:"name,omitempty"`
	// Namespace is the namespace for the Application
	Namespace string `json:"namespace,omitempty"`
	// Labels are the labels for the Application
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations are the annotations for the Application
	Annotations map[string]string `json:"annotations,omitempty"`
	// Finalizers are the finalizers for the Application
	Finalizers []string `json:"finalizers,omitempty"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ApplicationSetTemplateMeta) DeepCopyInto(out *ApplicationSetTemplateMeta) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Finalizers != nil {
		in, out := &in.Finalizers, &out.Finalizers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy returns a deep copy of the ApplicationSetTemplateMeta
func (in *ApplicationSetTemplateMeta) DeepCopy() *ApplicationSetTemplateMeta {
	if in == nil {
		return nil
	}
	out := new(ApplicationSetTemplateMeta)
	in.DeepCopyInto(out)
	return out
}

// ApplicationSetSyncPolicy controls the sync behavior of the ApplicationSet
type ApplicationSetSyncPolicy struct {
	// PreserveResourcesOnDeletion preserves the resources when the ApplicationSet is deleted
	PreserveResourcesOnDeletion bool `json:"preserveResourcesOnDeletion,omitempty"`
}

// Application is a definition of Application resource
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ApplicationSpec   `json:"spec"`
	Status            ApplicationStatus `json:"status,omitempty"`
}

// DeepCopyObject returns a deep copy of the Application
func (in *Application) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(Application)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *Application) DeepCopyInto(out *Application) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy returns a deep copy of the Application
func (in *Application) DeepCopy() *Application {
	if in == nil {
		return nil
	}
	out := new(Application)
	in.DeepCopyInto(out)
	return out
}

// ApplicationList is a list of Application resources
// +kubebuilder:object:root=true
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

// DeepCopyObject returns a deep copy of the ApplicationList
func (in *ApplicationList) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(ApplicationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ApplicationList) DeepCopyInto(out *ApplicationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Application, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy returns a deep copy of the ApplicationList
func (in *ApplicationList) DeepCopy() *ApplicationList {
	if in == nil {
		return nil
	}
	out := new(ApplicationList)
	in.DeepCopyInto(out)
	return out
}

// ApplicationSpec represents desired application state
type ApplicationSpec struct {
	// Source is a reference to the source of the application manifests
	Source *ApplicationSource `json:"source,omitempty"`
	// Destination is a reference to the target cluster and namespace
	Destination ApplicationDestination `json:"destination"`
	// Project is a reference to the project this application belongs to
	Project string `json:"project"`
	// SyncPolicy controls when and how a sync will be performed
	SyncPolicy *SyncPolicy `json:"syncPolicy,omitempty"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ApplicationSpec) DeepCopyInto(out *ApplicationSpec) {
	*out = *in
	if in.Source != nil {
		in, out := &in.Source, &out.Source
		*out = new(ApplicationSource)
		(*in).DeepCopyInto(*out)
	}
	out.Destination = in.Destination
	if in.SyncPolicy != nil {
		in, out := &in.SyncPolicy, &out.SyncPolicy
		*out = new(SyncPolicy)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy returns a deep copy of the ApplicationSpec
func (in *ApplicationSpec) DeepCopy() *ApplicationSpec {
	if in == nil {
		return nil
	}
	out := new(ApplicationSpec)
	in.DeepCopyInto(out)
	return out
}

// ApplicationStatus contains status information for the application
type ApplicationStatus struct {
	// Health contains information about the health status
	Health HealthStatus `json:"health,omitempty"`
	// Sync contains information about the sync status
	Sync SyncStatus `json:"sync,omitempty"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ApplicationStatus) DeepCopyInto(out *ApplicationStatus) {
	*out = *in
	out.Health = in.Health
	out.Sync = in.Sync
}

// DeepCopy returns a deep copy of the ApplicationStatus
func (in *ApplicationStatus) DeepCopy() *ApplicationStatus {
	if in == nil {
		return nil
	}
	out := new(ApplicationStatus)
	in.DeepCopyInto(out)
	return out
}

// HealthStatus contains information about the health state
type HealthStatus struct {
	// Status holds the status code
	Status string `json:"status,omitempty"`
	// Message is a human-readable informational message
	Message string `json:"message,omitempty"`
}

// SyncStatus contains information about the sync state
type SyncStatus struct {
	// Status is the sync state
	Status string `json:"status"`
	// Revision contains the revision the sync was performed against
	Revision string `json:"revision,omitempty"`
}

// ApplicationSource contains information about the source of the application manifests
type ApplicationSource struct {
	// RepoURL is the URL to the repository
	RepoURL string `json:"repoURL"`
	// Path is the directory path within the repository
	Path string `json:"path,omitempty"`
	// TargetRevision is the revision to sync to
	TargetRevision string `json:"targetRevision,omitempty"`
	// Kustomize holds kustomize specific options
	Kustomize *ApplicationSourceKustomize `json:"kustomize,omitempty"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ApplicationSource) DeepCopyInto(out *ApplicationSource) {
	*out = *in
	if in.Kustomize != nil {
		in, out := &in.Kustomize, &out.Kustomize
		*out = new(ApplicationSourceKustomize)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy returns a deep copy of the ApplicationSource
func (in *ApplicationSource) DeepCopy() *ApplicationSource {
	if in == nil {
		return nil
	}
	out := new(ApplicationSource)
	in.DeepCopyInto(out)
	return out
}

// ApplicationSourceKustomize holds options specific to Kustomize
type ApplicationSourceKustomize struct {
	// NamePrefix is a prefix appended to resources
	NamePrefix string `json:"namePrefix,omitempty"`
	// NameSuffix is a suffix appended to resources
	NameSuffix string `json:"nameSuffix,omitempty"`
	// Namespace sets the namespace that Kustomize adds to all resources
	Namespace string `json:"namespace,omitempty"`
	// CommonLabels is a list of additional labels to add to rendered manifests
	CommonLabels map[string]string `json:"commonLabels,omitempty"`
	// CommonAnnotations is a list of additional annotations to add to rendered manifests
	CommonAnnotations map[string]string `json:"commonAnnotations,omitempty"`
	// Images is a list of Kustomize image overrides
	Images []string `json:"images,omitempty"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ApplicationSourceKustomize) DeepCopyInto(out *ApplicationSourceKustomize) {
	*out = *in
	if in.CommonLabels != nil {
		in, out := &in.CommonLabels, &out.CommonLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.CommonAnnotations != nil {
		in, out := &in.CommonAnnotations, &out.CommonAnnotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Images != nil {
		in, out := &in.Images, &out.Images
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy returns a deep copy of the ApplicationSourceKustomize
func (in *ApplicationSourceKustomize) DeepCopy() *ApplicationSourceKustomize {
	if in == nil {
		return nil
	}
	out := new(ApplicationSourceKustomize)
	in.DeepCopyInto(out)
	return out
}

// ApplicationDestination contains information about the target cluster and namespace
type ApplicationDestination struct {
	// Server is the Kubernetes API server URL
	Server string `json:"server,omitempty"`
	// Namespace is the target namespace
	Namespace string `json:"namespace,omitempty"`
	// Name is the cluster name
	Name string `json:"name,omitempty"`
}

// SyncPolicy controls when a sync will be performed
type SyncPolicy struct {
	// Automated will keep an application synced to the target revision
	Automated *SyncPolicyAutomated `json:"automated,omitempty"`
	// SyncOptions allow you to specify whole app sync-options
	SyncOptions []string `json:"syncOptions,omitempty"`
	// Retry controls failed sync retry behavior
	Retry *RetryStrategy `json:"retry,omitempty"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *SyncPolicy) DeepCopyInto(out *SyncPolicy) {
	*out = *in
	if in.Automated != nil {
		in, out := &in.Automated, &out.Automated
		*out = new(SyncPolicyAutomated)
		**out = **in
	}
	if in.SyncOptions != nil {
		in, out := &in.SyncOptions, &out.SyncOptions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Retry != nil {
		in, out := &in.Retry, &out.Retry
		*out = new(RetryStrategy)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy returns a deep copy of the SyncPolicy
func (in *SyncPolicy) DeepCopy() *SyncPolicy {
	if in == nil {
		return nil
	}
	out := new(SyncPolicy)
	in.DeepCopyInto(out)
	return out
}

// SyncPolicyAutomated controls the behavior of an automated sync
type SyncPolicyAutomated struct {
	// Prune specifies whether to delete resources from the cluster that are not found in the sources anymore
	Prune bool `json:"prune,omitempty"`
	// SelfHeal specifies whether to revert resources back to their desired state upon modification
	SelfHeal bool `json:"selfHeal,omitempty"`
	// AllowEmpty allows apps have zero live resources
	AllowEmpty bool `json:"allowEmpty,omitempty"`
}

// RetryStrategy controls the retry behavior
type RetryStrategy struct {
	// Limit is the maximum number of attempts for retrying a failed sync
	Limit int64 `json:"limit,omitempty"`
	// Backoff controls how to backoff on subsequent retries of failed syncs
	Backoff *Backoff `json:"backoff,omitempty"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *RetryStrategy) DeepCopyInto(out *RetryStrategy) {
	*out = *in
	if in.Backoff != nil {
		in, out := &in.Backoff, &out.Backoff
		*out = new(Backoff)
		**out = **in
	}
}

// DeepCopy returns a deep copy of the RetryStrategy
func (in *RetryStrategy) DeepCopy() *RetryStrategy {
	if in == nil {
		return nil
	}
	out := new(RetryStrategy)
	in.DeepCopyInto(out)
	return out
}

// Backoff specifies backoff parameters
type Backoff struct {
	// Duration is the amount to back off
	Duration string `json:"duration,omitempty"`
	// Factor is the factor to back off by
	Factor *int64 `json:"factor,omitempty"`
	// MaxDuration is the maximum amount of time allowed for the backoff
	MaxDuration string `json:"maxDuration,omitempty"`
}
