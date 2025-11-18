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

package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	previewv1alpha1 "github.com/mikelane/previewd/api/v1alpha1"
	"github.com/mikelane/previewd/internal/cost"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// PreviewEnvironmentReconciler reconciles a PreviewEnvironment object
type PreviewEnvironmentReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	CostEstimator *cost.Estimator
}

// +kubebuilder:rbac:groups=preview.previewd.io,resources=previewenvironments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=preview.previewd.io,resources=previewenvironments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=preview.previewd.io,resources=previewenvironments/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.1/pkg/reconcile
func (r *PreviewEnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the PreviewEnvironment instance
	var preview previewv1alpha1.PreviewEnvironment
	if err := r.Get(ctx, req.NamespacedName, &preview); err != nil {
		// Resource not found, return without error
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Initialize cost estimator if not already done
	if r.CostEstimator == nil {
		r.CostEstimator = cost.NewEstimator(nil)
	}

	// Skip cost calculation if environment namespace is not set
	if preview.Status.Namespace == "" {
		log.Info("Preview environment namespace not yet created, skipping cost estimation")
		return ctrl.Result{}, nil
	}

	// List all pods in the preview environment namespace
	var podList corev1.PodList
	if err := r.List(ctx, &podList, client.InNamespace(preview.Status.Namespace)); err != nil {
		log.Error(err, "Failed to list pods", "namespace", preview.Status.Namespace)
		return ctrl.Result{}, err
	}

	// Parse TTL from spec
	ttl, err := parseTTL(preview.Spec.TTL)
	if err != nil {
		log.Error(err, "Failed to parse TTL, using default", "ttl", preview.Spec.TTL)
		ttl = 4 * time.Hour
	}

	// Check if spot instances should be used
	useSpot := checkSpotInstance(&preview)

	// Calculate cost estimate
	costEstimate := r.CostEstimator.EstimateEnvironmentCost(podList.Items, ttl, useSpot)

	// Update status with cost estimate
	preview.Status.CostEstimate = costEstimate

	// Update the status
	if err := r.Status().Update(ctx, &preview); err != nil {
		log.Error(err, "Failed to update preview environment status")
		return ctrl.Result{}, err
	}

	log.Info("Updated cost estimate",
		"namespace", preview.Status.Namespace,
		"hourlyCost", costEstimate.HourlyCost,
		"totalCost", costEstimate.TotalCost,
		"useSpot", useSpot)

	// Requeue after 5 minutes to update cost estimates
	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PreviewEnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&previewv1alpha1.PreviewEnvironment{}).
		Named("previewenvironment").
		Complete(r)
}

// parseTTL parses a TTL string and returns a time.Duration.
// Supports formats like "4h", "30m", "2d" (days).
// Returns default of 4 hours if empty or invalid.
func parseTTL(ttl string) (time.Duration, error) {
	if ttl == "" {
		return 4 * time.Hour, nil
	}

	// Handle days specially (e.g., "2d" -> 48h)
	if strings.HasSuffix(ttl, "d") {
		daysStr := strings.TrimSuffix(ttl, "d")
		var days int
		_, err := fmt.Sscanf(daysStr, "%d", &days)
		if err != nil {
			return 0, fmt.Errorf("invalid TTL format: %s", ttl)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	// Parse standard Go duration formats
	duration, err := time.ParseDuration(ttl)
	if err != nil {
		return 0, fmt.Errorf("invalid TTL format: %s", ttl)
	}

	return duration, nil
}

// checkSpotInstance checks if the preview environment should use spot instances
func checkSpotInstance(preview *previewv1alpha1.PreviewEnvironment) bool {
	if preview.Annotations == nil {
		return false
	}

	spotAnnotation, exists := preview.Annotations["previewd.io/use-spot"]
	if !exists {
		return false
	}

	return spotAnnotation == "true"
}
