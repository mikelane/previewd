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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-github/v66/github"
)

// TestNewClient tests the creation of a new GitHub client
func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		wantError bool
	}{
		{
			name:      "Valid token creates client",
			token:     "github_pat_test123",
			wantError: false,
		},
		{
			name:      "Empty token creates client",
			token:     "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.token)
			if tt.wantError && err == nil {
				t.Errorf("NewClient() expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("NewClient() unexpected error: %v", err)
			}
			if !tt.wantError && client == nil {
				t.Errorf("NewClient() returned nil client")
			}
		})
	}
}

// TestGetPullRequest tests fetching pull request metadata
func TestGetPullRequest(t *testing.T) {
	tests := []struct {
		name       string
		owner      string
		repo       string
		number     int
		mockPR     *github.PullRequest
		mockError  error
		wantPR     *PullRequest
		wantError  bool
		statusCode int
	}{
		{
			name:   "Successfully fetches pull request",
			owner:  "mikelane",
			repo:   "previewd",
			number: 42,
			mockPR: &github.PullRequest{
				Number: github.Int(42),
				Title:  github.String("feat: add awesome feature"),
				Body:   github.String("This PR adds an awesome feature"),
				Head: &github.PullRequestBranch{
					SHA: github.String("abc123"),
					Ref: github.String("feature-branch"),
				},
				Base: &github.PullRequestBranch{
					Ref: github.String("main"),
				},
				User: &github.User{
					Login: github.String("mikelane"),
				},
				State:     github.String("open"),
				CreatedAt: &github.Timestamp{Time: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				UpdatedAt: &github.Timestamp{Time: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
				Labels: []*github.Label{
					{Name: github.String("feature")},
					{Name: github.String("enhancement")},
				},
			},
			wantPR: &PullRequest{
				Number:      42,
				Title:       "feat: add awesome feature",
				Description: "This PR adds an awesome feature",
				HeadSHA:     "abc123",
				BaseBranch:  "main",
				HeadBranch:  "feature-branch",
				Author:      "mikelane",
				State:       "open",
				Labels:      []string{"feature", "enhancement"},
				CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			wantError:  false,
			statusCode: http.StatusOK,
		},
		{
			name:       "Handles not found error",
			owner:      "mikelane",
			repo:       "previewd",
			number:     999,
			mockPR:     nil,
			wantPR:     nil,
			wantError:  true,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Handles rate limit error",
			owner:      "mikelane",
			repo:       "previewd",
			number:     1,
			mockPR:     nil,
			wantPR:     nil,
			wantError:  true,
			statusCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/repos/%s/%s/pulls/%d", tt.owner, tt.repo, tt.number)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				if tt.statusCode != http.StatusOK {
					w.WriteHeader(tt.statusCode)
					if tt.statusCode == http.StatusForbidden {
						w.Write([]byte(`{"message":"API rate limit exceeded"}`))
					} else if tt.statusCode == http.StatusNotFound {
						w.Write([]byte(`{"message":"Not Found"}`))
					}
					return
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.mockPR)
			}))
			defer server.Close()

			// Create client with test server
			client := &githubClient{
				client: github.NewClient(nil),
				retryConfig: &RetryConfig{
					MaxRetries:     3,
					InitialBackoff: 100 * time.Millisecond,
					MaxBackoff:     30 * time.Second,
					BackoffFactor:  2.0,
				},
			}
			client.client.BaseURL, _ = client.client.BaseURL.Parse(server.URL + "/")

			// Execute test
			pr, err := client.GetPullRequest(context.Background(), tt.owner, tt.repo, tt.number)

			// Verify results
			if tt.wantError && err == nil {
				t.Errorf("GetPullRequest() expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("GetPullRequest() unexpected error: %v", err)
			}
			if !tt.wantError && pr == nil {
				t.Errorf("GetPullRequest() returned nil PR")
			}
			if tt.wantPR != nil && pr != nil {
				if pr.Number != tt.wantPR.Number {
					t.Errorf("PR.Number = %d, want %d", pr.Number, tt.wantPR.Number)
				}
				if pr.Title != tt.wantPR.Title {
					t.Errorf("PR.Title = %s, want %s", pr.Title, tt.wantPR.Title)
				}
				if pr.HeadSHA != tt.wantPR.HeadSHA {
					t.Errorf("PR.HeadSHA = %s, want %s", pr.HeadSHA, tt.wantPR.HeadSHA)
				}
				if pr.Author != tt.wantPR.Author {
					t.Errorf("PR.Author = %s, want %s", pr.Author, tt.wantPR.Author)
				}
				if len(pr.Labels) != len(tt.wantPR.Labels) {
					t.Errorf("PR.Labels length = %d, want %d", len(pr.Labels), len(tt.wantPR.Labels))
				}
			}
		})
	}
}

