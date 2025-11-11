# Previewd Architecture Documentation

## Overview

This directory contains comprehensive architecture documentation for Previewd, a Kubernetes operator that automatically creates, manages, and destroys preview environments for pull requests.

**Last Updated**: 2025-11-09
**Version**: v0.1.0 (in development)
**Status**: Active Development

---

## Document Index

### Core Architecture

1. **[System Architecture](SYSTEM_ARCHITECTURE.md)** - High-level system design
   - Overview and design philosophy
   - Component interactions and data flow
   - Integration with external systems (GitHub, ArgoCD, DNS, AI)
   - Scalability and performance considerations
   - Security architecture
   - Observability and monitoring
   - Cost optimization strategies

2. **[Component Design](COMPONENT_DESIGN.md)** - Detailed component specifications
   - Package structure and organization
   - Controller reconciliation logic
   - Webhook server implementation
   - GitHub client integration
   - ArgoCD integration patterns
   - AI engine design (v0.2.0+)
   - Cost estimator implementation
   - Testing strategies

3. **[Data Model](DATA_MODEL.md)** - CRD schema and data structures
   - PreviewEnvironment custom resource definition
   - Status conditions and phase lifecycle
   - Related Kubernetes resources (Namespace, ApplicationSet, Ingress)
   - Resource relationships and ownership
   - Validation rules and admission webhooks
   - Indexing and query patterns

4. **[Sequence Diagrams](SEQUENCE_DIAGRAMS.md)** - Key workflow interactions
   - PR opened → preview environment created
   - PR updated → environment updated
   - PR closed → environment destroyed
   - TTL expired → automatic cleanup
   - AI-powered service detection (v0.2.0+)
   - Multi-service deployment
   - Error recovery scenarios

5. **[Deployment Architecture](DEPLOYMENT_ARCHITECTURE.md)** - Infrastructure and operations
   - Single-cluster vs multi-cluster deployments
   - Component specifications and resource requirements
   - Infrastructure dependencies (ArgoCD, cert-manager, external-dns)
   - Installation methods (Kustomize, Helm)
   - Network architecture and policies
   - Observability setup (Prometheus, Grafana)
   - Security and RBAC
   - Disaster recovery

---

## Architecture Decision Records (ADRs)

Architectural decisions are documented using the ADR pattern. Each ADR explains the context, options considered, decision made, and consequences.

### Decisions

1. **[ADR-001: Use ArgoCD for GitOps-Based Deployments](ADR-001-argocd-gitops.md)**
   - **Status**: Accepted
   - **Summary**: Use ArgoCD ApplicationSet pattern for deploying preview environments instead of native Kubernetes resources, Helm, or Flux
   - **Key Rationale**: ApplicationSet enables one-to-many deployments (one ApplicationSet → multiple Applications), GitOps principles, automatic sync, and excellent observability

2. **[ADR-002: Namespace-per-PR Isolation Strategy](ADR-002-namespace-isolation.md)**
   - **Status**: Accepted
   - **Summary**: Use dedicated Kubernetes namespace per PR instead of shared namespace, virtual clusters, or separate clusters
   - **Key Rationale**: Namespaces provide strong isolation (network, resources, RBAC) with negligible overhead, simple cleanup via cascade deletion, and native Kubernetes support

3. **[ADR-003: AI Integration Strategy](ADR-003-ai-integration.md)**
   - **Status**: Accepted
   - **Summary**: Phase 1 (v0.1.0) no AI, Phase 2 (v0.2.0) OpenAI integration, Phase 3 (v0.3.0) add Ollama support
   - **Key Rationale**: Prove operator works without AI first, add OpenAI for intelligent features (service detection, cost optimization), support local LLMs for on-prem deployments

---

## Quick Start Guide

### Reading Order for New Contributors

1. Start with **[System Architecture](SYSTEM_ARCHITECTURE.md)** for high-level understanding
2. Review **[ADR-001](ADR-001-argocd-gitops.md)** and **[ADR-002](ADR-002-namespace-isolation.md)** for key design decisions
3. Read **[Data Model](DATA_MODEL.md)** to understand the PreviewEnvironment CRD
4. Study **[Sequence Diagrams](SEQUENCE_DIAGRAMS.md)** for workflow details
5. Reference **[Component Design](COMPONENT_DESIGN.md)** when implementing features
6. Consult **[Deployment Architecture](DEPLOYMENT_ARCHITECTURE.md)** for operational details

### Reading Order for Operators/SREs

1. Start with **[Deployment Architecture](DEPLOYMENT_ARCHITECTURE.md)** for infrastructure requirements
2. Review **[System Architecture](SYSTEM_ARCHITECTURE.md)** for observability and security
3. Study **[ADR-002](ADR-002-namespace-isolation.md)** for isolation strategy
4. Reference **[Data Model](DATA_MODEL.md)** for resource quotas and network policies

### Reading Order for Product Managers

