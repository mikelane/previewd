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

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// ValidateSignature verifies the HMAC-SHA256 signature of a GitHub webhook payload.
// It returns true if the signature is valid, false otherwise.
//
// The signature should be in the format "sha256=<hex-encoded-hmac>".
// Both the signature and secret must be non-empty for validation to succeed.
func ValidateSignature(payload []byte, signature string, secret string) bool {
	// Reject empty signature or secret
	if signature == "" || secret == "" {
		return false
	}

	// Ensure signature starts with "sha256="
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	// Extract hex-encoded HMAC from signature
	receivedMAC := strings.TrimPrefix(signature, "sha256=")

	// Compute expected HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison to prevent timing attacks
	return hmac.Equal([]byte(receivedMAC), []byte(expectedMAC))
}
