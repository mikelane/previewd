# ADR-002: Namespace-per-PR Isolation Strategy

## Status

**Accepted** - 2025-11-09

## Context

Preview environments must be isolated to prevent interference between different pull requests. Multiple isolation strategies exist in Kubernetes, each with trade-offs in security, resource usage, and operational complexity.

### Requirements

1. **Security isolation**: Prevent PRs from accessing each other's resources
2. **Resource isolation**: Limit resource consumption per PR
3. **Network isolation**: Control ingress/egress traffic per PR
4. **Simple cleanup**: Easy deletion when PR is closed
5. **Cost efficiency**: Minimal overhead per preview
6. **Scalability**: Support 100-1000+ concurrent previews
7. **Developer experience**: Simple, predictable URLs (pr-123.preview.company.com)

### Options Considered

#### Option 1: Shared Namespace with Labels

**Approach**: All previews in single namespace, resources labeled with PR number

**Pros**:
- Lowest resource overhead (one namespace)
- Simple RBAC (single namespace policy)
- Fast creation (no namespace overhead)

**Cons**:
- ❌ **Weak isolation**: Services can access each other (same namespace)
- ❌ **No resource quotas**: Can't limit CPU/memory per PR
- ❌ **Complex cleanup**: Must delete all resources with label selector
- ❌ **Network policy limitations**: Can't deny all by default (affects all PRs)
- ❌ **Name collisions**: Must ensure unique names (pr-123-auth vs pr-456-auth)

**Example**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pr-123-auth
  namespace: previews  # Shared namespace
  labels:
    preview.previewd.io/pr: "123"
```

---

#### Option 2: Namespace-per-PR (Chosen)

**Approach**: Each PR gets dedicated namespace (pr-123, pr-456, etc.)

**Pros**:
- ✅ **Strong isolation**: Complete separation (network, resources, RBAC)
- ✅ **Resource quotas**: Enforce CPU/memory limits per PR
- ✅ **Network policies**: Deny-all by default, allow specific ingress
- ✅ **Simple cleanup**: Delete namespace → all resources deleted (cascade)
- ✅ **No name collisions**: Services can have same names across previews
- ✅ **Clear ownership**: Namespace has owner reference to PreviewEnvironment CR
- ✅ **Observability**: Easy to query resources per PR (`kubectl get all -n pr-123`)

**Cons**:
- ⚠️ **Resource overhead**: Each namespace consumes ~1Mi memory, negligible CPU
- ⚠️ **Etcd storage**: More objects in etcd (100 namespaces = ~100KB)
- ⚠️ **Namespace creation time**: ~500ms per namespace (acceptable)

**Example**:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: pr-123
  labels:
    preview.previewd.io/pr: "123"
    preview.previewd.io/repository: company-monorepo
  ownerReferences:
    - apiVersion: preview.previewd.io/v1alpha1
      kind: PreviewEnvironment
      name: pr-123
      controller: true
```

---

#### Option 3: Virtual Clusters (vcluster)

**Approach**: Each PR gets virtual Kubernetes cluster (nested control plane)

**Pros**:
- ✅ **Maximum isolation**: Complete API server per PR
- ✅ **Full Kubernetes API**: Can use CRDs, RBAC, etc. inside vcluster
- ✅ **Multi-tenancy**: Strong security guarantees

**Cons**:
- ❌ **High overhead**: ~200Mi memory, 0.2 CPU per vcluster
- ❌ **Complexity**: Additional control plane components
- ❌ **Slow creation**: ~30 seconds per vcluster
- ❌ **Cost**: 100 vclusters = 20 CPU, 20Gi memory (vs namespaces: ~0.1 CPU, 100Mi)
- ❌ **Overkill**: Preview environments don't need full cluster isolation

---

#### Option 4: Cluster-per-PR

**Approach**: Each PR gets dedicated Kubernetes cluster

**Pros**:
- ✅ **Perfect isolation**: Complete cluster separation
- ✅ **No shared control plane**: Blast radius limited to single PR

