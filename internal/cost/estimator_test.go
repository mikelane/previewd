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
		name   string
		config *Config
		want   *Config
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
		name     string
		pod      *corev1.Pod
		duration time.Duration
		want     float64
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
		name    string
		pods    []corev1.Pod
		ttl     time.Duration
		want    *v1alpha1.CostEstimate
		useSpot bool
	}{
		{
			name: "estimates cost for single pod environment",
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
			ttl: 4 * time.Hour,
			want: &v1alpha1.CostEstimate{
				Currency:   "USD",
				HourlyCost: "0.0500",
				TotalCost:  "0.2000",
			},
			useSpot: false,
		},
		{
			name: "estimates cost for multi-pod environment",
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
			ttl: 8 * time.Hour,
			want: &v1alpha1.CostEstimate{
				Currency:   "USD",
				HourlyCost: "0.1750", // Total hourly: (0.5*0.04 + 1*0.005) + (1*0.04 + 2*0.005) + (2*0.04 + 4*0.005) = 0.025 + 0.05 + 0.1 = 0.175
				TotalCost:  "1.4000", // 0.175 * 8 = 1.4
			},
			useSpot: false,
		},
		{
			name: "estimates cost with spot instances",
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
			ttl: 24 * time.Hour,
			want: &v1alpha1.CostEstimate{
				Currency:   "USD",
				HourlyCost: "0.0700", // (2*0.04 + 4*0.005) * 0.7 = 0.1 * 0.7 = 0.07
				TotalCost:  "1.6800", // 0.07 * 24 = 1.68
			},
			useSpot: true,
		},
		{
			name: "handles empty pod list",
			pods: []corev1.Pod{},
			ttl:  4 * time.Hour,
			want: &v1alpha1.CostEstimate{
				Currency:   "USD",
				HourlyCost: "0.0000",
				TotalCost:  "0.0000",
			},
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
		pods      []corev1.Pod
		startTime time.Time
		endTime   time.Time
		want      float64
		useSpot   bool
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

func TestUpdateConfigConcurrency(t *testing.T) {
	estimator := NewEstimator(nil)

	// Create multiple configs to update concurrently
	configs := []*Config{
		{CPUCostPerHour: 0.05, MemoryCostPerHour: 0.006, SpotDiscount: 0.25, Currency: "USD"},
		{CPUCostPerHour: 0.06, MemoryCostPerHour: 0.007, SpotDiscount: 0.35, Currency: "EUR"},
		{CPUCostPerHour: 0.07, MemoryCostPerHour: 0.008, SpotDiscount: 0.45, Currency: "GBP"},
		{CPUCostPerHour: 0.08, MemoryCostPerHour: 0.009, SpotDiscount: 0.55, Currency: "JPY"},
	}

	// Run concurrent updates
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(iteration int) {
			config := configs[iteration%len(configs)]
			estimator.UpdateConfig(config)
			// Also read config concurrently to test RLock
			_ = estimator.GetConfig()
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify estimator still has a valid config
	finalConfig := estimator.GetConfig()
	if finalConfig == nil {
		t.Fatal("GetConfig returned nil after concurrent updates")
	}

	// Check that final config is one of our test configs
	validConfig := false
	for _, cfg := range configs {
		if finalConfig.CPUCostPerHour == cfg.CPUCostPerHour &&
			finalConfig.MemoryCostPerHour == cfg.MemoryCostPerHour &&
			finalConfig.SpotDiscount == cfg.SpotDiscount &&
			finalConfig.Currency == cfg.Currency {
			validConfig = true
			break
		}
	}

	if !validConfig {
		t.Errorf("Final config does not match any test config: %+v", finalConfig)
	}
}

func TestFormatCostVerySmall(t *testing.T) {
	tests := []struct {
		name string
		cost float64
		want string
	}{
		{
			name: "formats cost less than $0.01",
			cost: 0.0042,
			want: "0.0042",
		},
		{
			name: "formats cost less than $0.001",
			cost: 0.0003,
			want: "0.0003",
		},
		{
			name: "formats cost close to zero",
			cost: 0.00001,
			want: "0.0000",
		},
		{
			name: "formats exact zero",
			cost: 0.0,
			want: "0.0000",
		},
		{
			name: "formats normal cost",
			cost: 1.2345,
			want: "1.2345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCost(tt.cost)
			if got != tt.want {
				t.Errorf("formatCost(%v) = %v, want %v", tt.cost, got, tt.want)
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
