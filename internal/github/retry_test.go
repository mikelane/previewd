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
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-github/v66/github"
)

// TestRetryWithBackoff tests the exponential backoff retry mechanism
//
//nolint:gocyclo // Complex test with multiple test cases
//nolint:gocyclo // Complex test with multiple test cases
func TestRetryWithBackoff(t *testing.T) {
	//nolint:govet,staticcheck // Field alignment not critical for test struct
	tests := []struct {
		name         string
		maxRetries   int
		statusCodes  []int
		wantAttempts int
		wantError    bool
		minTotalTime time.Duration
		maxTotalTime time.Duration
	}{
		{
			name:         "Succeeds on first attempt",
			maxRetries:   3,
			statusCodes:  []int{http.StatusOK},
			wantAttempts: 1,
			wantError:    false,
			minTotalTime: 0,
			maxTotalTime: 100 * time.Millisecond,
		},
		{
			name:         "Retries on 429 and succeeds",
			maxRetries:   3,
			statusCodes:  []int{http.StatusTooManyRequests, http.StatusOK},
			wantAttempts: 2,
			wantError:    false,
			minTotalTime: 80 * time.Millisecond, // Allow for jitter (100ms - 20%)
			maxTotalTime: 500 * time.Millisecond,
		},
		{
			name:         "Retries on 502 and succeeds",
			maxRetries:   3,
			statusCodes:  []int{http.StatusBadGateway, http.StatusOK},
			wantAttempts: 2,
			wantError:    false,
			minTotalTime: 80 * time.Millisecond, // Allow for jitter (100ms - 20%)
			maxTotalTime: 500 * time.Millisecond,
		},
		{
			name:         "Exhausts retries on persistent errors",
			maxRetries:   2,
			statusCodes:  []int{http.StatusTooManyRequests, http.StatusTooManyRequests, http.StatusTooManyRequests},
			wantAttempts: 3, // Initial + 2 retries
			wantError:    true,
			minTotalTime: 240 * time.Millisecond, // Allow for jitter (~100ms + ~200ms - 20%)
			maxTotalTime: 1000 * time.Millisecond,
		},
		{
			name:         "Does not retry on 404",
			maxRetries:   3,
			statusCodes:  []int{http.StatusNotFound},
			wantAttempts: 1,
			wantError:    true,
			minTotalTime: 0,
			maxTotalTime: 100 * time.Millisecond,
		},
		{
			name:         "Does not retry on 401",
			maxRetries:   3,
			statusCodes:  []int{http.StatusUnauthorized},
			wantAttempts: 1,
			wantError:    true,
			minTotalTime: 0,
			maxTotalTime: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attempts int32
			statusIndex := 0

			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&attempts, 1)

				if statusIndex < len(tt.statusCodes) {
					statusCode := tt.statusCodes[statusIndex]
					statusIndex++

					w.WriteHeader(statusCode)
					//nolint:staticcheck // QF1003: if-else is idiomatic for test
					if statusCode == http.StatusTooManyRequests {
						w.Header().Set("X-RateLimit-Remaining", "0")
						w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(1*time.Hour).Unix()))
						w.Write([]byte(`{"message":"API rate limit exceeded"}`)) //nolint:errcheck,gosec
					} else if statusCode == http.StatusBadGateway {
						w.Write([]byte(`{"message":"Bad Gateway"}`)) //nolint:errcheck,gosec
					} else if statusCode == http.StatusNotFound {
						w.Write([]byte(`{"message":"Not Found"}`)) //nolint:errcheck,gosec
					} else if statusCode == http.StatusUnauthorized {
						w.Write([]byte(`{"message":"Bad credentials"}`)) //nolint:errcheck,gosec
					}
				} else {
					w.WriteHeader(http.StatusOK)
				}
			}))
			defer server.Close()

			// Create client with retry config
			config := &RetryConfig{
				MaxRetries:     tt.maxRetries,
				InitialBackoff: 100 * time.Millisecond,
				MaxBackoff:     1 * time.Second,
				BackoffFactor:  2.0,
			}

			client := &githubClient{
				retryConfig: config,
			}

			start := time.Now()

			// Execute operation with retry
			err := client.executeWithRetry(context.Background(), func() error {
				resp, err := http.Get(server.URL) //nolint:noctx
				if err != nil {
					return err
				}
				defer resp.Body.Close() //nolint:errcheck,gosec

				if resp.StatusCode != http.StatusOK {
					// Simulate GitHub error response for retryable errors
					if resp.StatusCode == http.StatusTooManyRequests ||
						resp.StatusCode == http.StatusBadGateway ||
						resp.StatusCode == http.StatusServiceUnavailable ||
						resp.StatusCode == http.StatusGatewayTimeout {
						return &github.ErrorResponse{
							Response: resp,
							Message:  fmt.Sprintf("Request failed with status: %d", resp.StatusCode),
						}
					}
					return fmt.Errorf("request failed with status: %d", resp.StatusCode)
				}
				return nil
			})

			elapsed := time.Since(start)
			actualAttempts := int(atomic.LoadInt32(&attempts))

			// Verify results
			if tt.wantError && err == nil {
				t.Errorf("executeWithRetry() expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("executeWithRetry() unexpected error: %v", err)
			}
			if actualAttempts != tt.wantAttempts {
				t.Errorf("executeWithRetry() made %d attempts, want %d", actualAttempts, tt.wantAttempts)
			}
			if elapsed < tt.minTotalTime {
				t.Errorf("executeWithRetry() took %v, want at least %v", elapsed, tt.minTotalTime)
			}
			if elapsed > tt.maxTotalTime {
				t.Errorf("executeWithRetry() took %v, want at most %v", elapsed, tt.maxTotalTime)
			}
		})
	}
}

