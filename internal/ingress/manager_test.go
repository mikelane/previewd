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

package ingress

import (
	"context"
	"testing"

	previewv1alpha1 "github.com/mikelane/previewd/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewManager(t *testing.T) {
	scheme := runtime.NewScheme()
	c := fake.NewClientBuilder().WithScheme(scheme).Build()

	baseDomain := "preview.example.com"
	certIssuer := "letsencrypt-prod"

	m := NewManager(c, scheme, baseDomain, certIssuer)

	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	if m.client == nil {
		t.Error("Manager client is nil")
	}
	if m.scheme == nil {
		t.Error("Manager scheme is nil")
	}
	if m.baseDomain != baseDomain {
		t.Errorf("baseDomain = %v, want %v", m.baseDomain, baseDomain)
	}
	if m.certIssuer != certIssuer {
		t.Errorf("certIssuer = %v, want %v", m.certIssuer, certIssuer)
	}
}

// setupTestClient creates a fake Kubernetes client with necessary schemes
func setupTestClient(t *testing.T, namespace string) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := previewv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add preview scheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add core scheme: %v", err)
	}
	if err := networkingv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add networking scheme: %v", err)
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(ns).
		Build()
}

func TestManager_EnsureIngress_HostAndTLS(t *testing.T) {
	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-123",
			Namespace: "previewd-system",
			UID:       "test-uid-1",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   123,
			Repository: "owner/repo",
			Services:   []string{"auth", "api", "frontend"},
		},
	}
	namespace := "preview-pr-123-abc123"

	c := setupTestClient(t, namespace)
	m := NewManager(c, c.Scheme(), "preview.example.com", "letsencrypt-prod")
	err := m.EnsureIngress(context.Background(), preview, namespace)

	if err != nil {
		t.Fatalf("EnsureIngress() error = %v", err)
	}

	ingress := &networkingv1.Ingress{}
	err = c.Get(context.Background(), types.NamespacedName{
		Name:      "preview-ingress",
		Namespace: namespace,
	}, ingress)
	if err != nil {
		t.Fatalf("failed to get ingress: %v", err)
	}

	// Verify host
	expectedHost := "pr-123.preview.example.com"
	if len(ingress.Spec.Rules) == 0 {
		t.Fatal("ingress has no rules")
	}
	if ingress.Spec.Rules[0].Host != expectedHost {
		t.Errorf("host = %v, want %v", ingress.Spec.Rules[0].Host, expectedHost)
	}

	// Verify TLS
	if len(ingress.Spec.TLS) == 0 {
		t.Fatal("ingress has no TLS configuration")
	}
	if ingress.Spec.TLS[0].SecretName != "pr-123-tls" {
		t.Errorf("TLS secret = %v, want %v", ingress.Spec.TLS[0].SecretName, "pr-123-tls")
	}
	if len(ingress.Spec.TLS[0].Hosts) == 0 || ingress.Spec.TLS[0].Hosts[0] != expectedHost {
		t.Errorf("TLS host = %v, want %v", ingress.Spec.TLS[0].Hosts, []string{expectedHost})
	}
}

func TestManager_EnsureIngress_CertManagerAnnotation(t *testing.T) {
	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-456",
			Namespace: "previewd-system",
			UID:       "test-uid-2",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   456,
			Repository: "owner/repo",
			Services:   []string{"api"},
		},
	}
	namespace := "preview-pr-456-def456"

	c := setupTestClient(t, namespace)
	m := NewManager(c, c.Scheme(), "preview.example.com", "letsencrypt-prod")
	err := m.EnsureIngress(context.Background(), preview, namespace)

	if err != nil {
		t.Fatalf("EnsureIngress() error = %v", err)
	}

	ingress := &networkingv1.Ingress{}
	err = c.Get(context.Background(), types.NamespacedName{
		Name:      "preview-ingress",
		Namespace: namespace,
	}, ingress)
	if err != nil {
		t.Fatalf("failed to get ingress: %v", err)
	}

	// Verify cert-manager annotation
	certIssuerAnnotation := "cert-manager.io/cluster-issuer"
	if ingress.Annotations[certIssuerAnnotation] != "letsencrypt-prod" {
		t.Errorf("cert-manager annotation = %v, want %v",
			ingress.Annotations[certIssuerAnnotation], "letsencrypt-prod")
	}
}

