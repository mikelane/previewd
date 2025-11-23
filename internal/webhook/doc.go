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

// Package webhook provides GitHub webhook handling for Previewd.
//
// This package implements an HTTP server that receives GitHub pull request webhook
// events and translates them into PreviewEnvironment Kubernetes resources.
//
// Key features:
//   - Validates GitHub webhook signatures using HMAC-SHA256
//   - Handles pull_request events (opened, synchronize, closed, reopened)
//   - Creates, updates, and deletes PreviewEnvironment resources
//   - Provides per-repository rate limiting
//   - Health check and readiness endpoints
//
// Webhook Security:
//
// All webhook requests must include a valid X-Hub-Signature-256 header containing
// an HMAC-SHA256 signature computed with the webhook secret. Requests with invalid
// or missing signatures are rejected with HTTP 401.
//
// Event Handling:
//
// The webhook server processes the following pull_request actions:
//   - opened: Creates a new PreviewEnvironment
//   - synchronize: Updates the PreviewEnvironment with new head SHA
//   - reopened: Recreates the PreviewEnvironment if deleted
//   - closed: Deletes the PreviewEnvironment
//
// Rate Limiting:
//
// Requests are rate-limited per repository using a token bucket algorithm.
// The default limit is 10 requests per second per repository. Requests
// exceeding the limit receive HTTP 429 Too Many Requests.
//
// Example usage:
//
//	server := webhook.NewServer(
//		k8sClient,
//		"webhook-secret",
//		8080,
//	)
//	if err := server.Start(ctx); err != nil {
//		log.Fatal(err)
//	}
package webhook
