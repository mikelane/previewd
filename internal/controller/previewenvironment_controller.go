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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// finalizerName is the finalizer used to prevent deletion until cleanup is complete
	finalizerName = "preview.previewd.io/finalizer"

	// defaultRequeueAfter is the default duration to requeue the reconciliation loop.
	// This value controls how frequently the operator checks each PreviewEnvironment
	// for state changes. Keep this reasonably large (minutes, not seconds) to avoid
	// excessive API server load. Specific events (CR creation, deletion, etc.) will
	// trigger immediate reconciliation via webhooks.
	defaultRequeueAfter = 5 * time.Minute
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
	logger := logf.FromContext(ctx)

	// Fetch the PreviewEnvironment resource
	previewEnv := &previewv1alpha1.PreviewEnvironment{}
	if err := r.Get(ctx, req.NamespacedName, previewEnv); err != nil {
		if apierrors.IsNotFound(err) {
			// Resource has been deleted, nothing to do
			logger.Info("PreviewEnvironment resource not found, likely deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request
		logger.Error(err, "Failed to get PreviewEnvironment")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !previewEnv.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, previewEnv)
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(previewEnv, finalizerName) {
		controllerutil.AddFinalizer(previewEnv, finalizerName)
		if err := r.Update(ctx, previewEnv); err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		logger.Info("Added finalizer to PreviewEnvironment")

		// Re-fetch the resource after updating finalizer to ensure we have latest version
		if err := r.Get(ctx, req.NamespacedName, previewEnv); err != nil {
			logger.Error(err, "Failed to re-fetch PreviewEnvironment after finalizer update")
			return ctrl.Result{}, err
		}
	}

	// Initialize status if this is a new resource
	if err := r.initializeStatus(ctx, previewEnv); err != nil {
		logger.Error(err, "Failed to initialize status")
		return ctrl.Result{}, err
	}

	// TODO(#3): Create namespace using namespace manager
	// This will create a dedicated namespace for the preview environment
	// with appropriate RBAC and resource quotas.

	// TODO(#4): Deploy services via ArgoCD ApplicationSet
	// Parse the repository and create an ArgoCD ApplicationSet that deploys
	// services to the preview namespace based on the repository structure.

	// Perform cost estimation after status is initialized
	if err := r.estimateAndUpdateCosts(ctx, previewEnv); err != nil {
		logger.Error(err, "Failed to estimate costs (non-fatal, will retry)")
		// Log the error but don't fail - cost estimation is best-effort
	}

	// TODO(#5): Implement TTL-based expiration check and cleanup
	// Check if the environment has exceeded its TTL (time-to-live) and
	// trigger cleanup if necessary.

	// Requeue after the default interval for periodic reconciliation
	return ctrl.Result{RequeueAfter: defaultRequeueAfter}, nil
}

// handleDeletion performs cleanup when a PreviewEnvironment is being deleted
func (r *PreviewEnvironmentReconciler) handleDeletion(ctx context.Context, previewEnv *previewv1alpha1.PreviewEnvironment) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	if controllerutil.ContainsFinalizer(previewEnv, finalizerName) {
		// Perform cleanup here (will be implemented in future PRs)
		// For now, just remove the finalizer to allow deletion
		logger.Info("Performing cleanup for PreviewEnvironment deletion")

		controllerutil.RemoveFinalizer(previewEnv, finalizerName)
		if err := r.Update(ctx, previewEnv); err != nil {
			logger.Error(err, "Failed to remove finalizer")
			return ctrl.Result{}, err
		}
		logger.Info("Removed finalizer, PreviewEnvironment can now be deleted")
	}

	return ctrl.Result{}, nil
}

// estimateAndUpdateCosts performs cost estimation for the preview environment
func (r *PreviewEnvironmentReconciler) estimateAndUpdateCosts(ctx context.Context, previewEnv *previewv1alpha1.PreviewEnvironment) error {
	logger := logf.FromContext(ctx)
	// Initialize cost estimator if not already done
	if r.CostEstimator == nil {
		r.CostEstimator = cost.NewEstimator(nil)
	}

	// Skip cost calculation if environment namespace is not set
	if previewEnv.Status.Namespace == "" {
		logger.Info("Preview environment namespace not yet created, skipping cost estimation")
		return nil
	}

	// List all pods in the preview environment namespace
	var podList corev1.PodList
	if err := r.List(ctx, &podList, client.InNamespace(previewEnv.Status.Namespace)); err != nil {
		return fmt.Errorf("failed to list pods in namespace %s: %w", previewEnv.Status.Namespace, err)
	}

	// Parse TTL from spec
	ttl, err := parseTTL(previewEnv.Spec.TTL)
	if err != nil {
		logger.Error(err, "Failed to parse TTL, using default", "ttl", previewEnv.Spec.TTL)
		ttl = 4 * time.Hour
	}

	// Check if spot instances should be used
	useSpot := checkSpotInstance(previewEnv)

	// Calculate cost estimate
	costEstimate := r.CostEstimator.EstimateEnvironmentCost(podList.Items, ttl, useSpot)

	// Update status with cost estimate
	previewEnv.Status.CostEstimate = costEstimate

	// Update the status
	if err := r.Status().Update(ctx, previewEnv); err != nil {
		return fmt.Errorf("failed to update preview environment status: %w", err)
	}

	logger.Info("Updated cost estimate",
		"namespace", previewEnv.Status.Namespace,
		"hourlyCost", costEstimate.HourlyCost,
		"totalCost", costEstimate.TotalCost,
		"useSpot", useSpot)

	return nil
}

// initializeStatus sets up initial status fields for a new PreviewEnvironment
func (r *PreviewEnvironmentReconciler) initializeStatus(ctx context.Context, previewEnv *previewv1alpha1.PreviewEnvironment) error {
	logger := logf.FromContext(ctx)

	// Check if status has already been initialized
	if previewEnv.Status.Phase != "" && previewEnv.Status.CreatedAt != nil {
		// Status already initialized, skip
		return nil
	}

	// Set initial phase
	previewEnv.Status.Phase = "Pending"

	// Set creation timestamp if not already set
	if previewEnv.Status.CreatedAt == nil {
		now := metav1.NewTime(time.Now())
		previewEnv.Status.CreatedAt = &now
	}

	// Set observed generation
	previewEnv.Status.ObservedGeneration = previewEnv.Generation

	// Set Ready condition to False
	meta.SetStatusCondition(&previewEnv.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		ObservedGeneration: previewEnv.Generation,
		Reason:             "Reconciling",
		Message:            "PreviewEnvironment is being reconciled",
		LastTransitionTime: metav1.Now(),
	})

	// Update status subresource
	if err := r.Status().Update(ctx, previewEnv); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	logger.Info("Initialized PreviewEnvironment status", "phase", previewEnv.Status.Phase)
	return nil
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