func TestManager_EnsureIngress_ExternalDNSAnnotation(t *testing.T) {
	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-789",
			Namespace: "previewd-system",
			UID:       "test-uid-3",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   789,
			Repository: "owner/repo",
			Services:   []string{"frontend"},
		},
	}
	namespace := "preview-pr-789-ghi789"

	c := setupTestClient(t, namespace)
	m := NewManager(c, c.Scheme(), "preview.example.com", "letsencrypt-prod")
	err := m.EnsureIngress(context.Background(), preview, namespace)

	if err != nil {
		t.Fatalf("EnsureIngress() error = %v", err)
	}

	ingress := &networkingv1.Ingress{}
	err = c.Get(context.Background(), types.NamespacedName{
		Name:      "preview-ingress",
		Namespace: namespace,
	}, ingress)
	if err != nil {
		t.Fatalf("failed to get ingress: %v", err)
	}

	// Verify external-dns annotation
	expectedHost := "pr-789.preview.example.com"
	dnsAnnotation := "external-dns.alpha.kubernetes.io/hostname"
	if ingress.Annotations[dnsAnnotation] != expectedHost {
		t.Errorf("external-dns annotation = %v, want %v",
			ingress.Annotations[dnsAnnotation], expectedHost)
	}
}

func TestManager_EnsureIngress_PathBasedRouting(t *testing.T) {
	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-111",
			Namespace: "previewd-system",
			UID:       "test-uid-4",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   111,
			Repository: "owner/repo",
			Services:   []string{"auth", "api", "frontend"},
		},
	}
	namespace := "preview-pr-111-jkl111"

	c := setupTestClient(t, namespace)
	m := NewManager(c, c.Scheme(), "preview.example.com", "letsencrypt-prod")
	err := m.EnsureIngress(context.Background(), preview, namespace)

	if err != nil {
		t.Fatalf("EnsureIngress() error = %v", err)
	}

	ingress := &networkingv1.Ingress{}
	err = c.Get(context.Background(), types.NamespacedName{
		Name:      "preview-ingress",
		Namespace: namespace,
	}, ingress)
	if err != nil {
		t.Fatalf("failed to get ingress: %v", err)
	}

	// Verify path rules
	if len(ingress.Spec.Rules) == 0 {
		t.Fatal("ingress has no rules")
	}

	httpRules := ingress.Spec.Rules[0].HTTP
	if httpRules == nil {
		t.Fatal("ingress has no HTTP rules")
	}

	if len(httpRules.Paths) != 3 {
		t.Errorf("expected 3 path rules, got %d", len(httpRules.Paths))
	}

	// Verify paths exist
	expectedPaths := map[string]string{
		"/auth": "preview-pr-111-auth",
		"/api":  "preview-pr-111-api",
		"/":     "preview-pr-111-frontend",
	}

	for _, path := range httpRules.Paths {
		expectedService, ok := expectedPaths[path.Path]
		if !ok {
			t.Errorf("unexpected path: %s", path.Path)
			continue
		}

		if path.Backend.Service.Name != expectedService {
			t.Errorf("path %s service = %v, want %v",
				path.Path, path.Backend.Service.Name, expectedService)
		}

		if path.Backend.Service.Port.Number != 8080 {
			t.Errorf("path %s port = %v, want %v",
				path.Path, path.Backend.Service.Port.Number, 8080)
		}
	}
}

