# PreviewEnvironment API Reference

This document provides detailed information about the PreviewEnvironment Custom Resource Definition (CRD).

## API Version

- **Group:** `preview.previewd.io`
- **Version:** `v1alpha1`
- **Kind:** `PreviewEnvironment`

## Resource Shortnames

- `preview`
- `previews`

## PreviewEnvironment

PreviewEnvironment is the Schema for the previewenvironments API. It represents a preview environment for a pull request, including all necessary configuration and status information.

### Spec Fields

The `spec` field defines the desired state of the PreviewEnvironment.

#### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `repository` | string | GitHub repository in "owner/repo" format | Pattern: `^[a-zA-Z0-9-]+/[a-zA-Z0-9-]+$` |
| `prNumber` | integer | Pull request number | Minimum: 1 |
| `headSHA` | string | Commit SHA of the PR head | Pattern: `^[a-f0-9]{40}$` (40-character lowercase hex) |

#### Optional Fields

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `baseBranch` | string | Base branch name (e.g., "main", "develop") | - |
| `headBranch` | string | Head branch name (e.g., "feature/my-feature") | - |
| `services` | []string | List of service names to deploy | - |
| `ttl` | string | Time-to-live duration for the preview environment | `"4h"` |

### Status Fields

The `status` field defines the observed state of the PreviewEnvironment. This is managed by the operator and should not be set directly by users.

| Field | Type | Description |
|-------|------|-------------|
| `phase` | string | Current phase of the preview environment. Valid values: `Pending`, `Creating`, `Ready`, `Updating`, `Deleting`, `Failed` |
| `url` | string | Public URL to access the preview environment |
| `namespace` | string | Kubernetes namespace created for this preview environment |
| `services` | []ServiceStatus | Status information for deployed services |
| `costEstimate` | CostEstimate | Estimated costs for running this environment |
| `conditions` | []metav1.Condition | Standard Kubernetes conditions |
| `createdAt` | metav1.Time | Timestamp when the environment was created |
| `expiresAt` | metav1.Time | Timestamp when the environment will be automatically deleted |
| `lastSyncedAt` | metav1.Time | Timestamp of the last successful sync |
| `observedGeneration` | int64 | Generation of the most recently observed spec |

### Nested Types

#### ServiceStatus

Represents the status of a deployed service.

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `name` | string | Service name | Yes |
| `ready` | boolean | Indicates if the service is ready | Yes |
| `url` | string | Service URL (if exposed) | No |

#### CostEstimate

Provides cost estimation for the preview environment.

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `currency` | string | Cost currency (e.g., "USD") | Yes |
| `hourlyCost` | string | Estimated hourly cost | Yes |
| `totalCost` | string | Total estimated cost based on TTL | No |

## Examples

### Minimal Example

```yaml
apiVersion: preview.previewd.io/v1alpha1
kind: PreviewEnvironment
metadata:
  name: pr-123
spec:
  repository: myorg/myrepo
  prNumber: 123
  headSHA: 1234567890abcdef1234567890abcdef12345678
```

This will create a preview environment with default TTL of 4 hours.

### Complete Example

```yaml
apiVersion: preview.previewd.io/v1alpha1
kind: PreviewEnvironment
metadata:
  name: pr-456-feature-auth
  namespace: preview-system
spec:
  repository: myorg/myrepo
  prNumber: 456
  headSHA: abcdef1234567890abcdef1234567890abcdef12
  baseBranch: main
  headBranch: feature/auth-improvements
  ttl: "8h"
  services:
    - api
    - web
    - worker
    - database
```

### With Status (Managed by Operator)

```yaml
apiVersion: preview.previewd.io/v1alpha1
kind: PreviewEnvironment
metadata:
  name: pr-789
spec:
  repository: myorg/myrepo
  prNumber: 789
  headSHA: fedcba0987654321fedcba0987654321fedcba09
status:
  phase: Ready
  url: https://pr-789.preview.example.com
  namespace: preview-pr-789
  services:
    - name: api
      ready: true
      url: https://api-pr-789.preview.example.com
    - name: web
      ready: true
      url: https://pr-789.preview.example.com
  costEstimate:
    currency: USD
    hourlyCost: "0.15"
    totalCost: "0.60"
  conditions:
    - type: Ready
      status: "True"
      reason: ServicesReady
      message: All services are healthy
      lastTransitionTime: "2025-11-09T12:00:00Z"
  createdAt: "2025-11-09T08:00:00Z"
  expiresAt: "2025-11-09T16:00:00Z"
  lastSyncedAt: "2025-11-09T12:00:00Z"
  observedGeneration: 1
```

## kubectl Commands

### Create a PreviewEnvironment

```bash
kubectl apply -f preview.yaml
```

### List PreviewEnvironments

```bash
kubectl get previews
# or
kubectl get previewenvironments
```

Output includes custom columns:
- **PR**: Pull request number
- **Phase**: Current phase (Pending, Creating, Ready, etc.)
- **URL**: Preview environment URL
- **Age**: Time since creation

Example output:
```
NAME      PR    PHASE    URL                                  AGE
pr-123    123   Ready    https://pr-123.preview.example.com   2h
pr-456    456   Creating                                      5m
```

### Get PreviewEnvironment Details

```bash
kubectl get preview pr-123 -o yaml
```

### Update Status (Operator Only)

The operator updates the status subresource:

```bash
kubectl patch preview pr-123 --subresource=status --type=merge -p '{"status":{"phase":"Ready"}}'
```

### Delete a PreviewEnvironment

```bash
kubectl delete preview pr-123
```

## Validation Rules

### Repository

- **Pattern**: `^[a-zA-Z0-9-]+/[a-zA-Z0-9-]+$`
- **Valid**: `myorg/myrepo`, `My-Org-123/my-repo-456`
- **Invalid**: `myorg`, `myorg/`, `/myrepo`, `myorg/my/repo`

### PRNumber

- **Minimum**: 1
- **Valid**: 1, 123, 999999
- **Invalid**: 0, -1

### HeadSHA

- **Pattern**: `^[a-f0-9]{40}$`
- **Length**: Exactly 40 characters
- **Characters**: Lowercase hexadecimal only (0-9, a-f)
- **Valid**: `1234567890abcdef1234567890abcdef12345678`
- **Invalid**:
  - `short` (too short)
  - `1234567890ABCDEF...` (uppercase not allowed)
  - `1234567890ghijkl...` (invalid hex characters)

### TTL

- **Format**: Duration string (e.g., "4h", "30m", "2h30m")
- **Default**: `"4h"`
- **Examples**: "1h", "2h30m", "8h", "24h"

### Phase (Status)

Valid values:
- `Pending`: Environment is queued for creation
- `Creating`: Resources are being provisioned
- `Ready`: Environment is fully operational
- `Updating`: Environment is being updated
- `Deleting`: Environment is being torn down
- `Failed`: Environment creation or operation failed

## Notes

- The `status` subresource is enabled, allowing separate RBAC permissions for status updates
- All timestamps use RFC3339 format
- The operator automatically sets owner references for garbage collection
- Finalizers ensure proper cleanup when a PreviewEnvironment is deleted
