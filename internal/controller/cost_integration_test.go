/*
MIT License

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

package controller

import (
	"context"
	"testing"
	"time"

	previewv1alpha1 "github.com/mikelane/previewd/api/v1alpha1"
	"github.com/mikelane/previewd/internal/cost"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	testScheme = runtime.NewScheme()
)

func init() {
	// Register our types with the scheme
	if err := clientgoscheme.AddToScheme(testScheme); err != nil {
		panic(err)
	}
	if err := previewv1alpha1.AddToScheme(testScheme); err != nil {
		panic(err)
	}
}

func TestReconciler_UpdatesCostEstimate(t *testing.T) {
	tests := []struct {
		pods     []corev1.Pod
		name     string
		wantCost *previewv1alpha1.CostEstimate
		preview  *previewv1alpha1.PreviewEnvironment
	}{
		{
			preview: &previewv1alpha1.PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-preview",
					Namespace: "default",
				},
				Spec: previewv1alpha1.PreviewEnvironmentSpec{
					Repository: "org/repo",
					PRNumber:   123,
					HeadSHA:    "1234567890123456789012345678901234567890",
					TTL:        "4h",
				},
				Status: previewv1alpha1.PreviewEnvironmentStatus{
					Phase:     "Ready",
					Namespace: "preview-pr-123",
				},
			},
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "app-pod",
						Namespace: "preview-pr-123",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "app",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("1"),
										corev1.ResourceMemory: resource.MustParse("2Gi"),
									},
								},
							},
						},
					},
				},
			},
			name: "updates cost estimate for preview environment",
			wantCost: &previewv1alpha1.CostEstimate{
				Currency:   "USD",
				HourlyCost: "0.0500",
				TotalCost:  "0.2000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with preview environment and pods
			objs := []client.Object{tt.preview}
			for i := range tt.pods {
				objs = append(objs, &tt.pods[i])
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects(objs...).
				WithStatusSubresource(tt.preview).
				Build()

			reconciler := &PreviewEnvironmentReconciler{
				Client:        fakeClient,
				Scheme:        testScheme,
				CostEstimator: cost.NewEstimator(nil),
			}

			// Reconcile
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      tt.preview.Name,
					Namespace: tt.preview.Namespace,
				},
			}

			_, err := reconciler.Reconcile(context.TODO(), req)
			if err != nil {
				t.Fatalf("Reconcile() error = %v", err)
			}

			// Check that cost estimate was updated
			var updated previewv1alpha1.PreviewEnvironment
			err = fakeClient.Get(context.TODO(), req.NamespacedName, &updated)
			if err != nil {
				t.Fatalf("Failed to get updated preview environment: %v", err)
			}

			if updated.Status.CostEstimate == nil {
				t.Fatal("Cost estimate was not set")
			}

			if updated.Status.CostEstimate.Currency != tt.wantCost.Currency {
				t.Errorf("Currency = %v, want %v", updated.Status.CostEstimate.Currency, tt.wantCost.Currency)
			}
			if updated.Status.CostEstimate.HourlyCost != tt.wantCost.HourlyCost {
				t.Errorf("HourlyCost = %v, want %v", updated.Status.CostEstimate.HourlyCost, tt.wantCost.HourlyCost)
			}
			if updated.Status.CostEstimate.TotalCost != tt.wantCost.TotalCost {
				t.Errorf("TotalCost = %v, want %v", updated.Status.CostEstimate.TotalCost, tt.wantCost.TotalCost)
			}
		})
	}
}

func TestParseTTL(t *testing.T) {
	tests := []struct {
		ttl     string
		name    string
		want    time.Duration
		wantErr bool
	}{
		{
			ttl:     "4h",
			name:    "parses hours",
			want:    4 * time.Hour,
			wantErr: false,
		},
		{
			ttl:     "30m",
			name:    "parses minutes",
			want:    30 * time.Minute,
			wantErr: false,
		},
		{
			ttl:     "2d",
			name:    "parses days",
			want:    48 * time.Hour,
			wantErr: false,
		},
		{
			ttl:     "1h30m",
			name:    "parses complex duration",
			want:    90 * time.Minute,
			wantErr: false,
		},
		{
			ttl:     "",
			name:    "handles empty string",
			want:    4 * time.Hour, // default
			wantErr: false,
		},
		{
			ttl:     "invalid",
			name:    "handles invalid format",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTTL(tt.ttl)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTTL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseTTL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckSpotInstance(t *testing.T) {
	tests := []struct {
		name     string
		preview  *previewv1alpha1.PreviewEnvironment
		wantSpot bool
	}{
		{
			name: "detects spot instance annotation",
			preview: &previewv1alpha1.PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"previewd.io/use-spot": "true",
					},
				},
			},
			wantSpot: true,
		},
		{
			name: "defaults to on-demand",
			preview: &previewv1alpha1.PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{},
			},
			wantSpot: false,
		},
		{
			name: "handles false annotation",
			preview: &previewv1alpha1.PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"previewd.io/use-spot": "false",
					},
				},
			},
			wantSpot: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkSpotInstance(tt.preview)
			if got != tt.wantSpot {
				t.Errorf("checkSpotInstance() = %v, want %v", got, tt.wantSpot)
			}
		})
	}
}