// TestGetPRFiles tests fetching PR file changes
func TestGetPRFiles(t *testing.T) {
	tests := []struct {
		name       string
		owner      string
		repo       string
		number     int
		mockFiles  []*github.CommitFile
		wantFiles  []*File
		wantError  bool
		statusCode int
		page       int
		perPage    int
		hasMore    bool
	}{
		{
			name:   "Successfully fetches single page of files",
			owner:  "mikelane",
			repo:   "previewd",
			number: 42,
			mockFiles: []*github.CommitFile{
				{
					Filename:  github.String("internal/github/client.go"),
					Status:    github.String("added"),
					Additions: github.Int(150),
					Deletions: github.Int(0),
					Changes:   github.Int(150),
					Patch:     github.String("@@ -0,0 +1,150 @@\n+package github\n+..."),
				},
				{
					Filename:  github.String("internal/github/client_test.go"),
					Status:    github.String("added"),
					Additions: github.Int(200),
					Deletions: github.Int(0),
					Changes:   github.Int(200),
					Patch:     github.String("@@ -0,0 +1,200 @@\n+package github\n+..."),
				},
			},
			wantFiles: []*File{
				{
					Filename:  "internal/github/client.go",
					Status:    "added",
					Additions: 150,
					Deletions: 0,
					Changes:   150,
					Patch:     "@@ -0,0 +1,150 @@\n+package github\n+...",
				},
				{
					Filename:  "internal/github/client_test.go",
					Status:    "added",
					Additions: 200,
					Deletions: 0,
					Changes:   200,
					Patch:     "@@ -0,0 +1,200 @@\n+package github\n+...",
				},
			},
			wantError:  false,
			statusCode: http.StatusOK,
			page:       1,
			perPage:    100,
			hasMore:    false,
		},
		{
			name:       "Handles empty file list",
			owner:      "mikelane",
			repo:       "previewd",
			number:     43,
			mockFiles:  []*github.CommitFile{},
			wantFiles:  []*File{},
			wantError:  false,
			statusCode: http.StatusOK,
			page:       1,
			perPage:    100,
			hasMore:    false,
		},
		{
			name:       "Handles error response",
			owner:      "mikelane",
			repo:       "previewd",
			number:     999,
			mockFiles:  nil,
			wantFiles:  nil,
			wantError:  true,
			statusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				expectedPath := fmt.Sprintf("/repos/%s/%s/pulls/%d/files", tt.owner, tt.repo, tt.number)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				if tt.statusCode != http.StatusOK {
					w.WriteHeader(tt.statusCode)
					if tt.statusCode == http.StatusNotFound {
						w.Write([]byte(`{"message":"Not Found"}`))
					}
					return
				}

				// Add Link header for pagination if there are more pages
				if tt.hasMore && callCount == 1 {
					w.Header().Set("Link", fmt.Sprintf(`<http://api.github.com/repos/%s/%s/pulls/%d/files?page=2>; rel="next"`, tt.owner, tt.repo, tt.number))
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.mockFiles)
			}))
			defer server.Close()

			// Create client with test server
			client := &githubClient{
				client: github.NewClient(nil),
				retryConfig: &RetryConfig{
					MaxRetries:     3,
					InitialBackoff: 100 * time.Millisecond,
					MaxBackoff:     30 * time.Second,
					BackoffFactor:  2.0,
				},
			}
			client.client.BaseURL, _ = client.client.BaseURL.Parse(server.URL + "/")

			// Execute test
			files, err := client.GetPRFiles(context.Background(), tt.owner, tt.repo, tt.number)

			// Verify results
			if tt.wantError && err == nil {
				t.Errorf("GetPRFiles() expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("GetPRFiles() unexpected error: %v", err)
			}
			if !tt.wantError && files == nil {
				t.Errorf("GetPRFiles() returned nil files")
			}
			if tt.wantFiles != nil && files != nil {
				if len(files) != len(tt.wantFiles) {
					t.Errorf("GetPRFiles() returned %d files, want %d", len(files), len(tt.wantFiles))
				}
				for i, file := range files {
					if i >= len(tt.wantFiles) {
						break
					}
					want := tt.wantFiles[i]
					if file.Filename != want.Filename {
						t.Errorf("File[%d].Filename = %s, want %s", i, file.Filename, want.Filename)
					}
					if file.Status != want.Status {
						t.Errorf("File[%d].Status = %s, want %s", i, file.Status, want.Status)
					}
					if file.Additions != want.Additions {
						t.Errorf("File[%d].Additions = %d, want %d", i, file.Additions, want.Additions)
					}
				}
			}
		})
	}
}

