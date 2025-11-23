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

// Package cost provides cost estimation for preview environment resources.
//
// This package calculates the estimated cloud infrastructure costs for preview
// environments by analyzing Kubernetes pod resource requests (CPU and memory).
//
// Key features:
//   - Calculates hourly, daily, and monthly cost estimates
//   - Configurable pricing for CPU and memory
//   - Tracks total cost across all preview environments
//   - Accounts for spot instance discounts
//   - Thread-safe cost calculations
//
// Cost Calculation:
//
// Costs are calculated based on pod resource requests:
//
//	CPU Cost = (Total CPU Cores) × (CPU Price Per Hour)
//	Memory Cost = (Total Memory GB) × (Memory Price Per Hour)
//	Total Hourly Cost = CPU Cost + Memory Cost
//	Daily Cost = Hourly Cost × 24
//	Monthly Cost = Hourly Cost × 730 (average hours per month)
//
// Default Pricing:
//
//   - CPU: $0.04 per core per hour
//   - Memory: $0.005 per GB per hour
//   - Spot Discount: 70% (when enabled)
//
// Example usage:
//
//	config := cost.DefaultConfig()
//	estimator := cost.NewEstimator(k8sClient, config)
//
//	// Estimate cost for a single environment
//	estimate, err := estimator.EstimateCost(ctx, "preview-pr-123")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Daily cost: %s%.2f\n", estimate.Currency, estimate.DailyCost)
//
//	// Get total cost across all environments
//	total, err := estimator.GetTotalCost(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Total monthly cost: %s%.2f\n", total.Currency, total.MonthlyCost)
package cost
