# Namespace Manager

The namespace manager handles the creation and lifecycle of Kubernetes namespaces for preview environments in the Previewd operator.

## Overview

Each preview environment runs in its own isolated Kubernetes namespace with:
- Resource quotas to prevent resource exhaustion
- Network policies for security isolation
- Deterministic naming for easy identification
- Automatic cleanup when preview environments are deleted

## Features

### Namespace Creation
- **Naming Pattern**: `preview-pr-{PR-NUMBER}-{REPO-HASH}`
- **Labels**: Automatic labeling for identification and filtering
- **Annotations**: Owner tracking for audit and debugging

### Resource Quotas
Default limits per namespace:
- CPU Requests: 2 cores
- Memory Requests: 4Gi
- CPU Limits: 4 cores
- Memory Limits: 8Gi
- PVCs: 0 (no persistent storage)
- LoadBalancers: 0 (use Ingress)

### Network Policies
Three-layer security model:
1. **Default Deny**: Blocks all traffic by default
2. **Selective Ingress**: Only from ingress controller
3. **Controlled Egress**: DNS, HTTPS, and intra-namespace only

## Usage

```go
import "github.com/mikelane/previewd/internal/namespace"

// Create manager
mgr := namespace.NewManager(k8sClient, scheme)

// Create namespace for preview environment
err := mgr.EnsureNamespace(ctx, preview)

// Apply resource quotas
nsName := mgr.GetNamespaceName(preview)
err = mgr.EnsureResourceQuota(ctx, preview, nsName)

// Apply network policies
err = mgr.EnsureNetworkPolicies(ctx, preview, nsName)

// Cleanup when preview is deleted
err = mgr.Cleanup(ctx, preview)
```

## Testing

Run unit tests:
```bash
go test ./internal/namespace/...
```

Run with coverage:
```bash
go test -cover ./internal/namespace/...
```

Run integration tests:
```bash
go test -tags=integration ./internal/namespace/...
```

## Design Decisions

### No Owner References
Kubernetes doesn't allow cross-namespace owner references. Instead of using owner references for garbage collection, we:
- Use labels to associate resources with preview environments
- Store owner information in annotations for debugging
- Implement explicit cleanup in the `Cleanup()` method

### Deterministic Naming
Namespace names include a hash of the repository name to ensure uniqueness when multiple repositories use the same PR numbers. This prevents namespace collisions in multi-tenant environments.

### Security First
The default-deny network policy ensures no accidental exposure. All communication must be explicitly allowed through specific policies.

## Future Enhancements

- [ ] Configurable resource quotas via CRD
- [ ] Custom network policy templates
- [ ] Namespace cost tracking and reporting
- [ ] Multi-tenancy support with namespace prefixes
- [ ] Automatic cleanup of orphaned namespaces