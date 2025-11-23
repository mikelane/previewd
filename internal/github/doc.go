// MIT License
//
// Copyright (c) 2025 Mike Lane
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package github provides GitHub API integration for Previewd.
//
// This package implements a client for interacting with the GitHub API to fetch
// pull request metadata and update commit statuses.
//
// Key features:
//   - Fetch pull request details (title, author, SHA, branches)
//   - Update commit status with preview environment information
//   - Retry logic with exponential backoff
//   - Rate limit handling
//   - Error handling and logging
//
// Authentication:
//
// The client requires a GitHub personal access token with the following scopes:
//   - repo (for accessing private repositories)
//   - repo:status (for updating commit status)
//
// Example usage:
//
//	client := github.NewClient(token)
//
//	// Fetch pull request metadata
//	pr, err := client.GetPullRequest(ctx, "owner", "repo", 123)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("PR #%d: %s\n", pr.Number, pr.Title)
//
//	// Update commit status
//	status := &github.CommitStatus{
//	    State:       "success",
//	    TargetURL:   "https://pr-123.preview.example.com",
//	    Description: "Preview environment ready",
//	    Context:     "previewd",
//	}
//	err = client.UpdateCommitStatus(ctx, "owner", "repo", pr.SHA, status)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Rate Limiting:
//
// The GitHub API has rate limits:
//   - 5,000 requests per hour for authenticated requests
//   - 60 requests per hour for unauthenticated requests
//
// The client automatically handles rate limit errors by waiting and retrying.
//
// Retry Logic:
//
// Failed requests are retried with exponential backoff:
//   - Initial backoff: 1 second
//   - Maximum backoff: 60 seconds
//   - Maximum retries: 3
//   - Backoff factor: 2.0
//
// Retries are performed for transient errors (network issues, rate limits, 5xx errors).
// Client errors (4xx except 429) are not retried.
package github
