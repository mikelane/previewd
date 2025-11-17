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

package cost

import (
	"testing"
	"time"

	"github.com/mikelane/previewd/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewEstimator(t *testing.T) {
	tests := []struct {
		want   *Config
		config *Config
		name   string
	}{
		{
			name: "creates estimator with default config",
			want: &Config{
				CPUCostPerHour:    0.04,
				MemoryCostPerHour: 0.005,
				SpotDiscount:      0.30,
				Currency:          "USD",
			},
			config: nil,
		},
		{
			name: "creates estimator with custom config",
			want: &Config{
				CPUCostPerHour:    0.08,
				MemoryCostPerHour: 0.01,
				SpotDiscount:      0.40,
				Currency:          "EUR",
			},
			config: &Config{
				CPUCostPerHour:    0.08,
				MemoryCostPerHour: 0.01,
				SpotDiscount:      0.40,
				Currency:          "EUR",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimator := NewEstimator(tt.config)
			if estimator == nil {
				t.Fatal("NewEstimator returned nil")
			}
			if estimator.config.CPUCostPerHour != tt.want.CPUCostPerHour {
				t.Errorf("CPUCostPerHour = %v, want %v", estimator.config.CPUCostPerHour, tt.want.CPUCostPerHour)
			}
			if estimator.config.MemoryCostPerHour != tt.want.MemoryCostPerHour {
				t.Errorf("MemoryCostPerHour = %v, want %v", estimator.config.MemoryCostPerHour, tt.want.MemoryCostPerHour)
			}
			if estimator.config.SpotDiscount != tt.want.SpotDiscount {
				t.Errorf("SpotDiscount = %v, want %v", estimator.config.SpotDiscount, tt.want.SpotDiscount)
			}
			if estimator.config.Currency != tt.want.Currency {
				t.Errorf("Currency = %v, want %v", estimator.config.Currency, tt.want.Currency)
			}
		})
	}
}

