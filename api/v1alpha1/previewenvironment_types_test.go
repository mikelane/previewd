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
	})
})
