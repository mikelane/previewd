# Previewd Architecture

This document describes the technical architecture of Previewd, an AI-powered Kubernetes operator for preview environments.

## Table of Contents

- [System Overview](#system-overview)
- [Component Architecture](#component-architecture)
- [Custom Resource Definitions](#custom-resource-definitions)
- [Reconciliation Loop](#reconciliation-loop)
- [AI Integration](#ai-integration)
- [Cost Optimization](#cost-optimization)
- [Security Model](#security-model)
- [Deployment Architecture](#deployment-architecture)

---

## System Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          GitHub                                  │
│                                                                   │
│  PR Opened ────────┐                                            │
│  PR Updated ───────┼───► Webhook ────► PreviewdAPI              │
│  PR Closed ────────┘                                            │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Previewd Operator                          │
│                                                                   │
│  ┌────────────────┐  ┌────────────────┐  ┌──────────────────┐ │
│  │   Webhook      │  │   AI Engine    │  │  Cost            │ │
│  │   Server       │  │                │  │  Optimizer       │ │
│  │                │  │  - Code Analysis │  │                │ │
│  │  Validates     │  │  - Data Gen    │  │  - Predict TTL   │ │
│  │  Creates CRs   │  │  - Test Select │  │  - Size Resources │ │
│  └────────┬───────┘  └────────┬───────┘  └────────┬─────────┘ │
│           │                   │                     │           │
│           └───────────────────┴─────────────────────┘           │
│                               │                                 │
│                    ┌──────────▼──────────┐                      │
│                    │   Reconciliation     │                      │
│                    │   Controller         │                      │
│                    │                      │                      │
│                    │  Watches:            │                      │
│                    │  - PreviewEnv CRs    │                      │
│                    │  - Namespaces        │                      │
│                    │  - ArgoCD Apps       │                      │
│                    └──────────┬───────────┘                      │
│                               │                                 │
│           ┌───────────────────┼───────────────────┐             │
│           │                   │                   │             │
│    ┌──────▼──────┐    ┌──────▼──────┐    ┌──────▼──────┐      │
│    │  Namespace  │    │   ArgoCD    │    │   Ingress   │      │
│    │  Manager    │    │   Manager   │    │   Manager   │      │
│    └─────────────┘    └─────────────┘    └─────────────┘      │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                          │
│                                                                   │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │  Namespace: pr-1234                                        │ │
│  │                                                             │ │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────────┐  │ │
│  │  │ Service │  │ Service │  │  Test   │  │   Ingress   │  │ │
│  │  │   A     │  │   B     │  │  Runner │  │             │  │ │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────────┘  │ │
│  │                                                             │ │
│  │  URL: https://pr-1234.preview.example.com                 │ │
│  └───────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### Component Diagram

```
┌──────────────┐       ┌──────────────┐       ┌──────────────┐
│   GitHub     │──────▶│  Previewd    │──────▶│  Kubernetes  │
│   Webhooks   │       │  Operator    │       │   API        │
└──────────────┘       └──────────────┘       └──────────────┘
                              │
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌──────────────┐       ┌──────────────┐     ┌──────────────┐
│   OpenAI     │       │   ArgoCD     │     │  PostgreSQL  │
│   API        │       │   API        │     │  (Metadata)  │
└──────────────┘       └──────────────┘     └──────────────┘
```

---

## Component Architecture

### 1. Webhook Server

**Responsibility:** Receive GitHub webhook events and create/update PreviewEnvironment CRs

**Implementation Status:** ✅ IMPLEMENTED (Issue #4)

**Implementation:**
```go
// internal/webhook/server.go
type Server struct {
    client          client.Client
    webhookSecret   string
    port            int
    logger          logr.Logger
}

func (s *Server) HandlePullRequest(w http.ResponseWriter, r *http.Request) {
    // 1. Validate GitHub signature (HMAC-SHA256)
    // 2. Parse webhook payload
    // 3. Determine action (opened, synchronize, closed, reopened)
    // 4. Create/Update/Delete PreviewEnvironment CR
    // 5. Return appropriate HTTP status
}
```

**Endpoints:**
- `POST /webhook` - GitHub webhook receiver
- `GET /healthz` - Health check
- `GET /readyz` - Readiness check

**Security:**
- Validates GitHub webhook signatures (HMAC-SHA256)
- Signature validation in `internal/webhook/signature.go`
- Rejects requests with invalid or missing signatures
- Logs all webhook events for audit trail

### 2. GitHub Client

**Responsibility:** Interact with GitHub API for PR metadata and commit status updates

**Implementation Status:** ✅ IMPLEMENTED (Issue #5)

**Implementation:**
```go
// internal/github/client.go
type Client struct {
    token      string
    httpClient *http.Client
    logger     logr.Logger
}

func (c *Client) GetPullRequest(ctx context.Context, owner, repo string, number int) (*PullRequest, error) {
    // Fetch PR metadata from GitHub API
}

func (c *Client) UpdateCommitStatus(ctx context.Context, owner, repo, sha string, status *CommitStatus) error {
    // Update commit status with preview URL and deployment state
}
```

**Features:**
- Fetches PR metadata (title, author, SHA, branches)
- Updates commit status with preview environment URL
- Implements retry logic with exponential backoff
- Respects GitHub API rate limits
- Uses GitHub API v3 (REST)

**Commit Status Updates:**
- Context: `previewd`
- States: `pending`, `success`, `failure`
- Includes preview URL as target URL
- Updates shown in PR checks UI

### 3. AI Engine (v0.2.0)

**Responsibility:** Analyze code changes, generate test data, predict costs

**Implementation Status:** 🔮 PLANNED FOR v0.2.0

**Modules:**

#### 3.1 Code Analyzer
```go
// internal/ai/code_analyzer.go
type CodeAnalyzer struct {
    llmClient *openai.Client
    cache     *cache.Cache
}

func (a *CodeAnalyzer) DetectServices(diff string) ([]string, error) {
    // 1. Check cache (identical diffs → same result)
    // 2. Construct prompt for LLM
    // 3. Call OpenAI API
    // 4. Parse and validate response
    // 5. Cache result
    // 6. Return services list
}
```

**Caching strategy:**
- Key: SHA256(diff)
- TTL: 7 days
- Invalidation: On CRD deletion

#### 3.2 Data Generator
```go
// internal/ai/data_generator.go
type DataGenerator struct {
    llmClient *openai.Client
    schemaDB  *sql.DB
}

func (g *DataGenerator) GenerateSyntheticData(schema *Schema, count int) (*Dataset, error) {
    // 1. Fetch production schema (anonymized)
    // 2. Construct prompt with schema + constraints
    // 3. Call OpenAI API
    // 4. Parse SQL INSERT statements
    // 5. Validate referential integrity
    // 6. Return dataset
}
```

**Data generation constraints:**
- Respect foreign keys
- Realistic distributions (names, emails, dates)
- Privacy: never use real production data
- Performance: generate in parallel for large datasets

#### 3.3 Cost Predictor
```go
// internal/ai/cost_predictor.go
type CostPredictor struct {
    llmClient *openai.Client
    history   *HistoryDB
}

func (p *CostPredictor) PredictLifespan(pr *PullRequest) (time.Duration, error) {
    // 1. Fetch user's historical PR patterns
    // 2. Analyze PR metadata (size, complexity)
    // 3. Call LLM for prediction
    // 4. Apply constraints (min 1h, max 7 days)
    // 5. Return duration
}
```

### 3. Reconciliation Controller

**Responsibility:** Manage lifecycle of PreviewEnvironment resources

**Implementation Status:** ✅ IMPLEMENTED (Issue #2)

**Core reconciliation logic:**
```go
// controllers/previewenvironment_controller.go
func (r *PreviewEnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Fetch PreviewEnvironment CR
    // 2. Check if being deleted (finalizers)
    // 3. Determine desired state
    // 4. Compare with actual state
    // 5. Reconcile differences:
    //    - Create namespace if missing
    //    - Deploy services via ArgoCD
    //    - Create ingress
    //    - Run tests
    //    - Update status
    // 6. Requeue if needed
}
```

**State transitions:**
```
Pending → Provisioning → Ready → Testing → Complete
                         ↓
                      Failed
                         ↓
                    Terminating → Terminated
```

### 4. Namespace Manager

**Responsibility:** Create and manage isolated namespaces per PR

**Implementation Status:** ✅ IMPLEMENTED (Issue #3)

**Implementation:**
```go
// internal/namespace/manager.go
type Manager struct {
    client client.Client
}

func (m *Manager) EnsureNamespace(envName string) (*corev1.Namespace, error) {
    ns := &corev1.Namespace{
        ObjectMeta: metav1.ObjectMeta{
            Name: fmt.Sprintf("preview-%s", envName),
            Labels: map[string]string{
                "previewd.io/managed": "true",
                "previewd.io/env":     envName,
            },
        },
    }
    // Create or update namespace
}
```

**Namespace naming:** `preview-{pr-number}-{sanitized-branch-name}`
**Resource quotas:** Applied per namespace to prevent runaway costs
**Network policies:** Isolate namespaces from each other

### 5. ArgoCD Manager

**Responsibility:** Deploy applications using GitOps

**Integration strategy:**
```go
// internal/argocd/manager.go
type Manager struct {
    argoClient argoclient.Client
}

func (m *Manager) DeployServices(env *PreviewEnvironment) error {
    // Create ArgoCD ApplicationSet for multi-service deployment
    appSet := &argocdv1alpha1.ApplicationSet{
        Spec: argocdv1alpha1.ApplicationSetSpec{
            Generators: []argocdv1alpha1.ApplicationSetGenerator{
                {
                    List: &argocdv1alpha1.ListGenerator{
                        Elements: env.Spec.Services,
                    },
                },
            },
            Template: argocdv1alpha1.ApplicationSetTemplate{
                // Template for each service
            },
        },
    }
    // Apply ApplicationSet
}
```

**Deployment flow:**
1. Create ArgoCD Application or ApplicationSet
2. Point to Git repo (user's app repo)
3. Use overlay for preview-specific config (URLs, DB credentials)
4. Wait for sync (with timeout)
5. Update PreviewEnvironment status

### 6. Ingress Manager

**Responsibility:** Create DNS and TLS for preview environments

**Implementation:**
```go
// internal/ingress/manager.go
type Manager struct {
    client       client.Client
    baseDomain   string
    certIssuer   string
}

func (m *Manager) EnsureIngress(env *PreviewEnvironment) (*networkingv1.Ingress, error) {
    host := fmt.Sprintf("pr-%d.%s", env.Spec.PRNumber, m.baseDomain)

    ingress := &networkingv1.Ingress{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("preview-%s", env.Name),
            Namespace: env.Spec.Namespace,
            Annotations: map[string]string{
                "cert-manager.io/cluster-issuer": m.certIssuer,
                "external-dns.alpha.kubernetes.io/hostname": host,
            },
        },
        Spec: networkingv1.IngressSpec{
            TLS: []networkingv1.IngressTLS{
                {
                    Hosts:      []string{host},
                    SecretName: fmt.Sprintf("tls-%s", env.Name),
                },
            },
            Rules: []networkingv1.IngressRule{
                {
                    Host: host,
                    IngressRuleValue: networkingv1.IngressRuleValue{
                        HTTP: &networkingv1.HTTPIngressRuleValue{
                            Paths: env.Spec.IngressPaths,
                        },
                    },
                },
            },
        },
    }
    // Create or update ingress
}
```

**DNS strategy:**
- Use External-DNS for automatic DNS record creation
- Pattern: `pr-{number}.preview.example.com`
- Wildcard cert: `*.preview.example.com` (or per-PR cert)

### 7. Cost Estimator

**Responsibility:** Calculate and track costs for preview environments

**Implementation Status:** ✅ IMPLEMENTED (Issue #9)

**Core implementation:**
```go
// internal/cost/estimator.go
type Estimator struct {
    client client.Client
    pricing PricingConfig
}

func (e *Estimator) EstimateCost(ctx context.Context, namespace string) (*CostEstimate, error) {
    // 1. Query all pods in namespace
    // 2. Sum CPU and memory requests
    // 3. Calculate costs based on pricing config
    // 4. Return hourly, daily, and monthly estimates
}
```

**Pricing Configuration:**
- CPU: $0.04 per core per hour (configurable)
- Memory: $0.005 per GB per hour (configurable)
- Configured via `--cpu-price-per-hour` and `--memory-price-per-gb-hour` flags

**Cost Optimization Strategies (Future):**

#### 7.1 Resource Sizing
```go
// internal/cost/optimizer.go
type Optimizer struct {
    aiEngine *ai.Engine
}

func (o *Optimizer) DetermineResourceTier(env *PreviewEnvironment) (ResourceTier, error) {
    // AI predicts based on:
    // - PR size (files changed, lines added)
    // - Service complexity
    // - Historical usage patterns

    // Tiers:
    // - minimal:  0.1 CPU,  256Mi RAM per service
    // - small:    0.25 CPU, 512Mi RAM
    // - medium:   0.5 CPU,  1Gi RAM
    // - large:    1 CPU,    2Gi RAM
}
```

#### 7.2 Spot Instances
```go
func (o *Optimizer) ShouldUseSpot(lifespan time.Duration) bool {
    // Use spot instances for short-lived environments (<4h)
    // Use on-demand for longer environments (>4h)
    return lifespan < 4 * time.Hour
}
```

#### 7.3 TTL Management
```go
func (o *Optimizer) DetermineTTL(env *PreviewEnvironment) time.Duration {
    // AI predicts based on:
    // - User's historical PR patterns
    // - PR complexity
    // - Time of day (extend TTL during work hours)

    // Default: 4 hours
    // Max: 7 days (force cleanup)
}
```

### 8. Cleanup Scheduler

**Responsibility:** Automatically delete expired preview environments based on TTL

**Implementation Status:** ✅ IMPLEMENTED (Issue #8)

**Implementation:**
```go
// internal/cleanup/scheduler.go
type Scheduler struct {
    client   client.Client
    interval time.Duration
    logger   logr.Logger
}

func (s *Scheduler) Start(ctx context.Context) error {
    // Run periodic cleanup every interval (default: 5 minutes)
    ticker := time.NewTicker(s.interval)
    for {
        select {
        case <-ticker.C:
            s.cleanupExpired(ctx)
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (s *Scheduler) cleanupExpired(ctx context.Context) {
    // 1. List all PreviewEnvironments
    // 2. Check status.expiresAt vs current time
    // 3. Delete expired environments
    // 4. Respect "do-not-expire" label override
    // 5. Emit Kubernetes events
}
```

**Features:**
- Runs every 5 minutes (configurable via `--cleanup-interval` flag)
- Checks `status.expiresAt` timestamp for each PreviewEnvironment
- Gracefully skips environments with `preview.previewd.io/do-not-expire: "true"` label
- Emits Kubernetes event with reason "TTLExpired" on deletion
- Respects context cancellation for graceful shutdown
- Logs cleanup actions for audit trail

**Expiration Logic:**
- `expiresAt = createdAt + TTL`
- Default TTL: 4 hours
- Maximum TTL: 7 days (configurable)

---

## Custom Resource Definitions

### PreviewEnvironment CRD

**Full specification:**
```yaml
apiVersion: previewd.io/v1alpha1
kind: PreviewEnvironment
metadata:
  name: pr-1234-feature-auth
  labels:
    previewd.io/pr: "1234"
    previewd.io/repo: "myorg-myapp"
spec:
  # GitHub PR metadata
  prNumber: 1234
  repository: "myorg/myapp"
  branch: "feature/new-auth"
  headSHA: "abc123def456"

  # Service deployment
  services:
    - name: auth-service
      image: myorg/auth:pr-1234
      port: 8080
      replicas: 1
      autoDetected: true  # AI detected vs user-specified

    - name: user-service
      image: myorg/users:pr-1234
      port: 8080
      replicas: 1
      autoDetected: true

  # Test data configuration
  testData:
    enabled: true
    strategy: synthetic  # synthetic | production-snapshot | minimal
    aiModel: gpt-4       # or: gpt-3.5-turbo, ollama-mistral
    generation:
      users: 100
      orders: 500
      products: 50
    seed: 12345          # For reproducibility

  # Testing configuration
  tests:
    enabled: true
    framework: pytest
    selector: ai-smart   # ai-smart | all | changed-files
    timeout: 10m
    parallel: true

  # Resource management
  resources:
    tier: small          # minimal | small | medium | large
    useSpot: true        # Use spot instances

  # Lifecycle
  ttl: 4h              # Time-to-live
  autoExtend: true     # AI extends TTL if PR active
  maxLifespan: 7d      # Force cleanup after this

  # Cost optimization
  costBudget:
    daily: "$5.00"
    total: "$50.00"

  # Ingress configuration
  ingress:
    enabled: true
    host: ""             # Auto-generated: pr-1234.preview.example.com
    tls: true

status:
  # Environment state
  phase: Ready         # Pending | Provisioning | Ready | Testing | Complete | Failed | Terminating | Terminated

  # URLs
  url: "https://pr-1234.preview.myapp.com"
  urls:
    - service: auth-service
      url: "https://pr-1234.preview.myapp.com/auth"
    - service: user-service
      url: "https://pr-1234.preview.myapp.com/users"

  # Test results
  tests:
    total: 42
    passed: 41
    failed: 1
    skipped: 0
    duration: 2m15s
    failedTests:
      - name: test_user_creation
        reason: "Timeout after 30s"

  # Cost tracking
  costs:
    estimatedDaily: "$2.34"
    actualToDate: "$8.45"
    breakdown:
      compute: "$6.00"
      storage: "$1.45"
      network: "$1.00"

  # AI metadata
  ai:
    codeAnalysisUsed: true
    servicesDetected: ["auth-service", "user-service"]
    confidence: 0.95
    dataGenerated: true
    lifespanPredicted: 6h

  # Conditions
  conditions:
    - type: NamespaceReady
      status: "True"
      lastTransitionTime: "2025-11-08T10:00:00Z"

    - type: ServicesDeployed
      status: "True"
      lastTransitionTime: "2025-11-08T10:02:00Z"

    - type: TestsCompleted
      status: "True"
      lastTransitionTime: "2025-11-08T10:04:00Z"

    - type: Ready
      status: "True"
      lastTransitionTime: "2025-11-08T10:04:15Z"

  # Timestamps
  createdAt: "2025-11-08T10:00:00Z"
  readyAt: "2025-11-08T10:04:15Z"
  deletedAt: ""
```

---

## Reconciliation Loop

### Reconciliation Logic

```go
func (r *PreviewEnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // 1. Fetch PreviewEnvironment
    env := &previewdv1alpha1.PreviewEnvironment{}
    if err := r.Get(ctx, req.NamespacedName, env); err != nil {
        if errors.IsNotFound(err) {
            return ctrl.Result{}, nil // Deleted
        }
        return ctrl.Result{}, err
    }

    // 2. Handle deletion (finalizers)
    if !env.ObjectMeta.DeletionTimestamp.IsZero() {
        return r.handleDeletion(ctx, env)
    }

    // 3. Ensure finalizer exists
    if !controllerutil.ContainsFinalizer(env, finalizerName) {
        controllerutil.AddFinalizer(env, finalizerName)
        return ctrl.Result{}, r.Update(ctx, env)
    }

    // 4. Reconcile based on phase
    switch env.Status.Phase {
    case "":
        return r.reconcilePending(ctx, env)
    case "Pending":
        return r.reconcileProvisioning(ctx, env)
    case "Provisioning":
        return r.reconcileReady(ctx, env)
    case "Ready":
        return r.reconcileTesting(ctx, env)
    case "Testing":
        return r.reconcileComplete(ctx, env)
    case "Complete":
        return r.reconcileTTL(ctx, env)
    default:
        return ctrl.Result{}, fmt.Errorf("unknown phase: %s", env.Status.Phase)
    }
}
```

### Phase Transitions

**Pending → Provisioning:**
1. AI analyzes code (if enabled)
2. Determine services to deploy
3. Create namespace
4. Update status to "Provisioning"

**Provisioning → Ready:**
1. Deploy services via ArgoCD
2. Create ingress
3. Wait for pods to be Ready
4. Generate test data (if enabled)
5. Update status to "Ready"

**Ready → Testing:**
1. Run integration tests
2. Collect results
3. Update status to "Testing"

**Testing → Complete:**
1. Post results to PR comment
2. Update status to "Complete"

**Complete → Terminated (TTL expired):**
1. Delete namespace
2. Delete ArgoCD applications
3. Delete DNS records
4. Update status to "Terminated"
5. Remove finalizer

---

## AI Integration

### OpenAI Integration

**Client configuration:**
```go
// internal/ai/client.go
type Client struct {
    apiKey     string
    httpClient *http.Client
    cache      *cache.Cache
    rateLimit  *rate.Limiter
}

func NewClient(apiKey string) *Client {
    return &Client{
        apiKey:     apiKey,
        httpClient: &http.Client{Timeout: 30 * time.Second},
        cache:      cache.New(24*time.Hour, 1*time.Hour),
        rateLimit:  rate.NewLimiter(rate.Every(time.Second), 10), // 10 req/sec
    }
}
```

**Prompt templates:**
```go
const codeAnalysisPrompt = `Analyze this pull request diff and determine which services are affected.

Repository: {{.Repository}}
PR: #{{.PRNumber}}
Branch: {{.Branch}}
Files changed: {{.FilesChanged}}

Diff:
{{.Diff}}

Task: Identify all services that are directly modified or depend on changed code.
Consider:
- Imports and package references
- API calls between services
- Database queries
- Environment variable references
- Configuration files

Return a JSON array of service names only.
Example: ["auth-service", "user-service", "payment-service"]

Service names:
`
```

### Local LLM Support (Ollama)

**Optional local LLM for on-prem deployments:**
```go
// internal/ai/ollama.go
type OllamaClient struct {
    baseURL string
    model   string
}

func (c *OllamaClient) Analyze(prompt string) (string, error) {
    // Call local Ollama instance
    // Same interface as OpenAI client
}
```

**Configuration:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: previewd-config
data:
  ai.provider: "ollama"         # or "openai"
  ai.baseURL: "http://ollama:11434"
  ai.model: "mistral"
```

---

## Cost Optimization

### Resource Quotas

**Per-namespace quotas:**
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: preview-quota
  namespace: preview-pr-1234
spec:
  hard:
    requests.cpu: "2"
    requests.memory: "4Gi"
    requests.storage: "10Gi"
    persistentvolumeclaims: "5"
    pods: "20"
```

### Cost Tracking

**Cost calculation:**
```go
// internal/cost/calculator.go
type Calculator struct {
    priceList *PriceList
}

type Cost struct {
    Compute  float64
    Storage  float64
    Network  float64
    Total    float64
}

func (c *Calculator) EstimateDailyCost(env *PreviewEnvironment) (*Cost, error) {
    cost := &Cost{}

    // Compute cost (based on pod resources)
    for _, service := range env.Spec.Services {
        cpu := service.Resources.Requests.CPU
        memory := service.Resources.Requests.Memory

        cost.Compute += c.priceList.ComputeHourly(cpu, memory) * 24
    }

    // Storage cost (PVCs)
    for _, pvc := range env.Status.PVCs {
        cost.Storage += c.priceList.StorageDaily(pvc.Size)
    }

    // Network cost (ingress/egress)
    cost.Network = c.estimateNetworkCost(env)

    cost.Total = cost.Compute + cost.Storage + cost.Network
    return cost, nil
}
```

---

## Security Model

### RBAC

**Operator service account:**
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: previewd-operator
  namespace: previewd-system

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: previewd-operator
rules:
  # PreviewEnvironment CRD
  - apiGroups: ["previewd.io"]
    resources: ["previewenvironments"]
    verbs: ["*"]

  # Namespace management
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["create", "get", "list", "delete"]

  # Service deployment
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets"]
    verbs: ["create", "get", "list", "update", "delete"]

  # ArgoCD integration
  - apiGroups: ["argoproj.io"]
    resources: ["applications", "applicationsets"]
    verbs: ["create", "get", "list", "update", "delete"]

  # Ingress
  - apiGroups: ["networking.k8s.io"]
    resources: ["ingresses"]
    verbs: ["create", "get", "list", "update", "delete"]
```

### Secrets Management

**AI API keys:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: previewd-ai-credentials
  namespace: previewd-system
type: Opaque
data:
  openai-api-key: <base64-encoded-key>
```

**GitHub webhook secret:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-webhook-secret
  namespace: previewd-system
type: Opaque
data:
  secret: <base64-encoded-secret>
```

---

## Deployment Architecture

### Operator Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: previewd-operator
  namespace: previewd-system
spec:
  replicas: 1  # Single replica (leader election for HA)
  selector:
    matchLabels:
      app: previewd-operator
  template:
    metadata:
      labels:
        app: previewd-operator
    spec:
      serviceAccountName: previewd-operator
      containers:
        - name: manager
          image: mikelane/previewd:latest
          command:
            - /manager
          env:
            - name: ENABLE_AI
              value: "true"
            - name: AI_PROVIDER
              value: "openai"
            - name: OPENAI_API_KEY
              valueFrom:
                secretKeyRef:
                  name: previewd-ai-credentials
                  key: openai-api-key
          resources:
            limits:
              cpu: 500m
              memory: 512Mi
            requests:
              cpu: 100m
              memory: 128Mi
```

### Webhook Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: previewd-webhook
  namespace: previewd-system
spec:
  type: LoadBalancer
  selector:
    app: previewd-operator
  ports:
    - name: webhook
      port: 443
      targetPort: 9443
      protocol: TCP
```

---

## Scalability & Performance

### Expected Load

**Small deployment:**
- 100 PRs/day
- 50 concurrent preview environments
- 5-10 services per environment

**Medium deployment:**
- 500 PRs/day
- 200 concurrent preview environments
- 10-20 services per environment

**Large deployment:**
- 2000+ PRs/day
- 1000+ concurrent preview environments
- 20+ services per environment

### Performance Targets

- **Environment creation:** <2 minutes (from PR open to ready)
- **Test execution:** <5 minutes
- **Environment deletion:** <30 seconds
- **API response time:** <200ms (p95)
- **Reconciliation loop:** <10 seconds

### Scaling Strategies

1. **Horizontal operator scaling** - Multiple replicas with leader election
2. **Shard by repository** - Different operators for different repos
3. **Async processing** - Queue webhook events
4. **Cache AI responses** - Reduce API calls
5. **Batch operations** - Group similar reconciliations

---

This architecture provides a solid foundation for building Previewd. The design is modular, testable, and extensible for future enhancements.
