// Copyright 2025 The Previewd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

// PullRequestEvent represents a GitHub pull_request webhook event
type PullRequestEvent struct {
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
	Action      string      `json:"action"`
	Number      int         `json:"number"`
}

// PullRequest contains PR metadata
type PullRequest struct {
	Head  Ref    `json:"head"`
	Base  Ref    `json:"base"`
	Title string `json:"title"`
	State string `json:"state"`
}

// Ref represents a git reference (branch)
type Ref struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

// Repository contains repository metadata
type Repository struct {
	FullName string `json:"full_name"`
	Name     string `json:"name"`
	Owner    Owner  `json:"owner"`
}

// Owner represents the repository owner
type Owner struct {
	Login string `json:"login"`
}
