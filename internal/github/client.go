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

package github

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/google/go-github/v66/github"
)

// RetryConfig defines the retry behavior for API calls
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
}

// githubClient implements the Client interface using go-github
type githubClient struct {
	client      *github.Client
	retryConfig *RetryConfig
}

// NewClient creates a new GitHub client with the provided token
func NewClient(token string) (Client, error) {
	var httpClient *http.Client
	if token != "" {
		httpClient = github.NewClient(nil).Client()
		httpClient.Transport = &github.BasicAuthTransport{
			Username: "token",
			Password: token,
		}
	}

	return &githubClient{
		client: github.NewClient(httpClient),
		retryConfig: &RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     30 * time.Second,
			BackoffFactor:  2.0,
		},
	}, nil
}

// GetPullRequest retrieves metadata about a pull request
func (c *githubClient) GetPullRequest(ctx context.Context, owner, repo string, number int) (*PullRequest, error) {
	var pr *github.PullRequest
	var err error

	err = c.executeWithRetry(ctx, func() error {
		pr, _, err = c.client.PullRequests.Get(ctx, owner, repo, number)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	return c.convertPullRequest(pr), nil
}

// GetPRFiles retrieves the list of files changed in a pull request
func (c *githubClient) GetPRFiles(ctx context.Context, owner, repo string, number int) ([]*File, error) {
	allFiles := []*File{} // Initialize as empty slice, not nil
	opts := &github.ListOptions{
		PerPage: 100,
	}

	for {
		var files []*github.CommitFile
		var resp *github.Response
		var err error

		err = c.executeWithRetry(ctx, func() error {
			files, resp, err = c.client.PullRequests.ListFiles(ctx, owner, repo, number, opts)
			return err
		})

		if err != nil {
			return nil, fmt.Errorf("failed to list PR files: %w", err)
		}

		for _, file := range files {
			allFiles = append(allFiles, c.convertFile(file))
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allFiles, nil
}

// UpdateCommitStatus updates the status of a commit
func (c *githubClient) UpdateCommitStatus(ctx context.Context, owner, repo, sha string, status *Status) error {
	repoStatus := &github.RepoStatus{
		State:       github.String(string(status.State)),
		TargetURL:   github.String(status.TargetURL),
		Description: github.String(status.Description),
		Context:     github.String(status.Context),
	}

	err := c.executeWithRetry(ctx, func() error {
		_, _, err := c.client.Repositories.CreateStatus(ctx, owner, repo, sha, repoStatus)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to update commit status: %w", err)
	}

	return nil
}

// executeWithRetry executes an operation with exponential backoff retry
func (c *githubClient) executeWithRetry(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		// Check if context is cancelled before attempting
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		lastErr = operation()

		// Success
		if lastErr == nil {
			return nil
		}

		// Check if error is retryable
		if !c.isRetryableError(lastErr) {
			return lastErr
		}

		// Don't retry if we've exhausted attempts
		if attempt == c.retryConfig.MaxRetries {
			break
		}

		// Calculate backoff with jitter
		backoff := c.calculateBackoff(attempt)

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			// Continue to next retry
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", c.retryConfig.MaxRetries, lastErr)
}

// isRetryableError determines if an error should trigger a retry
func (c *githubClient) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for GitHub API errors
	if ghErr, ok := err.(*github.ErrorResponse); ok {
		switch ghErr.Response.StatusCode {
		case http.StatusTooManyRequests,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout:
			return true
		case http.StatusForbidden:
			// Check if it's a rate limit error
			if ghErr.Message == "API rate limit exceeded" {
				return true
			}
		}
	}

	return false
}

// calculateBackoff calculates the backoff duration for a retry attempt
func (c *githubClient) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff with jitter
	multiplier := 1 << uint(attempt) // 2^attempt
	base := float64(c.retryConfig.InitialBackoff) * float64(multiplier)

	// Add jitter (±20%)
	jitter := (rand.Float64() * 0.4) - 0.2 // -0.2 to +0.2
	backoff := time.Duration(base * (1 + jitter))

	// Cap at max backoff
	if backoff > c.retryConfig.MaxBackoff {
		backoff = c.retryConfig.MaxBackoff
	}

	return backoff
}

// checkRateLimit checks response headers for rate limit information
func (c *githubClient) checkRateLimit(resp *http.Response) (bool, time.Duration) {
	if resp == nil {
		return false, 0
	}

	// Check primary rate limit
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	if remaining != "" {
		if rem, err := strconv.Atoi(remaining); err == nil && rem == 0 {
			// Rate limited - calculate wait time
			resetStr := resp.Header.Get("X-RateLimit-Reset")
			if resetStr != "" {
				if resetTime, err := strconv.ParseInt(resetStr, 10, 64); err == nil {
					waitTime := time.Until(time.Unix(resetTime, 0))
					if waitTime > 0 {
						return true, waitTime
					}
				}
			}
		}
	}

	// Check for secondary rate limit (403 without rate limit headers)
	if resp.StatusCode == http.StatusForbidden {
		// Default wait for secondary rate limit
		return true, 60 * time.Second
	}

	return false, 0
}

// convertPullRequest converts a GitHub PR to our domain model
func (c *githubClient) convertPullRequest(pr *github.PullRequest) *PullRequest {
	if pr == nil {
		return nil
	}

	result := &PullRequest{
		Number:      pr.GetNumber(),
		Title:       pr.GetTitle(),
		Description: pr.GetBody(),
		State:       pr.GetState(),
		CreatedAt:   pr.GetCreatedAt().Time,
		UpdatedAt:   pr.GetUpdatedAt().Time,
	}

	if pr.Head != nil {
		result.HeadSHA = pr.Head.GetSHA()
		result.HeadBranch = pr.Head.GetRef()
	}

	if pr.Base != nil {
		result.BaseBranch = pr.Base.GetRef()
	}

	if pr.User != nil {
		result.Author = pr.User.GetLogin()
	}

	// Convert labels
	for _, label := range pr.Labels {
		if label != nil {
			result.Labels = append(result.Labels, label.GetName())
		}
	}

	return result
}

// convertFile converts a GitHub CommitFile to our domain model
func (c *githubClient) convertFile(file *github.CommitFile) *File {
	if file == nil {
		return nil
	}

	return &File{
		Filename:  file.GetFilename(),
		Status:    file.GetStatus(),
		Additions: file.GetAdditions(),
		Deletions: file.GetDeletions(),
		Changes:   file.GetChanges(),
		Patch:     file.GetPatch(),
	}
}
