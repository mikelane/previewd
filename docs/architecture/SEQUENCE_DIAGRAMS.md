# Sequence Diagrams

## Overview

This document contains detailed sequence diagrams for key workflows in Previewd. These diagrams illustrate the interactions between components during critical operations.

## 1. PR Opened → Preview Environment Created

### Happy Path

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant GH as GitHub
    participant WH as Webhook Server
    participant K8s as Kubernetes API
    participant Ctrl as Controller
    participant GHC as GitHub Client
    participant AI as AI Engine
    participant AC as ArgoCD Manager
    participant DNS as external-dns
    participant CM as cert-manager

    Dev->>GH: Open Pull Request #123
    Note over Dev,GH: PR contains code changes<br/>to auth and api services

    GH->>WH: POST /webhook
    Note right of GH: Headers:<br/>X-GitHub-Event: pull_request<br/>X-Hub-Signature-256: sha256=...

    WH->>WH: Validate HMAC signature
    WH->>WH: Parse payload

    WH->>K8s: Create PreviewEnvironment CR
    Note right of WH: name: pr-123<br/>spec.prNumber: 123<br/>spec.headSHA: abc123

    K8s-->>WH: 201 Created
    WH->>GH: 201 Created
    GH-->>Dev: Notification: Checks running

    Note over K8s,Ctrl: Controller watches PreviewEnvironment CRs

    K8s->>Ctrl: Watch Event: ADDED
    Ctrl->>Ctrl: Reconcile(pr-123)

    Ctrl->>K8s: Get PreviewEnvironment CR
    K8s-->>Ctrl: CR data

    Ctrl->>K8s: Add finalizer
    Note right of Ctrl: Finalizer:<br/>preview.previewd.io/finalizer

    Ctrl->>K8s: Update status: Pending → Creating
    Note right of Ctrl: status.phase = Creating<br/>condition: Deploying = True

    Ctrl->>GHC: FetchDiff(company/monorepo, 123)
    GHC->>GH: GET /repos/company/monorepo/pulls/123.diff
    GH-->>GHC: Unified diff content
    GHC-->>Ctrl: diff string

    alt AI Enabled
        Ctrl->>AI: DetectServices(diff)
        AI->>AI: Check cache (hash: abc123)
        Note right of AI: Cache miss

        AI->>AI: Call OpenAI API
        Note right of AI: Prompt: Analyze this diff...<br/>Response: ["auth", "api"]

        AI->>AI: Cache response (1h TTL)
        AI-->>Ctrl: services: [auth, api]
    else Static Config
        Ctrl->>Ctrl: Use spec.services
        Note right of Ctrl: services: [auth, api]
    end

    Ctrl->>K8s: Create Namespace (pr-123)
    Note right of Ctrl: labels:<br/>preview.previewd.io/pr: "123"<br/>ownerReferences: [PreviewEnvironment]

    K8s-->>Ctrl: Namespace created

    Ctrl->>K8s: Create ResourceQuota
    Note right of Ctrl: limits:<br/>cpu: 2, memory: 4Gi

    Ctrl->>K8s: Create NetworkPolicy
    Note right of Ctrl: Default deny + allow ingress

    Ctrl->>AC: BuildApplicationSet(pr-123, [auth, api])
    AC->>AC: Generate ApplicationSet manifest
    AC-->>Ctrl: ApplicationSet object

    Ctrl->>K8s: Create ApplicationSet (argocd ns)
    Note right of Ctrl: Generates Applications:<br/>preview-123-auth<br/>preview-123-api

    K8s-->>Ctrl: ApplicationSet created

    Note over K8s: ArgoCD ApplicationSet controller<br/>generates Application CRs

    Note over K8s: ArgoCD syncs Applications<br/>(pulls from Git, deploys to pr-123 ns)

    Ctrl->>K8s: Create Ingress (pr-123 ns)
    Note right of Ctrl: host: pr-123.preview.company.com<br/>annotations:<br/>cert-manager.io/cluster-issuer<br/>external-dns.alpha.../hostname

    K8s-->>Ctrl: Ingress created

    Note over DNS: external-dns watches Ingress

    DNS->>K8s: Get Ingress annotations
    K8s-->>DNS: hostname: pr-123.preview.company.com
    DNS->>DNS: Create DNS A record
    Note right of DNS: pr-123.preview.company.com<br/>→ LoadBalancer IP

    Note over CM: cert-manager watches Ingress

    CM->>K8s: Get Ingress TLS config
    K8s-->>CM: secretName: pr-123-tls
    CM->>CM: Request Let's Encrypt cert
    CM->>K8s: Create Secret (pr-123-tls)

    Ctrl->>K8s: Watch Applications (pr-123-auth, pr-123-api)
    loop Poll Application status (every 10s)
        K8s-->>Ctrl: Application.status.health: Progressing
        Note right of Ctrl: Wait for health: Healthy
    end

    K8s-->>Ctrl: All Applications: Healthy + Synced

    Ctrl->>Ctrl: Calculate cost
    Note right of Ctrl: CPU: 1.5 cores<br/>Memory: 3Gi<br/>Hourly: $0.075

    Ctrl->>K8s: Update PreviewEnvironment status
    Note right of Ctrl: phase: Ready<br/>url: https://pr-123....<br/>services: [{auth, ready}, {api, ready}]<br/>costEstimate: {...}<br/>expiresAt: now + 4h

    K8s-->>Ctrl: Status updated

    Ctrl->>GHC: UpdateCommitStatus(abc123, success)
    GHC->>GH: POST /repos/.../statuses/abc123
    Note right of GHC: state: success<br/>target_url: https://pr-123...<br/>context: previewd<br/>description: Preview ready

    GH-->>GHC: 201 Created
    GHC-->>Ctrl: Status updated

    GH->>Dev: Notification: Preview ready
    Note right of GH: Status check: ✅ previewd<br/>View deployment

    Dev->>Dev: Click preview URL
    Note right of Dev: Opens https://pr-123.preview.company.com
