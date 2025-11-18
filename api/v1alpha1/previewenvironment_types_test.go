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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("PreviewEnvironment", func() {
	Context("Repository field", func() {
		It("exists and accepts valid repository format", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-preview",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
				},
			}

			Expect(preview.Spec.Repository).To(Equal("owner/repo"))
		})

		It("rejects invalid repository format via Kubernetes API validation", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-repo-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "invalid-format",
				},
			}

			err := k8sClient.Create(ctx, preview)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("spec.repository"))
		})
	})

	Context("PRNumber field", func() {
		It("exists and accepts valid PR number", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pr-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   123,
				},
			}

			Expect(preview.Spec.PRNumber).To(Equal(123))
		})

		It("rejects PR number less than 1", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-pr-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   0,
				},
			}

			err := k8sClient.Create(ctx, preview)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("spec.prNumber"))
		})
	})

	Context("HeadSHA field", func() {
		It("exists and accepts valid 40-character SHA", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sha-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   123,
					HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
				},
			}

			Expect(preview.Spec.HeadSHA).To(Equal("1234567890abcdef1234567890abcdef12345678"))
		})

		It("rejects invalid SHA format", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-sha-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   123,
					HeadSHA:    "short",
				},
			}

			err := k8sClient.Create(ctx, preview)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("spec.headSHA"))
		})
	})

	Context("Optional fields", func() {
		It("accepts Services slice when provided", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "services-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   123,
					HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
					Services:   []string{"api", "web", "worker"},
				},
			}

			Expect(preview.Spec.Services).To(HaveLen(3))
			Expect(preview.Spec.Services).To(ContainElements("api", "web", "worker"))
		})
	})

	Context("Status fields", func() {
		It("accepts Phase field", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "status-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   123,
					HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
				},
				Status: PreviewEnvironmentStatus{
					Phase: "Ready",
				},
			}

			Expect(preview.Status.Phase).To(Equal("Ready"))
		})

		It("accepts URL field", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "url-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   123,
					HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
				},
				Status: PreviewEnvironmentStatus{
					URL: "https://pr-123.preview.example.com",
				},
			}

			Expect(preview.Status.URL).To(Equal("https://pr-123.preview.example.com"))
		})

		It("accepts Namespace field", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "namespace-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   123,
					HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
				},
				Status: PreviewEnvironmentStatus{
					Namespace: "preview-pr-123",
				},
			}

			Expect(preview.Status.Namespace).To(Equal("preview-pr-123"))
		})

		It("accepts timestamps", func() {
			now := metav1.Now()
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "timestamps-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   123,
					HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
				},
				Status: PreviewEnvironmentStatus{
					CreatedAt:          &now,
					ExpiresAt:          &now,
					LastSyncedAt:       &now,
					ObservedGeneration: 1,
				},
			}

			Expect(preview.Status.CreatedAt).NotTo(BeNil())
			Expect(preview.Status.ExpiresAt).NotTo(BeNil())
			Expect(preview.Status.LastSyncedAt).NotTo(BeNil())
			Expect(preview.Status.ObservedGeneration).To(Equal(int64(1)))
		})
	})

	Context("Complete valid PreviewEnvironment", func() {
		It("creates successfully with all required fields", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "complete-valid-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "myorg/myrepo",
					PRNumber:   123,
					HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
				},
			}

			err := k8sClient.Create(ctx, preview)
			Expect(err).NotTo(HaveOccurred())

			// Fetch it back
			fetched := &PreviewEnvironment{}
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(preview), fetched)
			Expect(err).NotTo(HaveOccurred())

			// Verify spec fields
			Expect(fetched.Spec.Repository).To(Equal("myorg/myrepo"))
			Expect(fetched.Spec.PRNumber).To(Equal(123))
			Expect(fetched.Spec.HeadSHA).To(Equal("1234567890abcdef1234567890abcdef12345678"))

			// Cleanup
			err = k8sClient.Delete(ctx, preview)
			Expect(err).NotTo(HaveOccurred())
		})

		It("allows status updates", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "status-update-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "myorg/myrepo",
					PRNumber:   456,
					HeadSHA:    "def4567890abc1234def4567890abc1234def456",
				},
			}

			err := k8sClient.Create(ctx, preview)
			Expect(err).NotTo(HaveOccurred())

			// Update status
			preview.Status.Phase = "Ready"
			preview.Status.URL = "https://pr-456.preview.example.com"
			err = k8sClient.Status().Update(ctx, preview)
			Expect(err).NotTo(HaveOccurred())

			// Fetch and verify
			fetched := &PreviewEnvironment{}
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(preview), fetched)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetched.Status.Phase).To(Equal("Ready"))
			Expect(fetched.Status.URL).To(Equal("https://pr-456.preview.example.com"))

			// Cleanup
			err = k8sClient.Delete(ctx, preview)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("DeepCopy methods", func() {
		It("returns nil when copying nil CostEstimate", func() {
			var original *CostEstimate
			copied := original.DeepCopy()
			Expect(copied).To(BeNil())
		})

		It("deep copies CostEstimate correctly", func() {
			original := &CostEstimate{
				Currency:   "USD",
				HourlyCost: "0.05",
				TotalCost:  "0.20",
			}

			copied := original.DeepCopy()
			Expect(copied).NotTo(BeIdenticalTo(original))
			Expect(copied.Currency).To(Equal(original.Currency))
			Expect(copied.HourlyCost).To(Equal(original.HourlyCost))
			Expect(copied.TotalCost).To(Equal(original.TotalCost))

			// Modify copy, original should be unchanged
			copied.Currency = "EUR"
			Expect(original.Currency).To(Equal("USD"))
		})

		It("returns nil when copying nil ServiceStatus", func() {
			var original *ServiceStatus
			copied := original.DeepCopy()
			Expect(copied).To(BeNil())
		})

		It("deep copies ServiceStatus correctly", func() {
			original := &ServiceStatus{
				Name:  "api",
				Ready: true,
				URL:   "https://api.example.com",
			}

			copied := original.DeepCopy()
			Expect(copied).NotTo(BeIdenticalTo(original))
			Expect(copied.Name).To(Equal(original.Name))
			Expect(copied.Ready).To(Equal(original.Ready))
			Expect(copied.URL).To(Equal(original.URL))
		})

		It("returns nil when copying nil PreviewEnvironmentSpec", func() {
			var original *PreviewEnvironmentSpec
			copied := original.DeepCopy()
			Expect(copied).To(BeNil())
		})

		It("deep copies PreviewEnvironmentSpec correctly", func() {
			original := &PreviewEnvironmentSpec{
				Repository: "owner/repo",
				PRNumber:   123,
				HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
				Services:   []string{"api", "web"},
			}

			copied := original.DeepCopy()
			Expect(copied).NotTo(BeIdenticalTo(original))
			Expect(copied.Repository).To(Equal(original.Repository))
			Expect(copied.PRNumber).To(Equal(original.PRNumber))
			Expect(copied.Services).To(Equal(original.Services))

			// Modify copy's slice, original should be unchanged
			copied.Services = append(copied.Services, "worker")
			Expect(original.Services).To(HaveLen(2))
		})

		It("returns nil when copying nil PreviewEnvironmentStatus", func() {
			var original *PreviewEnvironmentStatus
			copied := original.DeepCopy()
			Expect(copied).To(BeNil())
		})

		It("deep copies PreviewEnvironmentStatus with all nil fields", func() {
			original := &PreviewEnvironmentStatus{
				Phase:     "Pending",
				URL:       "",
				Namespace: "preview-test",
			}

			copied := original.DeepCopy()
			Expect(copied).NotTo(BeIdenticalTo(original))
			Expect(copied.Phase).To(Equal(original.Phase))
			Expect(copied.Conditions).To(BeNil())
			Expect(copied.Services).To(BeNil())
			Expect(copied.CostEstimate).To(BeNil())
			Expect(copied.CreatedAt).To(BeNil())
			Expect(copied.ExpiresAt).To(BeNil())
			Expect(copied.LastSyncedAt).To(BeNil())
		})

		It("deep copies PreviewEnvironmentStatus with Conditions", func() {
			original := &PreviewEnvironmentStatus{
				Phase: "Ready",
				Conditions: []metav1.Condition{
					{
						Type:   "Available",
						Status: metav1.ConditionTrue,
						Reason: "ResourcesReady",
					},
				},
			}

			copied := original.DeepCopy()
			Expect(copied).NotTo(BeIdenticalTo(original))
			Expect(copied.Conditions).To(HaveLen(1))
			Expect(copied.Conditions[0].Type).To(Equal(original.Conditions[0].Type))
		})

		It("deep copies PreviewEnvironmentStatus correctly", func() {
			now := metav1.Now()
			original := &PreviewEnvironmentStatus{
				Phase:     "Ready",
				URL:       "https://example.com",
				Namespace: "preview-123",
				Services: []ServiceStatus{
					{Name: "api", Ready: true},
				},
				CostEstimate: &CostEstimate{
					Currency:   "USD",
					HourlyCost: "0.05",
				},
				CreatedAt:          &now,
				ExpiresAt:          &now,
				LastSyncedAt:       &now,
				ObservedGeneration: 1,
			}

			copied := original.DeepCopy()
			Expect(copied).NotTo(BeIdenticalTo(original))
			Expect(copied.Phase).To(Equal(original.Phase))
			Expect(copied.CostEstimate).NotTo(BeIdenticalTo(original.CostEstimate))
			Expect(copied.CostEstimate.Currency).To(Equal(original.CostEstimate.Currency))
		})

		It("returns nil when copying nil PreviewEnvironment", func() {
			var original *PreviewEnvironment
			copied := original.DeepCopy()
			Expect(copied).To(BeNil())
		})

		It("returns nil from DeepCopyObject when PreviewEnvironment is nil", func() {
			var original *PreviewEnvironment
			copied := original.DeepCopyObject()
			Expect(copied).To(BeNil())
		})

		It("deep copies PreviewEnvironment correctly", func() {
			original := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   123,
					HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
				},
				Status: PreviewEnvironmentStatus{
					Phase: "Ready",
				},
			}

			copied := original.DeepCopy()
			Expect(copied).NotTo(BeIdenticalTo(original))
			Expect(copied.Name).To(Equal(original.Name))
			Expect(copied.Spec.Repository).To(Equal(original.Spec.Repository))
			Expect(copied.Status.Phase).To(Equal(original.Status.Phase))
		})

		It("deep copies PreviewEnvironment as Object", func() {
			original := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-object",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   123,
					HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
				},
			}

			copied := original.DeepCopyObject()
			Expect(copied).NotTo(BeIdenticalTo(original))

			// Verify it's a PreviewEnvironment type
			copiedPE, ok := copied.(*PreviewEnvironment)
			Expect(ok).To(BeTrue())
			Expect(copiedPE.Name).To(Equal(original.Name))
		})

		It("returns nil when copying nil PreviewEnvironmentList", func() {
			var original *PreviewEnvironmentList
			copied := original.DeepCopy()
			Expect(copied).To(BeNil())
		})

		It("returns nil from DeepCopyObject when PreviewEnvironmentList is nil", func() {
			var original *PreviewEnvironmentList
			copied := original.DeepCopyObject()
			Expect(copied).To(BeNil())
		})

		It("deep copies PreviewEnvironmentList correctly", func() {
			original := &PreviewEnvironmentList{
				Items: []PreviewEnvironment{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test1",
							Namespace: "default",
						},
						Spec: PreviewEnvironmentSpec{
							Repository: "owner/repo",
							PRNumber:   123,
							HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test2",
							Namespace: "default",
						},
						Spec: PreviewEnvironmentSpec{
							Repository: "owner/repo2",
							PRNumber:   456,
							HeadSHA:    "abcdef1234567890abcdef1234567890abcdef12",
						},
					},
				},
			}

			copied := original.DeepCopy()
			Expect(copied).NotTo(BeIdenticalTo(original))
			Expect(copied.Items).To(HaveLen(2))
			Expect(copied.Items[0].Name).To(Equal(original.Items[0].Name))
			Expect(copied.Items[1].Spec.PRNumber).To(Equal(original.Items[1].Spec.PRNumber))

			// Modify copy, original should be unchanged
			copied.Items = append(copied.Items, PreviewEnvironment{})
			Expect(original.Items).To(HaveLen(2))
		})

		It("deep copies PreviewEnvironmentList as Object", func() {
			original := &PreviewEnvironmentList{
				Items: []PreviewEnvironment{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "default",
						},
					},
				},
			}

			copied := original.DeepCopyObject()
			Expect(copied).NotTo(BeIdenticalTo(original))

			// Verify it's a PreviewEnvironmentList type
			copiedList, ok := copied.(*PreviewEnvironmentList)
			Expect(ok).To(BeTrue())
			Expect(copiedList.Items).To(HaveLen(1))
		})
	})
})
