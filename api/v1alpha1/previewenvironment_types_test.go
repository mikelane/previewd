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
					HeadSHA:    "abc1234567890def1234567890abc1234567890",
				},
			}

			Expect(preview.Spec.HeadSHA).To(Equal("abc1234567890def1234567890abc1234567890"))
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
		It("accepts BaseBranch when provided", func() {
			preview := &PreviewEnvironment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "basebranch-test",
					Namespace: "default",
				},
				Spec: PreviewEnvironmentSpec{
					Repository: "owner/repo",
					PRNumber:   123,
					HeadSHA:    "abc1234567890def1234567890abc1234567890",
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
					HeadSHA:    "abc1234567890def1234567890abc1234567890",
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
					HeadSHA:    "abc1234567890def1234567890abc1234567890",
					Services:   []string{"api", "web", "worker"},
				},
			}

			Expect(preview.Spec.Services).To(HaveLen(3))
			Expect(preview.Spec.Services).To(ContainElements("api", "web", "worker"))
		})
	})
})
