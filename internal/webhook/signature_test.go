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
	"testing"
)

// TestValidateSignature_ValidSignature verifies that a correctly signed payload is accepted
func TestValidateSignature_ValidSignature(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"action":"opened","number":123}`)
	// Precomputed HMAC-SHA256: echo -n '{"action":"opened","number":123}' | openssl dgst -sha256 -hmac 'test-secret'
	signature := "sha256=2c4854fbccd6d98cff684aedfef5f0edee3d89d30c1bae27c7e111bc1e82c282"

	valid := ValidateSignature(payload, signature, secret)

	if !valid {
		t.Error("ValidateSignature returns false for valid signature")
	}
}

// TestValidateSignature_InvalidSignature verifies that an incorrectly signed payload is rejected
func TestValidateSignature_InvalidSignature(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"action":"opened","number":123}`)
	signature := "sha256=0000000000000000000000000000000000000000000000000000000000000000"

	valid := ValidateSignature(payload, signature, secret)

	if valid {
		t.Error("ValidateSignature returns true for invalid signature")
	}
}

// TestValidateSignature_MissingSignature verifies that missing signature is rejected
func TestValidateSignature_MissingSignature(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"action":"opened","number":123}`)
	signature := ""

	valid := ValidateSignature(payload, signature, secret)

	if valid {
		t.Error("ValidateSignature returns true for missing signature")
	}
}

// TestValidateSignature_WrongAlgorithm verifies that SHA1 signatures are rejected
func TestValidateSignature_WrongAlgorithm(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"action":"opened","number":123}`)
	signature := "sha1=2c4854fbccd6d98cff684aedfef5f0edee3d89d30c1bae27"

	valid := ValidateSignature(payload, signature, secret)

	if valid {
		t.Error("ValidateSignature returns true for SHA1 signature (should require SHA256)")
	}
}

// TestValidateSignature_EmptySecret verifies that empty secret rejects all signatures
func TestValidateSignature_EmptySecret(t *testing.T) {
	secret := ""
	payload := []byte(`{"action":"opened","number":123}`)
	signature := "sha256=2c4854fbccd6d98cff684aedfef5f0edee3d89d30c1bae27c7e111bc1e82c282"

	valid := ValidateSignature(payload, signature, secret)

	if valid {
		t.Error("ValidateSignature returns true with empty secret")
	}
}