**Cons**:
- ❌ **Prohibitively expensive**: $50-100/month per cluster
- ❌ **Slow creation**: 5-10 minutes per cluster
- ❌ **Operational complexity**: Managing 100s of clusters
- ❌ **Not practical**: Designed for long-lived environments, not ephemeral previews

---

## Decision

**We will use namespace-per-PR isolation.**

### Rationale

1. **Sufficient isolation**: Namespaces provide strong security boundaries via NetworkPolicy and RBAC. Preview environments don't require cluster-level isolation.

2. **Resource efficiency**: Namespace overhead is negligible (~1Mi memory per namespace). 1000 namespaces = ~1Gi memory, vs vcluster (200Gi) or clusters ($50k/month).

3. **Simple lifecycle**: Delete namespace → cascade delete all resources. No need to track individual resources or use complex label selectors.

4. **Native Kubernetes**: Namespaces are first-class Kubernetes primitives with excellent tooling support.

5. **Resource quotas**: ResourceQuota objects enforce limits per namespace, preventing noisy neighbor issues.

6. **Network isolation**: NetworkPolicy applies per namespace, enabling deny-all-by-default with specific allow rules.

7. **Observability**: Easy to query resources (`kubectl get all -n pr-123`) and set up per-namespace monitoring.

### Trade-offs Accepted

- **Namespace overhead**: ~1Mi memory per namespace. For 1000 concurrent previews, this is ~1Gi memory total (negligible).

- **Namespace creation time**: ~500ms per namespace. This adds <1% to total preview creation time (<2 min target).

- **Etcd storage**: Each namespace adds ~1KB to etcd. 1000 namespaces = ~1MB (negligible).

---

## Implementation

### Namespace Naming Convention

**Pattern**: `pr-{number}` (e.g., `pr-123`, `pr-456`)

**Rationale**:
- Short and predictable
- Matches PR number (easy to correlate)
- No special characters (DNS-compatible)
- Unique per repository (PR numbers are unique)

**Collision handling**: If namespace already exists, operator reconciles existing namespace (idempotent).

### Namespace Metadata

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: pr-123
  labels:
    preview.previewd.io/pr: "123"
    preview.previewd.io/repository: company/monorepo
    preview.previewd.io/managed-by: previewd
  annotations:
    preview.previewd.io/created-at: "2025-11-09T10:30:00Z"
    preview.previewd.io/expires-at: "2025-11-09T14:30:00Z"
  ownerReferences:
    - apiVersion: preview.previewd.io/v1alpha1
      kind: PreviewEnvironment
      name: pr-123
      uid: abc-123-def
      controller: true
      blockOwnerDeletion: true
```

**Owner reference**: Enables cascade deletion (delete PreviewEnvironment → delete Namespace → delete all resources).

### Resource Quota

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: preview-quota
  namespace: pr-123
spec:
  hard:
    # Compute
    requests.cpu: "2"
    requests.memory: 4Gi
    limits.cpu: "4"
    limits.memory: 8Gi

    # Storage (prevent accidental PVC creation)
    persistentvolumeclaims: "0"

    # Networking
    services.loadbalancers: "0"  # Prevent external LBs
    services.nodeports: "0"      # Prevent NodePort services

    # Objects (prevent resource exhaustion)
    count/pods: "20"
    count/services: "10"
    count/configmaps: "10"
    count/secrets: "10"
```

**Rationale**:
- Prevents single PR from consuming excessive resources
- Blocks expensive resources (LoadBalancers, PVCs)
- Limits object count to prevent etcd exhaustion

### Network Policies

**Default deny all**:
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: pr-123
spec:
  podSelector: {}
  policyTypes:
    - Ingress
    - Egress
```

**Allow ingress from Ingress Controller**:
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-ingress
  namespace: pr-123
spec:
  podSelector: {}
  policyTypes:
    - Ingress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx
      ports:
        - protocol: TCP
          port: 8080
```

**Allow egress to DNS and internet**:
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-egress
  namespace: pr-123
