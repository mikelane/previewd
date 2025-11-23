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

// Package cleanup provides automatic TTL-based cleanup of expired preview environments.
//
// This package implements a background scheduler that periodically checks PreviewEnvironment
// resources for expiration and automatically deletes them to prevent resource waste.
//
// Key features:
//   - Periodic cleanup based on configurable interval (default: 5 minutes)
//   - Respects TTL (time-to-live) configured in PreviewEnvironment spec
//   - Honors "do-not-expire" label override for long-running environments
//   - Graceful shutdown via context cancellation
//   - Emits Kubernetes events when environments are deleted
//
// TTL Calculation:
//
// Each PreviewEnvironment has an expiration time calculated as:
//
//	expiresAt = createdAt + spec.ttl
//
// The default TTL is 4 hours, with a maximum of 7 days.
//
// Exempting Environments from Cleanup:
//
// To prevent a preview environment from being automatically deleted, add the
// "preview.previewd.io/do-not-expire" label with value "true":
//
//	apiVersion: preview.previewd.io/v1alpha1
//	kind: PreviewEnvironment
//	metadata:
//	  name: pr-123
//	  labels:
//	    preview.previewd.io/do-not-expire: "true"
//	spec:
//	  # ... spec fields ...
//
// Example usage:
//
//	scheduler := cleanup.NewScheduler(
//		k8sClient,
//		5*time.Minute, // Check every 5 minutes
//	)
//	if err := scheduler.Start(ctx); err != nil {
//		log.Fatal(err)
//	}
package cleanup
