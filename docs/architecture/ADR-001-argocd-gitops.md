# ADR-001: Use ArgoCD for GitOps-Based Deployments

## Status

**Accepted** - 2025-11-09

## Context

Previewd needs a deployment mechanism to create preview environments with multiple services. The operator must deploy applications reliably, support rollbacks, and provide observability into deployment status.

### Requirements

1. **Declarative deployments**: Define applications in Git, not imperative scripts
2. **Multi-service support**: Deploy 5-10 services per preview environment
3. **Namespace isolation**: Each preview in separate namespace
4. **Status visibility**: Know when applications are ready/healthy
5. **Automatic sync**: React to Git changes automatically
6. **Rollback capability**: Easy rollback on failure
7. **Multi-tenancy**: Support 100s of concurrent previews
8. **GitOps principles**: Git as source of truth

### Options Considered

#### Option 1: Native Kubernetes Deployments (kubectl apply)

**Approach**: Operator creates Deployment/Service/Ingress resources directly

**Pros**:
- Simple, no external dependencies
- Full control over resources
- Fast deployment (no intermediate layer)

**Cons**:
- ❌ No Git-based audit trail
- ❌ Operator manages all deployment logic (complex)
- ❌ No automatic sync on Git changes
- ❌ Difficult rollback (manual tracking of versions)
- ❌ No health checks built-in
- ❌ Requires implementing deployment strategies (rolling, canary)

**Example**:
```go
deployment := &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{Name: "auth", Namespace: "pr-123"},
    Spec: appsv1.DeploymentSpec{...},
}
k8sClient.Create(ctx, deployment)
```

---

#### Option 2: Helm (via Helm SDK)

**Approach**: Operator uses Helm SDK to install charts

**Pros**:
- Templating for configuration
- Mature ecosystem (Helm charts widely available)
- Versioned releases

**Cons**:
- ❌ Not true GitOps (state in Kubernetes, not Git)
- ❌ Helm SDK dependency in operator (complexity)
- ❌ No automatic sync (operator must poll Git)
- ❌ Release management complexity
- ❌ Drift detection not built-in

**Example**:
```go
actionConfig := &action.Configuration{}
install := action.NewInstall(actionConfig)
install.Namespace = "pr-123"
install.ReleaseName = "preview-123"
chart, _ := loader.Load("./charts/myapp")
install.Run(chart, values)
```

---

#### Option 3: Flux CD

**Approach**: Operator creates Flux Kustomization/HelmRelease resources

**Pros**:
- ✅ True GitOps (Git as source of truth)
- ✅ Automatic sync on Git changes
- ✅ Drift detection and reconciliation
- ✅ Simple CRD-based API
- ✅ Lightweight (fewer dependencies)

**Cons**:
- ⚠️ Less mature than ArgoCD
- ⚠️ Simpler UI (less observability)
- ⚠️ No ApplicationSet equivalent (multi-tenant complexity)
- ❌ Weaker multi-tenancy model

**Example**:
```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: preview-123
  namespace: flux-system
spec:
  sourceRef:
    kind: GitRepository
    name: myapp
  path: ./k8s/overlays/preview
  prune: true
  targetNamespace: pr-123
```

---

#### Option 4: ArgoCD (Chosen)

**Approach**: Operator creates ArgoCD ApplicationSet resources

**Pros**:
- ✅ True GitOps (Git as source of truth)
- ✅ Automatic sync on Git changes
- ✅ Drift detection and reconciliation
- ✅ **ApplicationSet**: Generate multiple Applications from template (perfect for multi-service previews)
- ✅ Multi-tenancy built-in (AppProjects)
- ✅ Rich UI for observability
- ✅ Health checks for all resource types
- ✅ Mature ecosystem (widely adopted)
- ✅ RBAC and security features
- ✅ Sync waves for dependency ordering

**Cons**:
- ⚠️ Heavier weight (more components: server, repo-server, controller)
- ⚠️ External dependency (cluster must have ArgoCD installed)
- ⚠️ Slight learning curve for users

**Example**:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: preview-123
  namespace: argocd
