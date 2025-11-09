/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PreviewEnvironmentSpec defines the desired state of PreviewEnvironment
type PreviewEnvironmentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// Repository is the GitHub repository in "owner/repo" format
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9-]+/[a-zA-Z0-9-]+$"
	Repository string `json:"repository"`

	// PRNumber is the pull request number
	// +kubebuilder:validation:Minimum=1
	PRNumber int `json:"prNumber"`

	// HeadSHA is the commit SHA of the PR head (40 character hex string)
	// +kubebuilder:validation:Pattern="^[a-f0-9]{40}$"
	HeadSHA string `json:"headSHA"`

	// BaseBranch is the base branch name (optional)
	// +optional
	BaseBranch string `json:"baseBranch,omitempty"`

	// HeadBranch is the head branch name (optional)
	// +optional
	HeadBranch string `json:"headBranch,omitempty"`

	// Services is a list of service names to deploy (optional)
	// +optional
	Services []string `json:"services,omitempty"`

	// TTL is the time-to-live duration for the preview environment
	// +kubebuilder:default="4h"
	// +optional
	TTL string `json:"ttl,omitempty"`
}

// PreviewEnvironmentStatus defines the observed state of PreviewEnvironment.
type PreviewEnvironmentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the PreviewEnvironment resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Phase represents the current phase of the preview environment
	// Valid values: Pending, Creating, Ready, Updating, Deleting, Failed
	// +kubebuilder:validation:Enum=Pending;Creating;Ready;Updating;Deleting;Failed
	// +optional
	Phase string `json:"phase,omitempty"`

	// URL is the public URL to access the preview environment
	// +optional
	URL string `json:"url,omitempty"`

	// Namespace is the Kubernetes namespace created for this preview environment
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Services contains status information for deployed services
	// +optional
	Services []ServiceStatus `json:"services,omitempty"`

	// CostEstimate provides estimated costs for running this environment
	// +optional
	CostEstimate *CostEstimate `json:"costEstimate,omitempty"`

	// CreatedAt is the timestamp when the environment was created
	// +optional
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// ExpiresAt is the timestamp when the environment will be automatically deleted
	// +optional
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`

	// LastSyncedAt is the timestamp of the last successful sync
	// +optional
	LastSyncedAt *metav1.Time `json:"lastSyncedAt,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed spec
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// ServiceStatus represents the status of a deployed service
type ServiceStatus struct {
	// Name is the service name
	Name string `json:"name"`

	// Ready indicates if the service is ready
	Ready bool `json:"ready"`

	// URL is the service URL (if exposed)
	// +optional
	URL string `json:"url,omitempty"`
}

// CostEstimate provides cost estimation for the preview environment
type CostEstimate struct {
	// Currency is the cost currency (e.g., USD)
	Currency string `json:"currency"`

	// HourlyCost is the estimated hourly cost
	HourlyCost string `json:"hourlyCost"`

	// TotalCost is the total estimated cost based on TTL
	// +optional
	TotalCost string `json:"totalCost,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="PR",type="integer",JSONPath=".spec.prNumber",description="Pull Request Number"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Current Phase"
// +kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.url",description="Preview URL"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Creation Time"
// +kubebuilder:resource:shortName=preview;previews

// PreviewEnvironment is the Schema for the previewenvironments API
type PreviewEnvironment struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of PreviewEnvironment
	// +required
	Spec PreviewEnvironmentSpec `json:"spec"`

	// status defines the observed state of PreviewEnvironment
	// +optional
	Status PreviewEnvironmentStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// PreviewEnvironmentList contains a list of PreviewEnvironment
type PreviewEnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PreviewEnvironment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PreviewEnvironment{}, &PreviewEnvironmentList{})
}