```

### Error Scenario: GitHub API Failure

```mermaid
sequenceDiagram
    participant Ctrl as Controller
    participant GHC as GitHub Client
    participant GH as GitHub
    participant K8s as Kubernetes API

    Ctrl->>GHC: FetchDiff(company/monorepo, 123)
    GHC->>GH: GET /repos/company/monorepo/pulls/123.diff
    GH-->>GHC: 500 Internal Server Error

    GHC-->>Ctrl: Error: API unavailable

    Ctrl->>K8s: Update PreviewEnvironment status
    Note right of Ctrl: phase: Failed<br/>condition:<br/>type: Failed<br/>status: True<br/>reason: GitHubAPIError<br/>message: Failed to fetch PR diff

    Ctrl->>K8s: Emit Event (Warning)
    Note right of Ctrl: reason: GitHubAPIError<br/>message: Will retry in 1m

    Ctrl->>Ctrl: Return error (retry with backoff)
    Note right of Ctrl: Requeue after 1m, 2m, 4m...<br/>Max retries: 5

    Note over Ctrl: Next reconcile attempt (1m later)

    Ctrl->>GHC: FetchDiff (retry)
    GHC->>GH: GET /repos/company/monorepo/pulls/123.diff
    GH-->>GHC: 200 OK (diff content)

    Ctrl->>Ctrl: Continue reconciliation
    Note right of Ctrl: Clear Failed condition<br/>Proceed with deployment
```

---

## 2. PR Updated → Environment Updated

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant GH as GitHub
    participant WH as Webhook Server
    participant K8s as Kubernetes API
    participant Ctrl as Controller
    participant AC as ArgoCD Manager
    participant GHC as GitHub Client

    Dev->>GH: Push new commit to PR #123
    Note over Dev,GH: New commit: def456<br/>(fixes bug in auth service)

    GH->>WH: POST /webhook
    Note right of GH: X-GitHub-Event: pull_request<br/>action: synchronize<br/>head.sha: def456

    WH->>WH: Validate signature + parse

    WH->>K8s: Get PreviewEnvironment (pr-123)
    K8s-->>WH: Current CR

    WH->>K8s: Patch PreviewEnvironment
    Note right of WH: spec.headSHA: abc123 → def456

    K8s-->>WH: 200 OK
    WH->>GH: 200 OK

    K8s->>Ctrl: Watch Event: MODIFIED
    Ctrl->>Ctrl: Reconcile(pr-123)

    Ctrl->>K8s: Get PreviewEnvironment
    K8s-->>Ctrl: CR with new headSHA

    Ctrl->>K8s: Update status: Ready → Updating
    Note right of Ctrl: phase: Updating<br/>condition: Deploying = True

    Ctrl->>GHC: UpdateCommitStatus(def456, pending)
    GHC->>GH: POST /repos/.../statuses/def456
    Note right of GHC: state: pending<br/>description: Updating preview...

    Ctrl->>AC: UpdateApplicationSet(pr-123, new SHA)
    AC->>K8s: Get ApplicationSet (preview-123)
    K8s-->>AC: Current ApplicationSet

    AC->>AC: Update image tags
    Note right of AC: Images:<br/>auth: company/auth:def456<br/>api: company/api:def456

    AC->>K8s: Update ApplicationSet
    K8s-->>AC: Updated

    Note over K8s: ArgoCD detects change

    Note over K8s: ArgoCD syncs new images<br/>(rolling update)

    loop Poll Application status
        Ctrl->>K8s: Get Applications
        K8s-->>Ctrl: health: Progressing
        Note right of Ctrl: Waiting for new pods
    end

    K8s-->>Ctrl: All Applications: Healthy + Synced

    Ctrl->>K8s: Update PreviewEnvironment status
    Note right of Ctrl: phase: Updating → Ready<br/>lastSyncedAt: now<br/>expiresAt: extended by +4h

    Ctrl->>GHC: UpdateCommitStatus(def456, success)
    GHC->>GH: POST /repos/.../statuses/def456
    Note right of GHC: state: success<br/>target_url: (same)<br/>description: Preview updated

    GH->>Dev: Notification: Preview updated
```

