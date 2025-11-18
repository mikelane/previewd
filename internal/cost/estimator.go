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

// Package cost provides cost estimation functionality for preview environments
package cost

import (
	"fmt"
	"sync"
	"time"

	"github.com/mikelane/previewd/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Config defines the pricing configuration for cost estimation
type Config struct {
	Currency          string
	CPUCostPerHour    float64
	MemoryCostPerHour float64
	SpotDiscount      float64
}

// DefaultConfig returns the default pricing configuration
func DefaultConfig() *Config {
	return &Config{
		CPUCostPerHour:    0.04,  // $0.04 per vCPU-hour
		MemoryCostPerHour: 0.005, // $0.005 per GB-hour
		SpotDiscount:      0.30,  // 30% discount for spot instances
		Currency:          "USD",
	}
}

// Estimator calculates costs for preview environments
type Estimator struct {
	config *Config
	mu     sync.RWMutex
}

// NewEstimator creates a new cost estimator with the given configuration.
// If config is nil, default configuration is used.
func NewEstimator(config *Config) *Estimator {
	if config == nil {
		config = DefaultConfig()
	}
	return &Estimator{
		config: config,
	}
}

// CalculatePodCost calculates the cost of running a pod for the specified duration.
// If useSpot is true, spot instance pricing is applied.
func (e *Estimator) CalculatePodCost(pod *corev1.Pod, duration time.Duration, useSpot bool) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var totalCPU float64
	var totalMemoryGB float64

	// Sum resources across all containers
	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests != nil {
			// Calculate CPU (convert milliCPU to CPU cores)
			if cpu, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
				totalCPU += float64(cpu.MilliValue()) / 1000.0
			}

			// Calculate memory (convert to GB)
			if memory, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
				// Convert bytes to GB
				totalMemoryGB += float64(memory.Value()) / (1024 * 1024 * 1024)
			}
		}
	}

	// Calculate hours from duration
	hours := duration.Hours()

	// Calculate base cost
	cpuCost := totalCPU * e.config.CPUCostPerHour * hours
	memoryCost := totalMemoryGB * e.config.MemoryCostPerHour * hours
	totalCost := cpuCost + memoryCost

	// Apply spot discount if applicable
	if useSpot {
		totalCost = totalCost * (1 - e.config.SpotDiscount)
	}

	return totalCost
}

// EstimateEnvironmentCost estimates the total cost of running all pods in an environment
func (e *Estimator) EstimateEnvironmentCost(pods []corev1.Pod, ttl time.Duration, useSpot bool) *v1alpha1.CostEstimate {
	var totalHourlyCost float64

	// Calculate hourly cost for each pod (based on 1 hour duration)
	for _, pod := range pods {
		podHourlyCost := e.CalculatePodCost(&pod, 1*time.Hour, useSpot)
		totalHourlyCost += podHourlyCost
	}

	// Calculate total cost based on TTL
	totalCost := totalHourlyCost * ttl.Hours()

	return &v1alpha1.CostEstimate{
		Currency:   e.config.Currency,
		HourlyCost: formatCost(totalHourlyCost),
		TotalCost:  formatCost(totalCost),
	}
}

// CalculateDailyCost calculates the daily cost from hourly cost
func (e *Estimator) CalculateDailyCost(hourlyCost float64) float64 {
	return hourlyCost * 24
}

// TrackActualCost tracks the actual cost of a completed environment
func (e *Estimator) TrackActualCost(namespace string, pods []corev1.Pod, actualDuration time.Duration, useSpot bool) float64 {
	var totalCost float64

	// Filter pods by namespace and calculate their costs
	for _, pod := range pods {
		if pod.Namespace == namespace {
			podCost := e.CalculatePodCost(&pod, actualDuration, useSpot)
			totalCost += podCost
		}
	}

	return totalCost
}

// formatCost formats a cost value as a string with 4 decimal places for transparency
func formatCost(cost float64) string {
	return fmt.Sprintf("%.4f", cost)
}

// GetConfig returns the current pricing configuration
func (e *Estimator) GetConfig() *Config {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config
}

// UpdateConfig updates the pricing configuration
func (e *Estimator) UpdateConfig(config *Config) {
	if config != nil {
		e.mu.Lock()
		defer e.mu.Unlock()
		e.config = config
	}
}

// ParseResourceQuantity parses a Kubernetes resource quantity and returns the value in the base unit
func ParseResourceQuantity(quantity resource.Quantity, resourceType corev1.ResourceName) float64 {
	switch resourceType {
	case corev1.ResourceCPU:
		// CPU is in millicores, convert to cores
		return float64(quantity.MilliValue()) / 1000.0
	case corev1.ResourceMemory:
		// Memory is in bytes, convert to GB
		return float64(quantity.Value()) / (1024 * 1024 * 1024)
	default:
		return 0
	}
}