// TestUpdateCommitStatus tests updating commit status
func TestUpdateCommitStatus(t *testing.T) {
	tests := []struct {
		name       string
		owner      string
		repo       string
		sha        string
		status     *Status
		wantError  bool
		statusCode int
	}{
		{
			name:  "Successfully updates commit status",
			owner: "mikelane",
			repo:  "previewd",
			sha:   "abc123def456",
			status: &Status{
				State:       StatusStatePending,
				TargetURL:   "https://preview.example.com/pr-42",
				Description: "Creating preview environment",
				Context:     "previewd/environment",
			},
			wantError:  false,
			statusCode: http.StatusCreated,
		},
		{
			name:  "Updates success status",
			owner: "mikelane",
			repo:  "previewd",
			sha:   "def456ghi789",
			status: &Status{
				State:       StatusStateSuccess,
				TargetURL:   "https://preview.example.com/pr-43",
				Description: "Preview environment ready",
				Context:     "previewd/environment",
			},
			wantError:  false,
			statusCode: http.StatusCreated,
		},
		{
			name:  "Handles unauthorized error",
			owner: "mikelane",
			repo:  "previewd",
			sha:   "invalid",
			status: &Status{
				State:       StatusStateError,
				Description: "Failed",
				Context:     "previewd/environment",
			},
			wantError:  true,
			statusCode: http.StatusUnauthorized,
		},
		{
			name:  "Handles not found error",
			owner: "mikelane",
			repo:  "previewd",
			sha:   "nonexistent",
			status: &Status{
				State:       StatusStateFailure,
				Description: "Failed",
				Context:     "previewd/environment",
			},
			wantError:  true,
			statusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/repos/%s/%s/statuses/%s", tt.owner, tt.repo, tt.sha)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				if r.Method != "POST" {
					t.Errorf("Expected method POST, got %s", r.Method)
				}

				// Verify request body
				var reqStatus github.RepoStatus
				if err := json.NewDecoder(r.Body).Decode(&reqStatus); err != nil {
					t.Errorf("Failed to decode request body: %v", err)
				}

				if tt.statusCode != http.StatusCreated {
					w.WriteHeader(tt.statusCode)
					if tt.statusCode == http.StatusUnauthorized {
						w.Write([]byte(`{"message":"Bad credentials"}`))
					} else if tt.statusCode == http.StatusNotFound {
						w.Write([]byte(`{"message":"Not Found"}`))
					}
					return
				}

				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"state":"` + string(tt.status.State) + `","context":"` + tt.status.Context + `"}`))
			}))
			defer server.Close()

			// Create client with test server
			client := &githubClient{
				client: github.NewClient(nil),
				retryConfig: &RetryConfig{
					MaxRetries:     3,
					InitialBackoff: 100 * time.Millisecond,
					MaxBackoff:     30 * time.Second,
					BackoffFactor:  2.0,
				},
			}
			client.client.BaseURL, _ = client.client.BaseURL.Parse(server.URL + "/")

			// Execute test
			err := client.UpdateCommitStatus(context.Background(), tt.owner, tt.repo, tt.sha, tt.status)

			// Verify results
			if tt.wantError && err == nil {
				t.Errorf("UpdateCommitStatus() expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("UpdateCommitStatus() unexpected error: %v", err)
			}
		})
	}
}

// TestGetPRFilesPagination tests pagination handling for large PRs
func TestGetPRFilesPagination(t *testing.T) {
	owner := "mikelane"
	repo := "previewd"
	number := 100
	pageCount := 0

	// Create test server that simulates pagination
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageCount++

		// Simulate 3 pages of results
		if pageCount == 1 {
			w.Header().Set("Link", fmt.Sprintf(`<http://api.github.com/repos/%s/%s/pulls/%d/files?page=2>; rel="next"`, owner, repo, number))
			files := []*github.CommitFile{
				{
					Filename:  github.String("file1.go"),
					Status:    github.String("added"),
					Additions: github.Int(10),
					Deletions: github.Int(0),
					Changes:   github.Int(10),
				},
			}
			json.NewEncoder(w).Encode(files)
		} else if pageCount == 2 {
			w.Header().Set("Link", fmt.Sprintf(`<http://api.github.com/repos/%s/%s/pulls/%d/files?page=3>; rel="next"`, owner, repo, number))
			files := []*github.CommitFile{
				{
					Filename:  github.String("file2.go"),
					Status:    github.String("modified"),
					Additions: github.Int(5),
					Deletions: github.Int(3),
					Changes:   github.Int(8),
				},
			}
			json.NewEncoder(w).Encode(files)
		} else if pageCount == 3 {
			// Last page, no Link header
			files := []*github.CommitFile{
				{
					Filename:  github.String("file3.go"),
					Status:    github.String("deleted"),
					Additions: github.Int(0),
					Deletions: github.Int(20),
					Changes:   github.Int(20),
				},
			}
			json.NewEncoder(w).Encode(files)
		} else {
			t.Errorf("Unexpected page request: %d", pageCount)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	// Create client with test server
	client := &githubClient{
		client: github.NewClient(nil),
		retryConfig: &RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     30 * time.Second,
			BackoffFactor:  2.0,
		},
	}
	client.client.BaseURL, _ = client.client.BaseURL.Parse(server.URL + "/")

	// Execute test
	files, err := client.GetPRFiles(context.Background(), owner, repo, number)

	// Verify results
	if err != nil {
		t.Errorf("GetPRFiles() unexpected error: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("GetPRFiles() returned %d files, want 3", len(files))
	}
	if pageCount != 3 {
		t.Errorf("GetPRFiles() made %d requests, want 3", pageCount)
	}

	// Verify file details
	expectedFiles := []struct {
		filename string
		status   string
	}{
		{"file1.go", "added"},
		{"file2.go", "modified"},
		{"file3.go", "deleted"},
	}

	for i, expected := range expectedFiles {
		if i >= len(files) {
			t.Errorf("Missing file at index %d", i)
			continue
		}
		if files[i].Filename != expected.filename {
			t.Errorf("File[%d].Filename = %s, want %s", i, files[i].Filename, expected.filename)
		}
		if files[i].Status != expected.status {
			t.Errorf("File[%d].Status = %s, want %s", i, files[i].Status, expected.status)
		}
	}
}