1. Start with **[System Architecture](SYSTEM_ARCHITECTURE.md)** overview
2. Review **[ADR-003](ADR-003-ai-integration.md)** for AI features and roadmap
3. Study **[Sequence Diagrams](SEQUENCE_DIAGRAMS.md)** for user-facing flows
4. Reference cost optimization sections in **[System Architecture](SYSTEM_ARCHITECTURE.md)** and **[ADR-002](ADR-002-namespace-isolation.md)**

---

## Architecture Principles

### 1. Simplicity Over Sophistication

**Principle**: Choose the simplest solution that meets requirements. Avoid over-engineering.

**Examples**:
- Single replica operator (v0.1.0) before adding leader election (v0.2.0)
- In-memory cache before Redis (v0.2.0 → v0.3.0)
- Static configuration before AI (v0.1.0 → v0.2.0)

### 2. GitOps by Default

**Principle**: All deployments declarative, Git as source of truth, automatic sync.

**Examples**:
- ArgoCD ApplicationSet for all preview deployments
- Application manifests stored in Git repositories
- Drift detection and reconciliation

### 3. Cost-Conscious Design

**Principle**: Optimize for minimal cost without sacrificing functionality.

**Examples**:
- Namespace-per-PR (1Mi overhead) vs virtual clusters (200Mi overhead)
- Spot instances for preview workloads
- AI response caching (80% cost reduction)
- TTL-based cleanup (prevent forgotten environments)

### 4. Security by Default

**Principle**: Secure by default, opt-in to less secure configurations.

**Examples**:
- NetworkPolicy deny-all by default
- ResourceQuota limits per namespace
- RBAC least privilege
- TLS everywhere (Ingress, webhooks)

### 5. Observability First-Class

**Principle**: Metrics, logs, traces, and events for every operation.

**Examples**:
- Prometheus metrics for all operations
- Structured logging (JSON format)
- OpenTelemetry tracing
- Kubernetes events for lifecycle changes
- Status conditions for readiness

### 6. Fail-Safe Fallbacks

**Principle**: Failures should not block user workflows. Always have fallback.

**Examples**:
- AI failure → fallback to static configuration
- GitHub API failure → retry with exponential backoff
- ArgoCD unavailable → queue operations, retry later

### 7. Idempotent Operations

**Principle**: Same input → same output. Operations can be safely retried.

**Examples**:
- Controller reconciliation loop (run multiple times safely)
- Webhook handlers (handle duplicate events)
- Resource creation (check existence before create)

### 8. Horizontal Scalability

**Principle**: Scale by adding instances, not bigger instances.

**Examples**:
- Webhook server: 3+ replicas behind load balancer
- Preview environments: Namespace-per-PR (scales to 1000s)
- Multi-cluster support for large deployments (v0.3.0)

---

## System Constraints

### Resource Limits

| Resource | Limit | Rationale |
|----------|-------|-----------|
| Max concurrent previews | 1000 per cluster | Namespace limit, can use multi-cluster for >1000 |
| Max services per preview | 20 | ResourceQuota count/pods limit |
| Max preview TTL | 7 days | Prevent forgotten environments |
| Max PR diff size for AI | 10,000 lines | OpenAI token limit (32K tokens) |
| Namespace creation time | <500ms | Kubernetes API performance |
| Preview creation time | <2 minutes | Target SLO for developer experience |

### Dependencies

| Dependency | Version | Required? | Notes |
|------------|---------|-----------|-------|
| Kubernetes | 1.28+ | Yes | Core platform |
| ArgoCD | 2.13+ | Yes | GitOps deployments |
| cert-manager | 1.16+ | Yes | TLS certificates |
| external-dns | 0.15+ | Yes | DNS management |
| Ingress NGINX | 1.11+ | Yes | HTTP(S) ingress |
| Prometheus | 2.50+ | No | Metrics (recommended) |
| Grafana | 10.0+ | No | Dashboards (recommended) |

### External APIs

| API | Purpose | Rate Limit | Fallback |
|-----|---------|------------|----------|
| GitHub API | PR metadata, diffs | 5000 req/hour | Queue requests, retry |
| OpenAI API | AI features (v0.2.0+) | 100 req/min | Static configuration |
| DNS Provider | DNS records | Varies | external-dns handles retries |

---

## Development Phases

### Phase 0: Setup & Learning (Weeks 1-2) - CURRENT

**Status**: In Progress

**Goals**:
- Set up development environment
- Learn Go idioms and Kubernetes client-go
- Scaffold operator with Kubebuilder

**Deliverables**:
- ✅ Kubebuilder project structure
- ✅ PreviewEnvironment CRD types
- ✅ Basic controller skeleton
- ✅ CI/CD pipeline
- ⏳ Architecture documentation (this document)

---

### Phase 1: Core Operator - No AI (Weeks 3-6)

**Status**: Not Started

**Goals**:
- Build functional operator without AI
- GitHub webhook integration
- ArgoCD ApplicationSet integration
- Namespace isolation with ResourceQuota
- TTL-based cleanup