func TestManager_EnsureIngress_OwnerTrackingAnnotations(t *testing.T) {
	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-222",
			Namespace: "previewd-system",
			UID:       "test-uid-5",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   222,
			Repository: "owner/repo",
			Services:   []string{"api"},
		},
	}
	namespace := "preview-pr-222-mno222"

	c := setupTestClient(t, namespace)
	m := NewManager(c, c.Scheme(), "preview.example.com", "letsencrypt-prod")
	err := m.EnsureIngress(context.Background(), preview, namespace)

	if err != nil {
		t.Fatalf("EnsureIngress() error = %v", err)
	}

	ingress := &networkingv1.Ingress{}
	err = c.Get(context.Background(), types.NamespacedName{
		Name:      "preview-ingress",
		Namespace: namespace,
	}, ingress)
	if err != nil {
		t.Fatalf("failed to get ingress: %v", err)
	}

	// Verify owner tracking via annotations (cross-namespace owner references are not allowed)
	if ingress.Annotations["preview.previewd.io/owner-uid"] != string(preview.UID) {
		t.Errorf("owner UID annotation = %v, want %v",
			ingress.Annotations["preview.previewd.io/owner-uid"], preview.UID)
	}
	if ingress.Annotations["preview.previewd.io/owner-name"] != preview.Name {
		t.Errorf("owner name annotation = %v, want %v",
			ingress.Annotations["preview.previewd.io/owner-name"], preview.Name)
	}
	if ingress.Annotations["preview.previewd.io/owner-namespace"] != preview.Namespace {
		t.Errorf("owner namespace annotation = %v, want %v",
			ingress.Annotations["preview.previewd.io/owner-namespace"], preview.Namespace)
	}
}

func TestManager_EnsureIngress_Idempotent(t *testing.T) {
	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-333",
			Namespace: "previewd-system",
			UID:       "test-uid-6",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   333,
			Repository: "owner/repo",
			Services:   []string{"api"},
		},
	}
	namespace := "preview-pr-333-pqr333"

	c := setupTestClient(t, namespace)
	m := NewManager(c, c.Scheme(), "preview.example.com", "letsencrypt-prod")

	// Call twice to test idempotency
	err := m.EnsureIngress(context.Background(), preview, namespace)
	if err != nil {
		t.Fatalf("first EnsureIngress() error = %v", err)
	}

	err = m.EnsureIngress(context.Background(), preview, namespace)
	if err != nil {
		t.Fatalf("second EnsureIngress() error = %v", err)
	}

	ingress := &networkingv1.Ingress{}
	err = c.Get(context.Background(), types.NamespacedName{
		Name:      "preview-ingress",
		Namespace: namespace,
	}, ingress)
	if err != nil {
		t.Fatalf("ingress should exist: %v", err)
	}
}

func TestManager_EnsureIngress_SSLRedirect(t *testing.T) {
	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-444",
			Namespace: "previewd-system",
			UID:       "test-uid-7",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   444,
			Repository: "owner/repo",
			Services:   []string{"api"},
		},
	}
	namespace := "preview-pr-444-stu444"

	scheme := runtime.NewScheme()
	if err := previewv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add preview scheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add core scheme: %v", err)
	}
	if err := networkingv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add networking scheme: %v", err)
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(ns).
		Build()

	m := NewManager(c, scheme, "preview.example.com", "letsencrypt-prod")
	err := m.EnsureIngress(context.Background(), preview, namespace)
	if err != nil {
		t.Fatalf("EnsureIngress() error = %v", err)
	}

	ingress := &networkingv1.Ingress{}
	err = c.Get(context.Background(), types.NamespacedName{
		Name:      "preview-ingress",
		Namespace: namespace,
	}, ingress)
	if err != nil {
		t.Fatalf("failed to get ingress: %v", err)
	}

	// Verify SSL redirect annotation
	sslRedirectAnnotation := "nginx.ingress.kubernetes.io/ssl-redirect"
	if ingress.Annotations[sslRedirectAnnotation] != "true" {
		t.Errorf("ssl-redirect annotation = %v, want %v",
			ingress.Annotations[sslRedirectAnnotation], "true")
	}
}