spec:
  generators:
    - list:
        elements:
          - service: auth
          - service: api
  template:
    metadata:
      name: 'preview-123-{{service}}'
    spec:
      source:
        repoURL: https://github.com/company/services
        path: '{{service}}/k8s'
      destination:
        namespace: pr-123
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
```

---

## Decision

**We will use ArgoCD with ApplicationSet for deployments.**

### Rationale

1. **ApplicationSet is perfect for preview environments**: One ApplicationSet generates multiple Application CRs (one per service), all sharing the same preview configuration (PR number, namespace, etc.)

2. **GitOps principles**: Git is the source of truth for application manifests. Operator only creates ApplicationSet, ArgoCD handles the rest.

3. **Observability**: ArgoCD provides rich UI and API to query application health, sync status, and events. Operator can watch Application CRs for readiness.

4. **Automatic sync**: ArgoCD automatically syncs changes from Git. When developers push new commits, ArgoCD updates preview environments without operator intervention.

5. **Multi-tenancy**: ArgoCD's AppProject feature supports RBAC and resource quotas per project. Preview environments can be isolated.

6. **Mature ecosystem**: ArgoCD is the most widely adopted GitOps tool for Kubernetes (CNCF graduated project).

7. **Separation of concerns**: Operator focuses on lifecycle (create/delete previews), ArgoCD focuses on deployment (sync applications).

### Trade-offs Accepted

- **External dependency**: ArgoCD must be installed in the cluster before Previewd. This is acceptable because ArgoCD is a standard component in modern Kubernetes environments.

- **Resource overhead**: ArgoCD adds ~500Mi memory, 0.5 CPU per cluster. This is negligible compared to preview environment costs.

- **Complexity**: ArgoCD has more components (server, controller, repo-server) than Flux. However, ArgoCD is well-documented and widely supported.

---

## Implementation

### Operator Responsibilities

1. **Create ApplicationSet** when PreviewEnvironment CR is created
2. **Watch Application CRs** to determine readiness
3. **Delete ApplicationSet** when PreviewEnvironment is deleted (cascade deletes Applications)
4. **Update ApplicationSet** when PreviewEnvironment spec changes (e.g., new commit)

### ApplicationSet Pattern

**List Generator**: Hardcoded list of services per preview

```yaml
generators:
  - list:
      elements:
        - service: auth
          repo: https://github.com/company/auth-service
        - service: api
          repo: https://github.com/company/api-service
```

**Template**: Shared configuration for all services

```yaml
template:
  metadata:
    name: 'preview-{{prNumber}}-{{service}}'
  spec:
    project: default
    source:
      repoURL: '{{repo}}'
      targetRevision: main
      path: k8s/overlays/preview
    destination:
      namespace: 'pr-{{prNumber}}'
    syncPolicy:
      automated:
        prune: true
        selfHeal: true
```

### Health Checks

Operator watches Application resources:

```go
app := &argov1alpha1.Application{}
err := k8sClient.Get(ctx, types.NamespacedName{Name: "preview-123-auth", Namespace: "argocd"}, app)

if app.Status.Health.Status == health.HealthStatusHealthy &&
   app.Status.Sync.Status == synccommon.SyncStatusCodeSynced {
    // Service is ready
}
```

### Owner References

ApplicationSet has owner reference to PreviewEnvironment:

```yaml
metadata:
  ownerReferences:
    - apiVersion: preview.previewd.io/v1alpha1
      kind: PreviewEnvironment
      name: pr-123
      uid: abc-123-def
      controller: true
      blockOwnerDeletion: true
```

When PreviewEnvironment is deleted → ApplicationSet is deleted → Applications are deleted → Resources are pruned.

---

## Alternatives for Specific Use Cases

### Future: Multi-Cluster Support (v0.3.0)

**Problem**: ArgoCD can deploy to remote clusters, but requires cluster credentials in ArgoCD namespace.

**Solution**: Use ArgoCD's [Cluster Secret](https://argo-cd.readthedocs.io/en/stable/operator-manual/declarative-setup/#clusters) to register remote clusters.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: preview-cluster-1
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: cluster
data:
  name: preview-cluster-1
  server: https://preview-cluster-1.example.com
  config: |
    {
      "bearerToken": "<token>",
      "tlsClientConfig": {...}
    }
```

### Future: On-Prem Air-Gapped Deployments (v0.3.0)

**Problem**: ArgoCD requires internet access to pull from Git.

**Solution**: Use private Git repositories or Git mirrors within air-gapped network.

---

## Consequences

### Positive

- ✅ **Declarative deployments**: All application state in Git
- ✅ **Automatic sync**: No polling required by operator
- ✅ **Observability**: Rich UI and API for deployment status
- ✅ **Rollback**: Built-in via ArgoCD rollback feature
- ✅ **Multi-tenancy**: ApplicationSet scales to 1000s of previews
- ✅ **Separation of concerns**: Operator doesn't implement deployment logic

### Negative

- ❌ **External dependency**: Requires ArgoCD installation
- ❌ **Resource overhead**: ArgoCD adds ~0.5 CPU, 500Mi memory
- ❌ **Learning curve**: Users must understand ArgoCD concepts

### Neutral

- ⚠️ **Flexibility**: Locked into ArgoCD patterns (acceptable for v0.1.0, can revisit in v1.0.0)
- ⚠️ **Complexity**: More moving parts (ApplicationSet → Application → Deployment)

---

## Validation

### Success Criteria

1. **Performance**: <2 minutes from ApplicationSet creation → all Applications synced
2. **Reliability**: 99.9% sync success rate
3. **Scalability**: Support 100+ ApplicationSets (1000+ Applications)
4. **Observability**: Operator can query Application status in <100ms

### Monitoring

- **Metric**: `previewd_argocd_applicationset_creation_duration_seconds`
- **Metric**: `previewd_argocd_application_sync_errors_total`
- **Alert**: Fire if >10% of Applications fail to sync

---

## References

- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [ApplicationSet Documentation](https://argo-cd.readthedocs.io/en/stable/user-guide/application-set/)
- [GitOps Principles](https://opengitops.dev/)
- [CNCF ArgoCD Project](https://www.cncf.io/projects/argo/)

---

**Author**: Mike Lane (@mikelane)
**Reviewers**: TBD
**Implementation**: Tracked in #TBD
