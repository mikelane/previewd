/*
Copyright (c) 2025 Mike Lane

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PreviewEnvironmentSpec defines the desired state of PreviewEnvironment
type PreviewEnvironmentSpec struct {
	IngressPort   *int32             `json:"ingressPort,omitempty"`
	ResourceQuota *ResourceQuotaSpec `json:"resourceQuota,omitempty"`
	Repository    string             `json:"repository"`
	HeadSHA       string             `json:"headSHA"`
	BaseBranch    string             `json:"baseBranch,omitempty"`
	HeadBranch    string             `json:"headBranch,omitempty"`
	TTL           string             `json:"ttl,omitempty"`
	Services      []string           `json:"services,omitempty"`
	PRNumber      int                `json:"prNumber"`
}

// ResourceQuotaSpec defines resource quota limits for a preview environment
type ResourceQuotaSpec struct {
	// RequestsCPU is the CPU requests limit (default: "2")
	// +optional
	RequestsCPU string `json:"requestsCpu,omitempty"`

	// LimitsCPU is the CPU limits (default: "4")
	// +optional
	LimitsCPU string `json:"limitsCpu,omitempty"`

	// RequestsMemory is the memory requests limit (default: "4Gi")
	// +optional
	RequestsMemory string `json:"requestsMemory,omitempty"`

	// LimitsMemory is the memory limits (default: "8Gi")
	// +optional
	LimitsMemory string `json:"limitsMemory,omitempty"`
}

// PreviewEnvironmentStatus defines the observed state of PreviewEnvironment.
type PreviewEnvironmentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

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

	// Services contains status information for deployed services
	// +optional
	Services []ServiceStatus `json:"services,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed spec
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// ServiceStatus represents the status of a deployed service
type ServiceStatus struct {
	// Name is the service name
	Name string `json:"name"`

	// URL is the service URL (if exposed)
	// +optional
	URL string `json:"url,omitempty"`

	// Ready indicates if the service is ready
	Ready bool `json:"ready"`
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
	Spec              PreviewEnvironmentSpec `json:"spec"`
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`
	Status            PreviewEnvironmentStatus `json:"status,omitempty,omitzero"`
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