func TestManager_GetIngressHost(t *testing.T) {
	tests := []struct {
		preview  *previewv1alpha1.PreviewEnvironment
		name     string
		wantHost string
	}{
		{
			name: "generates correct host for PR 123",
			preview: &previewv1alpha1.PreviewEnvironment{
				Spec: previewv1alpha1.PreviewEnvironmentSpec{
					PRNumber: 123,
				},
			},
			wantHost: "pr-123.preview.example.com",
		},
		{
			name: "generates correct host for PR 9999",
			preview: &previewv1alpha1.PreviewEnvironment{
				Spec: previewv1alpha1.PreviewEnvironmentSpec{
					PRNumber: 9999,
				},
			},
			wantHost: "pr-9999.preview.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			c := fake.NewClientBuilder().WithScheme(scheme).Build()
			m := NewManager(c, scheme, "preview.example.com", "letsencrypt-prod")

			got := m.GetIngressHost(tt.preview)
			if got != tt.wantHost {
				t.Errorf("GetIngressHost() = %v, want %v", got, tt.wantHost)
			}
		})
	}
}

func TestManager_EnsureIngress_PathType(t *testing.T) {
	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-555",
			Namespace: "previewd-system",
			UID:       "test-uid-8",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   555,
			Repository: "owner/repo",
			Services:   []string{"auth", "frontend"},
		},
	}
	namespace := "preview-pr-555-vwx555"

	scheme := runtime.NewScheme()
	if err := previewv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add preview scheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add core scheme: %v", err)
	}
	if err := networkingv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add networking scheme: %v", err)
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(ns).
		Build()

	m := NewManager(c, scheme, "preview.example.com", "letsencrypt-prod")
	err := m.EnsureIngress(context.Background(), preview, namespace)
	if err != nil {
		t.Fatalf("EnsureIngress() error = %v", err)
	}

	ingress := &networkingv1.Ingress{}
	err = c.Get(context.Background(), types.NamespacedName{
		Name:      "preview-ingress",
		Namespace: namespace,
	}, ingress)
	if err != nil {
		t.Fatalf("failed to get ingress: %v", err)
	}

	// Verify path type
	for _, path := range ingress.Spec.Rules[0].HTTP.Paths {
		if path.PathType == nil {
			t.Errorf("path %s has no PathType", path.Path)
			continue
		}

		expectedPathType := networkingv1.PathTypePrefix
		if *path.PathType != expectedPathType {
			t.Errorf("path %s PathType = %v, want %v",
				path.Path, *path.PathType, expectedPathType)
		}
	}
}

func TestManager_EnsureIngress_ServicePort(t *testing.T) {
	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-666",
			Namespace: "previewd-system",
			UID:       "test-uid-9",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   666,
			Repository: "owner/repo",
			Services:   []string{"api"},
		},
	}
	namespace := "preview-pr-666-yza666"

	scheme := runtime.NewScheme()
	if err := previewv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add preview scheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add core scheme: %v", err)
	}
	if err := networkingv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add networking scheme: %v", err)
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(ns).
		Build()

	m := NewManager(c, scheme, "preview.example.com", "letsencrypt-prod")
	err := m.EnsureIngress(context.Background(), preview, namespace)
	if err != nil {
		t.Fatalf("EnsureIngress() error = %v", err)
	}

	ingress := &networkingv1.Ingress{}
	err = c.Get(context.Background(), types.NamespacedName{
		Name:      "preview-ingress",
		Namespace: namespace,
	}, ingress)
	if err != nil {
		t.Fatalf("failed to get ingress: %v", err)
	}

	// Verify service port is 8080
	for _, path := range ingress.Spec.Rules[0].HTTP.Paths {
		if path.Backend.Service.Port.Number != 8080 {
			t.Errorf("path %s port = %v, want %v",
				path.Path, path.Backend.Service.Port.Number, 8080)
		}
	}
}