spec:
  podSelector: {}
  policyTypes:
    - Egress
  egress:
    # DNS (kube-system)
    - to:
        - namespaceSelector:
            matchLabels:
              name: kube-system
      ports:
        - protocol: UDP
          port: 53

    # HTTPS to external APIs
    - to:
        - podSelector: {}
      ports:
        - protocol: TCP
          port: 443

    # HTTP within namespace (service-to-service)
    - to:
        - podSelector: {}
      ports:
        - protocol: TCP
          port: 8080
```

### Cleanup Strategy

**Cascade deletion**:
1. Delete PreviewEnvironment CR
2. Finalizer runs (delete ApplicationSet)
3. Owner reference cascade deletes Namespace
4. Kubernetes deletes all resources in Namespace
5. external-dns removes DNS records
6. cert-manager deletes TLS certificates

**Grace period**: 30 seconds (terminationGracePeriodSeconds)

**Forced deletion**: If namespace stuck in Terminating state, operator can patch finalizers to force delete.

---

## Alternatives for Specific Use Cases

### Future: Shared Services (v0.3.0)

**Problem**: Some services (databases, caches) should be shared across previews to reduce cost.

**Solution**: Deploy shared services in separate namespace (`shared-services`), allow egress to specific services.

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-shared-postgres
  namespace: pr-123
spec:
  podSelector: {}
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              name: shared-services
        - podSelector:
            matchLabels:
              app: postgres
      ports:
        - protocol: TCP
          port: 5432
```

### Future: Multi-Tenancy (v0.3.0)

**Problem**: Multiple teams using Previewd, need isolation between teams.

**Solution**: Use namespace prefix (`team-a-pr-123`, `team-b-pr-123`) and RBAC to restrict access.

---

## Consequences

### Positive

- ✅ **Strong isolation**: Network, resource, and RBAC boundaries per PR
- ✅ **Simple cleanup**: Delete namespace → cascade delete all resources
- ✅ **Resource control**: ResourceQuota prevents noisy neighbors
- ✅ **Predictable**: Standard Kubernetes patterns, excellent tooling support
- ✅ **Cost-efficient**: Negligible overhead per namespace

### Negative

- ⚠️ **Namespace limit**: Kubernetes supports ~10,000 namespaces per cluster (sufficient for Previewd use case)
- ⚠️ **Creation time**: ~500ms per namespace (acceptable)

### Neutral

- ⚠️ **Not cluster-level isolation**: Namespaces share control plane. If absolute isolation required, use separate clusters (not needed for preview environments).

---

## Validation

### Success Criteria

1. **Isolation**: Services in pr-123 cannot access services in pr-456
2. **Resource limits**: PR cannot exceed ResourceQuota limits
3. **Cleanup**: Delete PreviewEnvironment → namespace deleted in <30 seconds
4. **Performance**: Namespace creation <500ms
5. **Scalability**: Support 1000+ namespaces per cluster

### Testing

**Isolation test**:
```bash
# Deploy services in pr-123
kubectl run curl --image=curlimages/curl -n pr-123 -- sleep 3600

# Attempt to access service in pr-456
kubectl exec -n pr-123 curl -- curl http://pr-456-auth.pr-456.svc.cluster.local

# Expected: Connection refused (NetworkPolicy blocks)
```

**Cleanup test**:
```bash
# Create PreviewEnvironment
kubectl apply -f preview-123.yaml

# Wait for namespace creation
kubectl wait --for=condition=Ready previewenvironment/pr-123 --timeout=2m

# Delete PreviewEnvironment
kubectl delete previewenvironment/pr-123

# Verify namespace deleted
kubectl get namespace pr-123
# Expected: NotFound (within 30s)
```

### Monitoring

- **Metric**: `kube_namespace_created` (Prometheus kube-state-metrics)
- **Metric**: `kube_namespace_status_phase{phase="Terminating"}` (alert if >5min)
- **Alert**: Fire if namespace stuck in Terminating state for >10 minutes

---

## References

- [Kubernetes Namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)
- [ResourceQuota](https://kubernetes.io/docs/concepts/policy/resource-quotas/)
- [NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [Owner References](https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/)
- [vcluster](https://www.vcluster.com/) (alternative considered)

---

**Author**: Mike Lane (@mikelane)
**Reviewers**: TBD
**Implementation**: Tracked in #TBD