---

## 3. PR Closed → Environment Destroyed

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant GH as GitHub
    participant WH as Webhook Server
    participant K8s as Kubernetes API
    participant Ctrl as Controller
    participant AC as ArgoCD
    participant DNS as external-dns
    participant GHC as GitHub Client

    Dev->>GH: Close/Merge Pull Request #123
    Note over Dev,GH: PR merged to main

    GH->>WH: POST /webhook
    Note right of GH: X-GitHub-Event: pull_request<br/>action: closed

    WH->>WH: Validate signature + parse

    WH->>K8s: Get PreviewEnvironment (pr-123)
    K8s-->>WH: Current CR

    WH->>K8s: Delete PreviewEnvironment (pr-123)
    Note right of WH: Sets deletionTimestamp

    K8s-->>WH: 200 OK (deletion initiated)
    WH->>GH: 200 OK

    K8s->>Ctrl: Watch Event: MODIFIED (deletion timestamp set)
    Ctrl->>Ctrl: Reconcile(pr-123) - Delete path

    Ctrl->>K8s: Get PreviewEnvironment
    K8s-->>Ctrl: CR with deletionTimestamp

    Ctrl->>K8s: Update status: Ready → Deleting
    Note right of Ctrl: phase: Deleting

    Ctrl->>GHC: UpdateCommitStatus(abc123, failure)
    GHC->>GH: POST /repos/.../statuses/abc123
    Note right of GHC: state: failure<br/>description: Preview destroyed<br/>(failure = inactive)

    Ctrl->>Ctrl: Run finalizers
    Note right of Ctrl: Finalizer:<br/>preview.previewd.io/finalizer

    Ctrl->>K8s: Get ApplicationSet (preview-123)
    K8s-->>Ctrl: ApplicationSet object

    Ctrl->>K8s: Delete ApplicationSet
    Note right of Ctrl: Owner reference cascade:<br/>- Applications deleted<br/>- Deployments deleted<br/>- Services deleted

    K8s-->>Ctrl: ApplicationSet deleted

    Note over AC: ArgoCD detects Application deletion

    AC->>K8s: Delete resources in pr-123 ns
    Note right of AC: Prune all synced resources

    K8s-->>AC: Resources deleted

    Note over DNS: external-dns detects Ingress deletion

    DNS->>K8s: Get Ingress (pr-123 ns)
    K8s-->>DNS: Not found

    DNS->>DNS: Delete DNS A record
    Note right of DNS: Remove:<br/>pr-123.preview.company.com

    Ctrl->>K8s: Delete Namespace (pr-123)
    Note right of Ctrl: Cascade delete:<br/>- All pods<br/>- All services<br/>- Ingress<br/>- ResourceQuota<br/>- NetworkPolicy

    K8s-->>Ctrl: Namespace deleted

    Ctrl->>K8s: Remove finalizer from CR
    Note right of Ctrl: Finalizers: []

    Ctrl->>K8s: Update PreviewEnvironment
    K8s-->>Ctrl: CR updated

    Note over K8s: No finalizers remaining

    K8s->>K8s: Delete PreviewEnvironment CR
    Note right of K8s: CR completely removed

    K8s->>Ctrl: Watch Event: DELETED
    Ctrl->>Ctrl: No action needed
    Note right of Ctrl: Reconciliation complete