func TestGenerateServiceName(t *testing.T) {
	tests := []struct {
		prNumber int
		service  string
		want     string
	}{
		{
			prNumber: 123,
			service:  "auth",
			want:     "preview-pr-123-auth",
		},
		{
			prNumber: 456,
			service:  "api",
			want:     "preview-pr-456-api",
		},
		{
			prNumber: 789,
			service:  "frontend",
			want:     "preview-pr-789-frontend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := generateServiceName(tt.prNumber, tt.service)
			if got != tt.want {
				t.Errorf("generateServiceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeneratePathForService(t *testing.T) {
	tests := []struct {
		service  string
		name     string
		wantPath string
	}{
		{
			name:     "frontend service gets root path",
			service:  "frontend",
			wantPath: "/",
		},
		{
			name:     "auth service gets /auth path",
			service:  "auth",
			wantPath: "/auth",
		},
		{
			name:     "api service gets /api path",
			service:  "api",
			wantPath: "/api",
		},
		{
			name:     "other service gets /<service> path",
			service:  "users",
			wantPath: "/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generatePathForService(tt.service)
			if got != tt.wantPath {
				t.Errorf("generatePathForService() = %v, want %v", got, tt.wantPath)
			}
		})
	}
}

func TestManager_EnsureIngress_PathOrdering(t *testing.T) {
	// CRITICAL: This test verifies that "/" (frontend) appears LAST in the paths array.
	// With PathTypePrefix, "/" matches ALL requests, so specific paths like "/auth" and "/api"
	// must come BEFORE "/" or they will be unreachable.
	preview := &previewv1alpha1.PreviewEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-999",
			Namespace: "previewd-system",
			UID:       "test-uid-ordering",
		},
		Spec: previewv1alpha1.PreviewEnvironmentSpec{
			PRNumber:   999,
			Repository: "owner/repo",
			// Services intentionally in non-sorted order to verify sorting
			Services: []string{"frontend", "auth", "api"},
		},
	}
	namespace := "preview-pr-999-order999"

	scheme := runtime.NewScheme()
	if err := previewv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add preview scheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add core scheme: %v", err)
	}
	if err := networkingv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add networking scheme: %v", err)
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(ns).
		Build()

	m := NewManager(c, scheme, "preview.example.com", "letsencrypt-prod")
	err := m.EnsureIngress(context.Background(), preview, namespace)
	if err != nil {
		t.Fatalf("EnsureIngress() error = %v", err)
	}

	ingress := &networkingv1.Ingress{}
	err = c.Get(context.Background(), types.NamespacedName{
		Name:      "preview-ingress",
		Namespace: namespace,
	}, ingress)
	if err != nil {
		t.Fatalf("failed to get ingress: %v", err)
	}

	// Verify path count
	paths := ingress.Spec.Rules[0].HTTP.Paths
	if len(paths) != 3 {
		t.Fatalf("expected 3 paths, got %d", len(paths))
	}

	// CRITICAL: Verify "/" (frontend) is LAST
	lastPath := paths[len(paths)-1]
	if lastPath.Path != "/" {
		t.Errorf("last path should be '/', got %q", lastPath.Path)
	}
	if lastPath.Backend.Service.Name != "preview-pr-999-frontend" {
		t.Errorf("last path should route to frontend service, got %q", lastPath.Backend.Service.Name)
	}

	// Verify specific paths come BEFORE "/"
	for i := 0; i < len(paths)-1; i++ {
		if paths[i].Path == "/" {
			t.Errorf("path '/' should be last, but found at index %d", i)
		}
	}

	// Verify all expected paths exist in correct order
	// Order should be: /api, /auth, / (alphabetical except "/" last)
	expectedOrder := []struct {
		path    string
		service string
	}{
		{"/api", "preview-pr-999-api"},
		{"/auth", "preview-pr-999-auth"},
		{"/", "preview-pr-999-frontend"},
	}

	for i, expected := range expectedOrder {
		if paths[i].Path != expected.path {
			t.Errorf("path[%d] = %q, want %q", i, paths[i].Path, expected.path)
		}
		if paths[i].Backend.Service.Name != expected.service {
			t.Errorf("path[%d] service = %q, want %q", i, paths[i].Backend.Service.Name, expected.service)
		}
	}
}
