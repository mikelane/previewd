/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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

	Context("TTL field", func() {
		It("defaults to 4h when not provided", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ttl-default-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   123,
					HeadSHA:    "abc1234567890def1234567890abc12345678901",
				},
			}

			// Create the resource
			err := k8sClient.Create(ctx, preview)
			Expect(err).NotTo(HaveOccurred())

			// Fetch it back
			fetched := &PreviewEnvironment{}
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(preview), fetched)
			Expect(err).NotTo(HaveOccurred())

			// Verify TTL is set to default
			Expect(fetched.Spec.TTL).To(Equal("4h"))

			// Cleanup
			err = k8sClient.Delete(ctx, preview)
			Expect(err).NotTo(HaveOccurred())
		})

		It("accepts custom TTL when provided", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ttl-custom-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   123,
					HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
					TTL:        "8h",
				},
			}

			err := k8sClient.Create(ctx, preview)
			Expect(err).NotTo(HaveOccurred())

			// Verify TTL is the custom value
			Expect(preview.Spec.TTL).To(Equal("8h"))

			// Cleanup
			err = k8sClient.Delete(ctx, preview)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Optional fields", func() {
		It("accepts BaseBranch when provided", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "basebranch-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   123,
					HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
					BaseBranch: "main",
				},
			}

			Expect(preview.Spec.BaseBranch).To(Equal("main"))
		})

		It("accepts HeadBranch when provided", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "headbranch-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   123,
					HeadSHA:    "1234567890abcdef1234567890abcdef12345678",
					HeadBranch: "feature/test",
				},
			}

			Expect(preview.Spec.HeadBranch).To(Equal("feature/test"))
		})

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
			Expect(fetched.Spec.TTL).To(Equal("4h")) // Default

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
})