```

---

## 4. TTL Expired → Automatic Cleanup

```mermaid
sequenceDiagram
    participant Sched as Cleanup Scheduler
    participant K8s as Kubernetes API
    participant Ctrl as Controller
    participant GHC as GitHub Client
    participant GH as GitHub

    Note over Sched: Runs every 5 minutes

    Sched->>K8s: List all PreviewEnvironments
    K8s-->>Sched: [pr-123, pr-456, pr-789]

    loop For each PreviewEnvironment
        Sched->>Sched: Check TTL
        Note right of Sched: pr-123:<br/>createdAt: 4.5 hours ago<br/>TTL: 4h<br/>expiresAt: 30 min ago

        alt TTL Expired
            Sched->>Sched: Check labels
            Note right of Sched: Labels:<br/>do-not-expire: not set<br/>✅ Can delete

            Sched->>K8s: Delete PreviewEnvironment (pr-123)
            Note right of Sched: Sets deletionTimestamp

            K8s-->>Sched: Deletion initiated

            K8s->>Ctrl: Watch Event: MODIFIED (deletion timestamp)
            Note over Ctrl: Standard deletion flow<br/>(see sequence #3)

        else TTL Not Expired
            Sched->>Sched: Skip (pr-456)
            Note right of Sched: pr-456:<br/>expiresAt: 2 hours from now<br/>❌ Skip
        end

        alt Has "do-not-expire" Label
            Sched->>Sched: Skip (pr-789)
            Note right of Sched: pr-789:<br/>labels.do-not-expire: "true"<br/>❌ Skip (manual override)
        end
    end

    Sched->>GHC: UpdateCommitStatus (expired PRs)
    loop For each deleted preview
        GHC->>GH: POST /repos/.../statuses/{sha}
        Note right of GHC: state: failure<br/>description: Preview expired (TTL: 4h)
    end

    Note over Sched: Sleep 5 minutes
```

---

## 5. AI-Powered Service Detection (v0.2.0+)

```mermaid
sequenceDiagram
    participant Ctrl as Controller
    participant GHC as GitHub Client
    participant GH as GitHub
    participant AI as AI Engine
    participant Cache as Cache
    participant OpenAI as OpenAI API

    Ctrl->>GHC: FetchDiff(company/monorepo, 123)
    GHC->>GH: GET /repos/company/monorepo/pulls/123.diff
    GH-->>GHC: Unified diff (2000 lines)
    GHC-->>Ctrl: diff string

    Ctrl->>AI: DetectServices(diff)

    AI->>AI: Hash diff content
    Note right of AI: SHA256 hash:<br/>abc123...

    AI->>Cache: Get(abc123)
    Cache-->>AI: Cache miss

    AI->>AI: Build prompt
    Note right of AI: Prompt template:<br/>You are analyzing a code diff...<br/><diff content><br/>Task: List affected services...

    AI->>AI: Truncate diff if needed
    Note right of AI: Max tokens: 4000<br/>Truncated: false

    AI->>OpenAI: CreateChatCompletion
    Note right of AI: model: gpt-4<br/>temperature: 0.0<br/>max_tokens: 1000

    OpenAI->>OpenAI: Analyze code diff
    Note right of OpenAI: Detects changes in:<br/>- auth/ directory<br/>- api/ directory<br/>- lib/ (shared, ignored)

    OpenAI-->>AI: Response (JSON)
    Note right of OpenAI: {"services": ["auth", "api"]}

    AI->>AI: Parse JSON response
    AI->>AI: Validate service names
    Note right of AI: Ensure services exist<br/>in known list

    AI->>Cache: Set(abc123, [auth, api], TTL: 1h)
    Cache-->>AI: Cached

    AI->>AI: Emit metrics
    Note right of AI: previewd_ai_requests_total{<br/>  cache_hit=false}<br/>previewd_ai_cost_usd_total +0.02

    AI-->>Ctrl: services: [auth, api]

    Note over Ctrl: Continue with deployment

    Note over Ctrl,Cache: Future request with same diff

    Ctrl->>AI: DetectServices(same diff)
    AI->>AI: Hash diff (same hash: abc123)
    AI->>Cache: Get(abc123)
    Cache-->>AI: [auth, api] (cached)

    AI->>AI: Emit metrics
    Note right of AI: previewd_ai_requests_total{<br/>  cache_hit=true}

    AI-->>Ctrl: services: [auth, api] (from cache)
    Note right of AI: No OpenAI API call<br/>Cost: $0.00
```

### AI Error Handling

```mermaid
sequenceDiagram
    participant Ctrl as Controller
    participant AI as AI Engine
    participant OpenAI as OpenAI API
    participant Config as ConfigMap

    Ctrl->>AI: DetectServices(diff)

    AI->>OpenAI: CreateChatCompletion
    OpenAI-->>AI: 429 Rate Limit Exceeded

    AI->>AI: Log error
    Note right of AI: level: warn<br/>msg: OpenAI rate limited<br/>fallback: static config

    AI->>Config: Get default services
    Config-->>AI: services: [auth, api, frontend]

    AI-->>Ctrl: services: [auth, api, frontend] (fallback)
    Note right of AI: fallback_reason: rate_limit

    Ctrl->>Ctrl: Continue deployment
    Note right of Ctrl: Uses static config<br/>No user-visible impact
```

---

## 6. Multi-Service Deployment

### Parallel Service Creation

```mermaid
sequenceDiagram
    participant Ctrl as Controller
    participant K8s as Kubernetes API
    participant AC as ArgoCD
    participant AS as ApplicationSet Controller
    participant App1 as Application: auth
    participant App2 as Application: api
    participant App3 as Application: frontend

    Ctrl->>K8s: Create ApplicationSet (preview-123)
    Note right of Ctrl: generators:<br/>- {service: auth}<br/>- {service: api}<br/>- {service: frontend}

    K8s-->>Ctrl: ApplicationSet created

    Note over AS: ApplicationSet controller<br/>watches for changes

    K8s->>AS: Watch Event: ADDED (ApplicationSet)
    AS->>AS: Generate Applications

    par Generate auth Application
        AS->>K8s: Create Application (preview-123-auth)
        K8s-->>AS: Created
    and Generate api Application
        AS->>K8s: Create Application (preview-123-api)
        K8s-->>AS: Created
    and Generate frontend Application
        AS->>K8s: Create Application (preview-123-frontend)
        K8s-->>AS: Created
    end

    Note over AC: ArgoCD Application controllers<br/>sync in parallel

    par Sync auth
        AC->>App1: Reconcile Application
        App1->>App1: Git fetch
        App1->>K8s: Create Deployment (pr-123-auth)
        App1->>K8s: Create Service (pr-123-auth)
    and Sync api
        AC->>App2: Reconcile Application
        App2->>App2: Git fetch
        App2->>K8s: Create Deployment (pr-123-api)
        App2->>K8s: Create Service (pr-123-api)
    and Sync frontend
        AC->>App3: Reconcile Application
        App3->>App3: Git fetch
        App3->>K8s: Create Deployment (pr-123-frontend)
        App3->>K8s: Create Service (pr-123-frontend)
    end

    Note over K8s: All pods starting

    par Wait for auth
        loop Poll auth status
            Ctrl->>App1: Get Application.status
            App1-->>Ctrl: health: Progressing
        end
        App1-->>Ctrl: health: Healthy
    and Wait for api
        loop Poll api status
            Ctrl->>App2: Get Application.status
            App2-->>Ctrl: health: Progressing
        end
        App2-->>Ctrl: health: Healthy
    and Wait for frontend
        loop Poll frontend status
            Ctrl->>App3: Get Application.status
            App3-->>Ctrl: health: Progressing
        end
        App3-->>Ctrl: health: Healthy
    end

    Ctrl->>K8s: Update PreviewEnvironment status
    Note right of Ctrl: services:<br/>- {auth, ready: true}<br/>- {api, ready: true}<br/>- {frontend, ready: true}<br/>phase: Ready
```

---

## 7. Failure Recovery: ArgoCD Sync Failure

```mermaid
sequenceDiagram
    participant Ctrl as Controller
    participant K8s as Kubernetes API
    participant App as Application: auth
    participant Git as Git Repository

    Note over Ctrl: Environment creating

    Ctrl->>App: Watch Application status
    App->>Git: Fetch manifests
    Git-->>App: 404 Not Found (path invalid)

    App->>App: Sync failed
    Note right of App: status.sync.status: Unknown<br/>status.conditions:<br/>- SyncError

    App-->>Ctrl: health: Missing, sync: Unknown

    Ctrl->>K8s: Update PreviewEnvironment status
    Note right of Ctrl: phase: Failed<br/>condition:<br/>type: Failed<br/>status: True<br/>reason: SyncFailed<br/>message: Application auth<br/>  failed to sync

    Ctrl->>K8s: Emit Event (Warning)
    Note right of Ctrl: reason: SyncFailed<br/>message: Check Application<br/>  preview-123-auth

    Ctrl->>Ctrl: Requeue with backoff
    Note right of Ctrl: Retry in 1m

    Note over Dev: Developer fixes Git path

    Note over Ctrl: Next reconcile (1m later)

    Ctrl->>App: Watch Application status
    App->>Git: Fetch manifests (retry)
    Git-->>App: 200 OK (manifests)

    App->>K8s: Create Deployment
    App->>K8s: Create Service

    App-->>Ctrl: health: Healthy, sync: Synced

    Ctrl->>K8s: Update PreviewEnvironment status
    Note right of Ctrl: phase: Failed → Ready<br/>Clear Failed condition
```

---

## 8. Cost Calculation Flow

```mermaid
sequenceDiagram
    participant Ctrl as Controller
    participant K8s as Kubernetes API
    participant CE as Cost Estimator
    participant Prom as Prometheus

    Ctrl->>CE: EstimateCost(preview: pr-123)

    CE->>K8s: List Pods (namespace: pr-123)
    K8s-->>CE: [pod-auth, pod-api, pod-frontend]

    loop For each pod
        CE->>CE: Sum container resources
        Note right of CE: pod-auth:<br/>cpu.requests: 100m<br/>memory.requests: 256Mi
    end

    CE->>CE: Calculate totals
    Note right of CE: Total CPU: 300m (0.3 cores)<br/>Total Memory: 768Mi (0.75 GB)

    CE->>CE: Apply pricing
    Note right of CE: CPU: 0.3 × $0.04/hr = $0.012<br/>Memory: 0.75 × $0.005/hr = $0.004<br/>Total: $0.016/hr

    CE->>CE: Project costs
    Note right of CE: Hourly: $0.016<br/>Daily: $0.384<br/>Monthly: $11.52

    CE-->>Ctrl: CostEstimate{<br/>  cpu: "300m",<br/>  memory: "768Mi",<br/>  hourlyCost: 0.016,<br/>  dailyCost: 0.384,<br/>  monthlyCost: 11.52<br/>}

    Ctrl->>K8s: Update PreviewEnvironment status
    Note right of Ctrl: status.costEstimate: {...}

    Ctrl->>Prom: Emit metric
    Note right of Ctrl: previewd_environment_cost_<br/>  estimate_usd{pr="123"} 0.016

    Note over Prom: Dashboard queries<br/>sum(previewd_environment_cost_<br/>  estimate_usd) = total cost
```

---

## 9. Leader Election (HA Setup)

```mermaid
sequenceDiagram
    participant Pod1 as Controller Pod 1
    participant Pod2 as Controller Pod 2
    participant Pod3 as Controller Pod 3
    participant K8s as Kubernetes API
    participant Lease as Lease (previewd-leader)

    Note over Pod1,Pod3: All pods start simultaneously

    par Pod1 tries to acquire lease
        Pod1->>K8s: Get Lease (previewd-leader)
        K8s-->>Pod1: Not found
        Pod1->>K8s: Create Lease (holder: pod1)
        K8s-->>Pod1: Lease created ✅
        Pod1->>Pod1: Become leader
        Note right of Pod1: Start reconciliation
    and Pod2 tries to acquire lease
        Pod2->>K8s: Get Lease (previewd-leader)
        K8s-->>Pod2: Lease exists (holder: pod1)
        Pod2->>Pod2: Become follower
        Note right of Pod2: Wait and watch
    and Pod3 tries to acquire lease
        Pod3->>K8s: Get Lease (previewd-leader)
        K8s-->>Pod3: Lease exists (holder: pod1)
        Pod3->>Pod3: Become follower
        Note right of Pod3: Wait and watch
    end

    loop Every 10s (renew lease)
        Pod1->>K8s: Update Lease (renew)
        K8s-->>Pod1: Lease renewed
    end

    Note over Pod1: Pod1 crashes

    Pod1->>Pod1: ❌ Crashed

    Pod2->>K8s: Get Lease
    K8s-->>Pod2: Lease expired (not renewed)

    Pod2->>K8s: Update Lease (holder: pod2)
    K8s-->>Pod2: Lease acquired ✅

    Pod2->>Pod2: Become leader
    Note right of Pod2: Start reconciliation<br/>(seamless takeover)

    Pod3->>K8s: Get Lease
    K8s-->>Pod3: Lease exists (holder: pod2)
    Pod3->>Pod3: Remain follower
```

---

**Document Status**: ✅ Complete
**Last Updated**: 2025-11-09
**Authors**: Mike Lane (@mikelane)
