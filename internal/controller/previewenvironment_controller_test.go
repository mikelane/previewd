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
	"time"

	previewv1alpha1 "github.com/mikelane/previewd/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("PreviewEnvironment Controller", func() {
	const (
		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When reconciling a resource", func() {
		const resourceName = "test-preview"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}

		var previewenvironment *previewv1alpha1.PreviewEnvironment

		BeforeEach(func() {
			By("creating the custom resource for the Kind PreviewEnvironment")
			previewenvironment = &previewv1alpha1.PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: previewv1alpha1.PreviewEnvironmentSpec{
					Repository: "owner/repo",
					HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
					PRNumber:   123,
					TTL:        "4h",
				},
			}
			Expect(k8sClient.Create(ctx, previewenvironment)).To(Succeed())
		})

		AfterEach(func() {
			// Clean up
			resource := &previewv1alpha1.PreviewEnvironment{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				By("Cleanup the specific resource instance PreviewEnvironment")
				// Remove finalizer to allow deletion in tests
				resource.Finalizers = []string{}
				if updateErr := k8sClient.Update(ctx, resource); updateErr != nil && !apierrors.IsNotFound(updateErr) {
					Expect(updateErr).NotTo(HaveOccurred())
				}
				if deleteErr := k8sClient.Delete(ctx, resource); deleteErr != nil && !apierrors.IsNotFound(deleteErr) {
					Expect(deleteErr).NotTo(HaveOccurred())
				}
			} else if !apierrors.IsNotFound(err) {
				// If error is not "not found", fail the test
				Expect(err).NotTo(HaveOccurred())
			}
			// If NotFound, resource was already cleaned up (expected for deletion tests)
		})

		Describe("Scenario: Reconcile newly created PreviewEnvironment", func() {
			It("adds a finalizer to metadata.finalizers", func() {
				By("reconciling the created resource")
				controllerReconciler := &PreviewEnvironmentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("checking the finalizer was added")
				updatedResource := &previewv1alpha1.PreviewEnvironment{}
				Eventually(func() []string {
					err := k8sClient.Get(ctx, typeNamespacedName, updatedResource)
					if err != nil {
						return nil
					}
					return updatedResource.GetFinalizers()
				}, timeout, interval).Should(ContainElement("preview.previewd.io/finalizer"))
			})

			It("sets status.phase to Pending", func() {
				By("reconciling the created resource")
				controllerReconciler := &PreviewEnvironmentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("checking the phase was set to Pending")
				updatedResource := &previewv1alpha1.PreviewEnvironment{}
				Eventually(func() string {
					err := k8sClient.Get(ctx, typeNamespacedName, updatedResource)
					if err != nil {
						return ""
					}
					return updatedResource.Status.Phase
				}, timeout, interval).Should(Equal("Pending"))
			})

			It("sets status.createdAt to current timestamp", func() {
				By("reconciling the created resource")
				controllerReconciler := &PreviewEnvironmentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				beforeReconcile := time.Now()
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("checking CreatedAt was set")
				updatedResource := &previewv1alpha1.PreviewEnvironment{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, typeNamespacedName, updatedResource)
					if err != nil {
						return false
					}
					return updatedResource.Status.CreatedAt != nil
				}, timeout, interval).Should(BeTrue())

				// Re-fetch to get the latest version
				Expect(k8sClient.Get(ctx, typeNamespacedName, updatedResource)).To(Succeed())

				// Verify timestamp is reasonable (allowing for second-level truncation in metav1.Time)
				Expect(updatedResource.Status.CreatedAt.Time).To(BeTemporally(">=", beforeReconcile.Truncate(time.Second)))
				Expect(updatedResource.Status.CreatedAt.Time).To(BeTemporally("<=", time.Now()))
			})

			It("sets status.observedGeneration to match metadata.generation", func() {
				By("reconciling the created resource")
				controllerReconciler := &PreviewEnvironmentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("checking ObservedGeneration matches Generation")
				updatedResource := &previewv1alpha1.PreviewEnvironment{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, typeNamespacedName, updatedResource)
					if err != nil {
						return false
					}
					return updatedResource.Status.ObservedGeneration == updatedResource.Generation
				}, timeout, interval).Should(BeTrue())
			})

			It("adds a Ready condition with status False", func() {
				By("reconciling the created resource")
				controllerReconciler := &PreviewEnvironmentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("checking Ready condition was added with status False")
				updatedResource := &previewv1alpha1.PreviewEnvironment{}
				Eventually(func() metav1.ConditionStatus {
					err := k8sClient.Get(ctx, typeNamespacedName, updatedResource)
					if err != nil {
						return ""
					}
					condition := meta.FindStatusCondition(updatedResource.Status.Conditions, "Ready")
					if condition != nil {
						return condition.Status
					}
					return ""
				}, timeout, interval).Should(Equal(metav1.ConditionFalse))
			})
		})

		Describe("Scenario: Finalizer prevents premature deletion", func() {
			It("prevents immediate deletion when finalizer is present", func() {
				By("ensuring the resource has a finalizer")
				controllerReconciler := &PreviewEnvironmentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				// First reconcile to add finalizer
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				// Get the resource with finalizer
				updatedResource := &previewv1alpha1.PreviewEnvironment{}
				Eventually(func() []string {
					err := k8sClient.Get(ctx, typeNamespacedName, updatedResource)
					if err != nil {
						return nil
					}
					return updatedResource.GetFinalizers()
				}, timeout, interval).Should(ContainElement("preview.previewd.io/finalizer"))

				By("deleting the PreviewEnvironment")
				Expect(k8sClient.Delete(ctx, updatedResource)).To(Succeed())

				By("confirming CR has deletionTimestamp set but is not removed from etcd")
				Eventually(func() bool {
					err := k8sClient.Get(ctx, typeNamespacedName, updatedResource)
					if err != nil {
						return false
					}
					return updatedResource.DeletionTimestamp != nil
				}, timeout, interval).Should(BeTrue())

				// Verify the resource still exists
				Expect(k8sClient.Get(ctx, typeNamespacedName, updatedResource)).To(Succeed())
				Expect(updatedResource.DeletionTimestamp).NotTo(BeNil())
			})
		})

		Describe("Scenario: Finalizer cleanup on deletion", func() {
			It("removes finalizer and deletes CR when deletion is requested", func() {
				By("ensuring the resource has a finalizer")
				controllerReconciler := &PreviewEnvironmentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				// First reconcile to add finalizer
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				// Get the resource with finalizer
				updatedResource := &previewv1alpha1.PreviewEnvironment{}
				Eventually(func() []string {
					getErr := k8sClient.Get(ctx, typeNamespacedName, updatedResource)
					if getErr != nil {
						return nil
					}
					return updatedResource.GetFinalizers()
				}, timeout, interval).Should(ContainElement("preview.previewd.io/finalizer"))

				By("deleting the PreviewEnvironment")
				Expect(k8sClient.Delete(ctx, updatedResource)).To(Succeed())

				By("reconciling to handle deletion")
				_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("checking the finalizer is removed and CR is deleted")
				Eventually(func() bool {
					getErr := k8sClient.Get(ctx, typeNamespacedName, updatedResource)
					return apierrors.IsNotFound(getErr)
				}, timeout, interval).Should(BeTrue())
			})
		})

		Describe("Scenario: Requeue after successful reconciliation", func() {
			It("returns Result{RequeueAfter: 5m} for Pending phase", func() {
				By("reconciling the created resource")
				controllerReconciler := &PreviewEnvironmentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("checking the result has RequeueAfter set to 5 minutes")
				Expect(result.RequeueAfter).To(Equal(5 * time.Minute))
			})
		})

		Describe("Scenario: Status updates are idempotent", func() {
			It("does not update status when spec hasn't changed", func() {
				By("reconciling the created resource twice")
				controllerReconciler := &PreviewEnvironmentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				// First reconcile
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				// Wait for status to be updated
				updatedResource := &previewv1alpha1.PreviewEnvironment{}
				Eventually(func() string {
					getErr := k8sClient.Get(ctx, typeNamespacedName, updatedResource)
					if getErr != nil {
						return ""
					}
					return updatedResource.Status.Phase
				}, timeout, interval).Should(Equal("Pending"))

				originalGeneration := updatedResource.Status.ObservedGeneration

				// Second reconcile without changes
				_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				// Check that ObservedGeneration hasn't changed
				Expect(k8sClient.Get(ctx, typeNamespacedName, updatedResource)).To(Succeed())
				Expect(updatedResource.Status.ObservedGeneration).To(Equal(originalGeneration))
			})
		})

		Describe("Scenario: Error handling and requeue", func() {
			It("handles NotFound errors gracefully", func() {
				By("attempting to reconcile a non-existent resource")
				controllerReconciler := &PreviewEnvironmentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "non-existent",
						Namespace: "default",
					},
				})

				By("expecting no error and empty result")
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(reconcile.Result{}))
			})
		})
	})

	// Table-driven tests for comprehensive coverage
	Describe("Phase Transitions", func() {
		type phaseTestCase struct {
			name          string
			initialPhase  string
			expectedPhase string
			shouldRequeue bool
			requeueAfter  time.Duration
		}

		DescribeTable("should handle phase transitions correctly",
			func(tc phaseTestCase) {
				ctx := context.Background()
				resourceName := fmt.Sprintf("test-preview-%d", GinkgoRandomSeed())
				typeNamespacedName := types.NamespacedName{
					Name:      resourceName,
					Namespace: "default",
				}

				// Create the resource
				resource := &previewv1alpha1.PreviewEnvironment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: previewv1alpha1.PreviewEnvironmentSpec{
						Repository: "owner/repo",
						HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
						PRNumber:   123,
						TTL:        "4h",
					},
				}

				if tc.initialPhase != "" {
					resource.Status.Phase = tc.initialPhase
				}

				Expect(k8sClient.Create(ctx, resource)).To(Succeed())

				defer func() {
					// Cleanup
					res := &previewv1alpha1.PreviewEnvironment{}
					if err := k8sClient.Get(ctx, typeNamespacedName, res); err == nil {
						res.Finalizers = []string{}
						_ = k8sClient.Update(ctx, res) //nolint:errcheck
						_ = k8sClient.Delete(ctx, res) //nolint:errcheck
					}
				}()

				// Reconcile
				controllerReconciler := &PreviewEnvironmentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				// Verify phase
				updatedResource := &previewv1alpha1.PreviewEnvironment{}
				Eventually(func() string {
					err := k8sClient.Get(ctx, typeNamespacedName, updatedResource)
					if err != nil {
						return ""
					}
					return updatedResource.Status.Phase
				}, time.Second*5).Should(Equal(tc.expectedPhase))

				// Verify requeue
				if tc.shouldRequeue {
					Expect(result.RequeueAfter).To(Equal(tc.requeueAfter))
				} else {
					Expect(result.RequeueAfter).To(BeZero())
				}
			},
			Entry("new resource starts in Pending",
				phaseTestCase{
					name:          "new-pending",
					initialPhase:  "",
					expectedPhase: "Pending",
					shouldRequeue: true,
					requeueAfter:  5 * time.Minute,
				}),
			Entry("Pending phase remains Pending",
				phaseTestCase{
					name:          "pending-stays",
					initialPhase:  "Pending",
					expectedPhase: "Pending",
					shouldRequeue: true,
					requeueAfter:  5 * time.Minute,
				}),
		)
	})

	Describe("Condition Management", func() {
		type conditionTestCase struct {
			name              string
			expectedType      string
			expectedReason    string
			expectedStatus    metav1.ConditionStatus
			initialConditions []metav1.Condition
		}

		DescribeTable("should manage conditions correctly",
			func(tc conditionTestCase) {
				ctx := context.Background()
				resourceName := fmt.Sprintf("test-preview-cond-%d", GinkgoRandomSeed())
				typeNamespacedName := types.NamespacedName{
					Name:      resourceName,
					Namespace: "default",
				}

				// Create the resource
				resource := &previewv1alpha1.PreviewEnvironment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: previewv1alpha1.PreviewEnvironmentSpec{
						Repository: "owner/repo",
						HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
						PRNumber:   123,
					},
				}

				if tc.initialConditions != nil {
					resource.Status.Conditions = tc.initialConditions
				}

				Expect(k8sClient.Create(ctx, resource)).To(Succeed())

				defer func() {
					// Cleanup
					res := &previewv1alpha1.PreviewEnvironment{}
					if err := k8sClient.Get(ctx, typeNamespacedName, res); err == nil {
						res.Finalizers = []string{}
						_ = k8sClient.Update(ctx, res) //nolint:errcheck
						_ = k8sClient.Delete(ctx, res) //nolint:errcheck
					}
				}()

				// Reconcile
				controllerReconciler := &PreviewEnvironmentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				// Verify condition
				updatedResource := &previewv1alpha1.PreviewEnvironment{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, typeNamespacedName, updatedResource)
					if err != nil {
						return false
					}
					condition := meta.FindStatusCondition(updatedResource.Status.Conditions, tc.expectedType)
					if condition != nil {
						return condition.Status == tc.expectedStatus &&
							condition.Reason == tc.expectedReason
					}
					return false
				}, time.Second*5).Should(BeTrue())
			},
			Entry("adds Ready condition to new resource",
				conditionTestCase{
					name:           "new-ready",
					expectedType:   "Ready",
					expectedStatus: metav1.ConditionFalse,
					expectedReason: "Reconciling",
				}),
			Entry("updates existing Ready condition",
				conditionTestCase{
					name: "update-ready",
					initialConditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Now(),
							Reason:             "Complete",
							Message:            "Environment is ready",
						},
					},
					expectedType:   "Ready",
					expectedStatus: metav1.ConditionFalse,
					expectedReason: "Reconciling",
				}),
		)
	})

	Describe("Error path handling", func() {
		const (
			timeout  = time.Second * 10
			interval = time.Millisecond * 250
		)

		Context("When reconciling with error conditions", func() {
			const resourceName = "test-error-preview"

			ctx := context.Background()

			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}

			var previewenvironment *previewv1alpha1.PreviewEnvironment

			BeforeEach(func() {
				By("creating the custom resource for error testing")
				previewenvironment = &previewv1alpha1.PreviewEnvironment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: previewv1alpha1.PreviewEnvironmentSpec{
						Repository: "owner/repo",
						HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
						PRNumber:   123,
						TTL:        "4h",
					},
				}
				Expect(k8sClient.Create(ctx, previewenvironment)).To(Succeed())
			})

			AfterEach(func() {
				// Clean up
				resource := &previewv1alpha1.PreviewEnvironment{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				if err == nil {
					By("Cleanup the specific resource instance PreviewEnvironment")
					resource.Finalizers = []string{}
					if updateErr := k8sClient.Update(ctx, resource); updateErr != nil && !apierrors.IsNotFound(updateErr) {
						Expect(updateErr).NotTo(HaveOccurred())
					}
					if deleteErr := k8sClient.Delete(ctx, resource); deleteErr != nil && !apierrors.IsNotFound(deleteErr) {
						Expect(deleteErr).NotTo(HaveOccurred())
					}
				} else if !apierrors.IsNotFound(err) {
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("handles reconciliation when resource doesn't exist", func() {
				By("attempting to reconcile a non-existent resource")
				controllerReconciler := &PreviewEnvironmentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "definitely-does-not-exist",
						Namespace: "default",
					},
				})

				By("expecting no error (graceful handling of missing resources)")
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(reconcile.Result{}))
			})

			It("handles reconciliation with successful finalizer updates", func() {
				By("reconciling the created resource to add finalizer")
				controllerReconciler := &PreviewEnvironmentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})

				By("expecting successful reconciliation with requeue")
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RequeueAfter).To(Equal(5 * time.Minute))

				By("verifying finalizer was added")
				updatedResource := &previewv1alpha1.PreviewEnvironment{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, typeNamespacedName, updatedResource)
					if err != nil {
						return false
					}
					return controllerutil.ContainsFinalizer(updatedResource, "preview.previewd.io/finalizer")
				}, timeout, interval).Should(BeTrue())
			})

			It("uses defaultRequeueAfter constant for requeue duration", func() {
				By("reconciling the created resource")
				controllerReconciler := &PreviewEnvironmentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})

				By("verifying the result uses defaultRequeueAfter constant")
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RequeueAfter).To(Equal(5 * time.Minute))
			})
		})
	})
})