// TestRateLimitHandling tests GitHub API rate limit detection and handling
func TestRateLimitHandling(t *testing.T) {
	//nolint:govet,staticcheck // Field alignment not critical for test struct
	tests := []struct {
		name            string
		remainingHeader string
		resetHeader     string
		statusCode      int
		wantWaitTime    time.Duration
		wantRateLimited bool
	}{
		{
			name:            "Detects rate limit from headers",
			remainingHeader: "0",
			resetHeader:     fmt.Sprintf("%d", time.Now().Add(5*time.Second).Unix()),
			statusCode:      http.StatusForbidden,
			wantWaitTime:    5 * time.Second,
			wantRateLimited: true,
		},
		{
			name:            "No rate limit when remaining > 0",
			remainingHeader: "100",
			resetHeader:     fmt.Sprintf("%d", time.Now().Add(1*time.Hour).Unix()),
			statusCode:      http.StatusOK,
			wantWaitTime:    0,
			wantRateLimited: false,
		},
		{
			name:            "Handles secondary rate limit",
			remainingHeader: "",
			resetHeader:     "",
			statusCode:      http.StatusForbidden,
			wantWaitTime:    60 * time.Second, // Default wait for secondary rate limit
			wantRateLimited: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.remainingHeader != "" {
					w.Header().Set("X-RateLimit-Remaining", tt.remainingHeader)
				}
				if tt.resetHeader != "" {
					w.Header().Set("X-RateLimit-Reset", tt.resetHeader)
				}
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusForbidden {
					w.Write([]byte(`{"message":"API rate limit exceeded"}`)) //nolint:errcheck,gosec
				}
			}))
			defer server.Close()

			// Create client
			client := &githubClient{
				retryConfig: &RetryConfig{
					MaxRetries:     3,
					InitialBackoff: 100 * time.Millisecond,
					MaxBackoff:     30 * time.Second,
					BackoffFactor:  2.0,
				},
			}

			// Make request and check rate limit
			resp, err := http.Get(server.URL) //nolint:noctx
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close() //nolint:errcheck,gosec

			rateLimited, waitTime := client.checkRateLimit(resp)

			// Verify results
			if rateLimited != tt.wantRateLimited {
				t.Errorf("checkRateLimit() rateLimited = %v, want %v", rateLimited, tt.wantRateLimited)
			}

			// Allow some tolerance for time calculations
			tolerance := 2 * time.Second
			if tt.wantWaitTime > 0 {
				if waitTime < tt.wantWaitTime-tolerance || waitTime > tt.wantWaitTime+tolerance {
					t.Errorf("checkRateLimit() waitTime = %v, want ~%v", waitTime, tt.wantWaitTime)
				}
			} else if waitTime != 0 {
				t.Errorf("checkRateLimit() waitTime = %v, want 0", waitTime)
			}
		})
	}
}

// TestContextCancellation tests that retries respect context cancellation
func TestContextCancellation(t *testing.T) {
	var attempts int32

	// Create test server that always returns 429
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"message":"API rate limit exceeded"}`)) //nolint:errcheck,gosec
	}))
	defer server.Close()

	// Create client with retry config
	config := &RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		BackoffFactor:  2.0,
	}

	client := &githubClient{
		retryConfig: config,
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	start := time.Now()

	// Execute operation with retry
	err := client.executeWithRetry(ctx, func() error {
		resp, err := http.Get(server.URL) //nolint:noctx
		if err != nil {
			return err
		}
		defer resp.Body.Close() //nolint:errcheck,gosec

		if resp.StatusCode != http.StatusOK {
			// Simulate GitHub error response for retryable errors
			if resp.StatusCode == http.StatusTooManyRequests {
				return &github.ErrorResponse{
					Response: resp,
					Message:  fmt.Sprintf("Request failed with status: %d", resp.StatusCode),
				}
			}
			return fmt.Errorf("request failed with status: %d", resp.StatusCode)
		}
		return nil
	})

	elapsed := time.Since(start)
	actualAttempts := int(atomic.LoadInt32(&attempts))

	// Verify that context cancellation stops retries
	if err == nil {
		t.Errorf("executeWithRetry() expected error due to context cancellation, got nil")
	}
	if actualAttempts > 2 {
		t.Errorf("executeWithRetry() made %d attempts, expected <= 2 due to timeout", actualAttempts)
	}
	if elapsed > 250*time.Millisecond {
		t.Errorf("executeWithRetry() took %v, expected to be canceled within 250ms", elapsed)
	}
}

// TestJitterInBackoff tests that backoff includes jitter to avoid thundering herd
func TestJitterInBackoff(t *testing.T) {
	config := &RetryConfig{
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     10 * time.Second,
		BackoffFactor:  2.0,
	}

	client := &githubClient{
		retryConfig: config,
	}

	// Calculate backoff multiple times and ensure they're not all identical
	backoffs := make([]time.Duration, 10)
	for i := range backoffs {
		backoffs[i] = client.calculateBackoff(1) // Same retry count
	}

	// Check that not all backoffs are identical (jitter is working)
	allSame := true
	first := backoffs[0]
	for _, b := range backoffs[1:] {
		if b != first {
			allSame = false
			break
		}
	}

	if allSame {
		t.Errorf("calculateBackoff() returned identical values, jitter not working")
	}

	// Verify backoffs are within expected range (base ± 20% jitter)
	base := config.InitialBackoff * time.Duration(config.BackoffFactor)
	minBackoff := time.Duration(float64(base) * 0.8)
	maxBackoff := time.Duration(float64(base) * 1.2)

	for i, b := range backoffs {
		if b < minBackoff || b > maxBackoff {
			t.Errorf("backoff[%d] = %v, want between %v and %v", i, b, minBackoff, maxBackoff)
		}
	}
}