func TestCalculatePodCost(t *testing.T) {
	tests := []struct {
		pod      *corev1.Pod
		name     string
		want     float64
		duration time.Duration
		useSpot  bool
	}{
		{
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "app",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"), // 0.5 CPU
									corev1.ResourceMemory: resource.MustParse("1Gi"),  // 1 GB
								},
							},
						},
					},
				},
			},
			name:     "calculates cost for pod with CPU and memory",
			want:     0.025, // (0.5 * 0.04) + (1 * 0.005) = 0.020 + 0.005 = 0.025
			duration: 1 * time.Hour,
			useSpot:  false,
		},
		{
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "app",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),   // 1 CPU
									corev1.ResourceMemory: resource.MustParse("2Gi"), // 2 GB
								},
							},
						},
						{
							Name: "sidecar",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),  // 0.2 CPU
									corev1.ResourceMemory: resource.MustParse("512Mi"), // 0.5 GB
								},
							},
						},
					},
				},
			},
			name:     "calculates cost for pod with multiple containers",
			want:     0.0605, // (1.2 * 0.04) + (2.5 * 0.005) = 0.048 + 0.0125 = 0.0605
			duration: 1 * time.Hour,
			useSpot:  false,
		},
		{
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "app",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
			},
			name:     "applies spot discount",
			want:     0.0315, // ((1 * 0.04) + (1 * 0.005)) * 0.7 = 0.045 * 0.7 = 0.0315
			duration: 1 * time.Hour,
			useSpot:  true,
		},
		{
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "app",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("2"),
									corev1.ResourceMemory: resource.MustParse("4Gi"),
								},
							},
						},
					},
				},
			},
			name:     "calculates cost for 24 hours",
			want:     2.40, // ((2 * 0.04) + (4 * 0.005)) * 24 = 0.1 * 24 = 2.40
			duration: 24 * time.Hour,
			useSpot:  false,
		},
		{
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      "app",
							Resources: corev1.ResourceRequirements{},
						},
					},
				},
			},
			name:     "handles pod with no resource requests",
			want:     0.0,
			duration: 1 * time.Hour,
			useSpot:  false,
		},
		{
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "app",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
			},
			name:     "handles fractional hours",
			want:     0.0675,           // ((1 * 0.04) + (1 * 0.005)) * 1.5 = 0.045 * 1.5 = 0.0675
			duration: 90 * time.Minute, // 1.5 hours
			useSpot:  false,
		},
	}

	estimator := NewEstimator(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimator.CalculatePodCost(tt.pod, tt.duration, tt.useSpot)
			// Allow for small floating point differences
			if diff := abs(got - tt.want); diff > 0.0001 {
				t.Errorf("CalculatePodCost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstimateEnvironmentCost(t *testing.T) {
	tests := []struct {
		pods    []corev1.Pod
		name    string
		want    *v1alpha1.CostEstimate
		ttl     time.Duration
		useSpot bool
	}{
		{
			pods: []corev1.Pod{
				{
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
			name: "estimates cost for single pod environment",
			want: &v1alpha1.CostEstimate{
				Currency:   "USD",
				HourlyCost: "0.05",
				TotalCost:  "0.20",
			},
			ttl:     4 * time.Hour,
			useSpot: false,
		},
		{
			pods: []corev1.Pod{
				{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "frontend",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("500m"),
										corev1.ResourceMemory: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
				{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "backend",
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
				{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "database",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("2"),
										corev1.ResourceMemory: resource.MustParse("4Gi"),
									},
								},
							},
						},
					},
				},
			},
			name: "estimates cost for multi-pod environment",
			want: &v1alpha1.CostEstimate{
				Currency:   "USD",
				HourlyCost: "0.18", // Total hourly: (0.5*0.04 + 1*0.005) + (1*0.04 + 2*0.005) + (2*0.04 + 4*0.005) = 0.025 + 0.05 + 0.1 = 0.175 rounded to 0.18
				TotalCost:  "1.40", // 0.175 * 8 = 1.4
			},
			ttl:     8 * time.Hour,
			useSpot: false,
		},
		{
			pods: []corev1.Pod{
				{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "app",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("2"),
										corev1.ResourceMemory: resource.MustParse("4Gi"),
									},
								},
							},
						},
					},
				},
			},
			name: "estimates cost with spot instances",
			want: &v1alpha1.CostEstimate{
				Currency:   "USD",
				HourlyCost: "0.07", // (2*0.04 + 4*0.005) * 0.7 = 0.1 * 0.7 = 0.07
				TotalCost:  "1.68", // 0.07 * 24 = 1.68
			},
			ttl:     24 * time.Hour,
			useSpot: true,
		},
		{
			pods: []corev1.Pod{},
			name: "handles empty pod list",
			want: &v1alpha1.CostEstimate{
				Currency:   "USD",
				HourlyCost: "0.00",
				TotalCost:  "0.00",
			},
			ttl:     4 * time.Hour,
			useSpot: false,
		},
	}

	estimator := NewEstimator(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimator.EstimateEnvironmentCost(tt.pods, tt.ttl, tt.useSpot)
			if got.Currency != tt.want.Currency {
				t.Errorf("Currency = %v, want %v", got.Currency, tt.want.Currency)
			}
			if got.HourlyCost != tt.want.HourlyCost {
				t.Errorf("HourlyCost = %v, want %v", got.HourlyCost, tt.want.HourlyCost)
			}
			if got.TotalCost != tt.want.TotalCost {
				t.Errorf("TotalCost = %v, want %v", got.TotalCost, tt.want.TotalCost)
			}
		})
	}
}

func TestCalculateDailyCost(t *testing.T) {
	tests := []struct {
		name       string
		hourlyCost float64
		want       float64
	}{
		{
			name:       "calculates daily cost from hourly",
			hourlyCost: 0.10,
			want:       2.40,
		},
		{
			name:       "handles zero cost",
			hourlyCost: 0.00,
			want:       0.00,
		},
		{
			name:       "handles fractional cents",
			hourlyCost: 0.123,
			want:       2.952,
		},
	}

	estimator := NewEstimator(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimator.CalculateDailyCost(tt.hourlyCost)
			if diff := abs(got - tt.want); diff > 0.0001 {
				t.Errorf("CalculateDailyCost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrackActualCost(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		startTime time.Time
		endTime   time.Time
		pods      []corev1.Pod
		useSpot   bool
		want      float64
	}{
		{
			name:      "tracks actual cost for completed environment",
			namespace: "preview-pr-123",
			startTime: time.Now().Add(-2 * time.Hour),
			endTime:   time.Now(),
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "preview-pr-123",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
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
			useSpot: false,
			want:    0.10, // (1*0.04 + 2*0.005) * 2 = 0.05 * 2 = 0.10
		},
	}

	estimator := NewEstimator(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := tt.endTime.Sub(tt.startTime)
			got := estimator.TrackActualCost(tt.namespace, tt.pods, duration, tt.useSpot)
			if diff := abs(got - tt.want); diff > 0.001 {
				t.Errorf("TrackActualCost() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