**Deliverables**:
- PreviewEnvironment reconciliation loop
- GitHub webhook handler
- ApplicationSet creation/deletion
- Ingress + external-dns integration
- Basic cost estimation
- Unit + integration tests (>80% coverage)

**Milestone**: v0.1.0 release

---

### Phase 2: AI Integration (Weeks 7-10)

**Status**: Not Started

**Goals**:
- Add AI-powered features
- OpenAI API integration
- Service detection from code diffs
- Cost prediction and optimization

**Deliverables**:
- AI engine with OpenAI client
- Service detection implementation
- Response caching
- Feature flags
- A/B testing framework
- Cost tracking for AI API calls

**Milestone**: v0.2.0 release

---

### Phase 3: Production Polish (Weeks 11-12)

**Status**: Not Started

**Goals**:
- Production-ready operator
- Security hardening
- Performance optimization
- Documentation and examples

**Deliverables**:
- Security audit
- Performance testing (1000s of PRs)
- Ollama support (local LLM)
- Helm chart
- Documentation site
- Example applications
- Demo video

**Milestone**: v0.2.0 public launch

---

## Contributing to Architecture

### Proposing Changes

1. **Create ADR**: Use ADR template (see existing ADRs for format)
2. **Document alternatives**: List at least 3 options with pros/cons
3. **Justify decision**: Explain why chosen option is best
4. **Update diagrams**: Keep Mermaid diagrams up-to-date
5. **Get review**: Architecture changes require review from maintainer

### ADR Template

```markdown
# ADR-XXX: Title

## Status

**Proposed** | **Accepted** | **Deprecated** | **Superseded by ADR-YYY**

## Context

What problem are we solving? What are the requirements?

## Options Considered

### Option 1: ...

**Pros**:
- ✅ Advantage 1
- ✅ Advantage 2

**Cons**:
- ❌ Disadvantage 1
- ❌ Disadvantage 2

### Option 2: ...

## Decision

What did we decide? Why?

## Consequences

What are the positive/negative/neutral consequences?

## References

Links to documentation, blog posts, etc.
```

---

## Glossary

| Term | Definition |
|------|------------|
| **Preview Environment** | Ephemeral, isolated environment for testing pull request changes |
| **ApplicationSet** | ArgoCD resource that generates multiple Application CRs from template |
| **Reconciliation** | Kubernetes controller pattern: observe desired state → make current state match |
| **Owner Reference** | Kubernetes mechanism for cascade deletion (delete parent → delete children) |
| **Finalizer** | Kubernetes mechanism to run cleanup logic before resource deletion |
| **TTL** | Time-to-Live, duration before preview environment is automatically deleted |
| **GitOps** | Declarative deployments where Git is source of truth, automatic sync |
| **ResourceQuota** | Kubernetes resource that limits CPU/memory/object count per namespace |
| **NetworkPolicy** | Kubernetes resource that controls ingress/egress traffic rules |

---

## Useful Commands

### Query Resources

```bash
# List all preview environments
kubectl get previewenvironments -A

# Get preview environment details
kubectl get previewenvironment pr-123 -o yaml

# List all preview namespaces
kubectl get namespaces -l preview.previewd.io/managed-by=previewd

# Get ApplicationSet for preview
kubectl get applicationset preview-123 -n argocd -o yaml

# Get Applications for preview
kubectl get applications -n argocd -l preview.previewd.io/pr=123

# Get all resources in preview namespace
kubectl get all -n pr-123
```

### Debug

```bash
# Check operator logs
kubectl logs -n previewd-system deployment/previewd-controller-manager -f

# Check webhook logs
kubectl logs -n previewd-system deployment/previewd-webhook -f

# Check ArgoCD sync status
kubectl get applications -n argocd -l preview.previewd.io/pr=123 \
  -o jsonpath='{.items[*].status.sync.status}'

# Check events
kubectl get events -n pr-123 --sort-by='.lastTimestamp'
```

### Metrics

```bash
# Query Prometheus for active previews
curl -s 'http://prometheus:9090/api/v1/query?query=previewd_environments_total{status="active"}'

# Query total cost
curl -s 'http://prometheus:9090/api/v1/query?query=sum(previewd_environment_cost_estimate_usd)'
```

---

## Architecture Reviews

### Scheduled Reviews

- **Monthly**: Review open ADRs, update status
- **Quarterly**: Review architecture for technical debt, refactoring opportunities
- **Before major releases**: Comprehensive architecture review

### Review Checklist

- [ ] ADRs up-to-date?
- [ ] Diagrams accurate?
- [ ] New features documented?
- [ ] Performance/scalability concerns addressed?
- [ ] Security vulnerabilities identified?
- [ ] Cost implications considered?

---

## Contact

**Architecture Owner**: Mike Lane (@mikelane)
- GitHub: [@mikelane](https://github.com/mikelane)
- Email: mikelane@gmail.com

**Questions?** Open a GitHub Discussion in the [Previewd repository](https://github.com/mikelane/previewd/discussions)

---

**Document Status**: ✅ Complete
**Last Updated**: 2025-11-09
**Next Review**: 2025-12-09
