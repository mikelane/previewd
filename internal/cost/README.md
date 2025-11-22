# Cost Estimator

The cost estimator package provides resource-based cost estimation for preview environments in the Previewd operator.

## Overview

This package calculates estimated costs for running Kubernetes preview environments based on pod resource requests (CPU and memory). It supports both on-demand and spot instance pricing models.

## Features

- **Resource-based cost calculation**: Estimates costs based on CPU and memory resource requests
- **Spot instance support**: Applies configurable discounts for spot instances
- **Configurable pricing**: Customizable cost per CPU-hour and memory GB-hour
- **Multiple currency support**: Default USD with ability to configure other currencies
- **TTL-aware estimates**: Calculates total costs based on environment time-to-live
- **Actual cost tracking**: Tracks actual costs based on runtime duration

## Default Pricing

The default pricing configuration (USD):
- CPU: $0.04 per vCPU-hour
- Memory: $0.005 per GB-hour
- Spot discount: 30% off on-demand prices

## Usage

### Basic Usage

```go
import "github.com/mikelane/previewd/internal/cost"

// Create estimator with default config
estimator := cost.NewEstimator(nil)

// Calculate cost for a single pod
podCost := estimator.CalculatePodCost(pod, 4*time.Hour, false)

// Estimate environment cost for multiple pods
pods := []corev1.Pod{...}
ttl := 8 * time.Hour
useSpot := true
estimate := estimator.EstimateEnvironmentCost(pods, ttl, useSpot)
```

### Custom Pricing Configuration

```go
config := &cost.Config{
    CPUCostPerHour:    0.08,   // $0.08 per vCPU-hour
    MemoryCostPerHour: 0.01,   // $0.01 per GB-hour
    SpotDiscount:      0.40,   // 40% discount
    Currency:          "EUR",
}

estimator := cost.NewEstimator(config)
```

### Controller Integration

The cost estimator is integrated with the PreviewEnvironment controller:

1. The controller lists all pods in the preview environment namespace
2. Calculates cost estimates based on resource requests
3. Updates the PreviewEnvironment status with cost information
4. Re-calculates costs every 5 minutes

### Spot Instance Configuration

To use spot instance pricing, add an annotation to the PreviewEnvironment:

```yaml
apiVersion: preview.previewd.io/v1alpha1
kind: PreviewEnvironment
metadata:
  name: my-preview
  annotations:
    previewd.io/use-spot: "true"
spec:
  # ...
```

## Cost Formula

The cost calculation formula is:

```
Cost = (CPU_cores * CPU_rate * hours) + (Memory_GB * Memory_rate * hours)
```

For spot instances:
```
Spot_Cost = Cost * (1 - Spot_discount)
```

## API Reference

### Types

#### Config
```go
type Config struct {
    CPUCostPerHour    float64 // Cost per vCPU hour
    MemoryCostPerHour float64 // Cost per GB memory hour
    SpotDiscount      float64 // Discount for spot instances (0.3 = 30%)
    Currency          string  // Currency code (e.g., "USD")
}
```

#### Estimator
```go
type Estimator struct {
    // Contains pricing configuration
}
```

### Functions

#### NewEstimator
```go
func NewEstimator(config *Config) *Estimator
```
Creates a new cost estimator. If config is nil, uses default configuration.

#### CalculatePodCost
```go
func (e *Estimator) CalculatePodCost(pod *corev1.Pod, duration time.Duration, useSpot bool) float64
```
Calculates the cost of running a pod for the specified duration.

#### EstimateEnvironmentCost
```go
func (e *Estimator) EstimateEnvironmentCost(pods []corev1.Pod, ttl time.Duration, useSpot bool) *v1alpha1.CostEstimate
```
Estimates the total cost of running all pods in an environment.

#### CalculateDailyCost
```go
func (e *Estimator) CalculateDailyCost(hourlyCost float64) float64
```
Calculates daily cost from hourly cost (hourly * 24).

#### TrackActualCost
```go
func (e *Estimator) TrackActualCost(namespace string, pods []corev1.Pod, actualDuration time.Duration, useSpot bool) float64
```
Tracks the actual cost of a completed environment based on actual runtime.

## Testing

The package includes comprehensive unit tests with >80% code coverage:

```bash
go test ./internal/cost/... -cover
```

Tests cover:
- Default and custom configuration
- Single and multi-container pods
- Spot instance pricing
- Empty resource requests
- Various time durations
- TTL parsing (hours, minutes, days)

## Future Enhancements

Potential improvements for future versions:
- Storage cost estimation
- Network egress cost estimation
- GPU resource cost support
- Cloud provider-specific pricing APIs
- Historical cost tracking and reporting
- Cost alerts and budgets
- Multi-region pricing support
