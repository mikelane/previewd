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
	"time"
)

// Client interface defines the contract for interacting with GitHub API
type Client interface {
	// GetPullRequest retrieves metadata about a pull request
	GetPullRequest(ctx context.Context, owner, repo string, number int) (*PullRequest, error)
	// GetPRFiles retrieves the list of files changed in a pull request
	GetPRFiles(ctx context.Context, owner, repo string, number int) ([]*File, error)
	// UpdateCommitStatus updates the status of a commit
	UpdateCommitStatus(ctx context.Context, owner, repo, sha string, status *Status) error
}

// PullRequest represents GitHub pull request metadata
type PullRequest struct {
	Number      int
	Title       string
	Description string
	HeadSHA     string
	BaseBranch  string
	HeadBranch  string
	Author      string
	State       string // open, closed, merged
	Labels      []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// File represents a file changed in a pull request
type File struct {
	Filename  string
	Status    string // added, removed, modified, renamed
	Additions int
	Deletions int
	Changes   int
	Patch     string
}

// Status represents a commit status to be set on GitHub
type Status struct {
	State       StatusState // pending, success, error, failure
	TargetURL   string      // URL for more details
	Description string      // Short description of the status
	Context     string      // A unique name for this status check
}

// StatusState represents the state of a commit status
type StatusState string

const (
	// StatusStatePending indicates that the status is pending
	StatusStatePending StatusState = "pending"
	// StatusStateSuccess indicates that the status succeeded
	StatusStateSuccess StatusState = "success"
	// StatusStateError indicates that the status errored
	StatusStateError StatusState = "error"
	// StatusStateFailure indicates that the status failed
	StatusStateFailure StatusState = "failure"
)
