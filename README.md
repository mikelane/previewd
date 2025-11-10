# Previewd

**AI-Powered Preview Environments for Kubernetes**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28+-326CE5?logo=kubernetes)](https://kubernetes.io/)

---

## Vision

Previewd is a Kubernetes operator that automatically creates, manages, and destroys preview environments for pull requests. Using AI, it intelligently determines which services to deploy, generates realistic test data, optimizes costs, and runs only the tests that matter.

### The Problem We Solve

**Manual preview environments are painful:**
- ⏰ Take 30+ minutes to set up manually
- 💰 Waste money running unnecessary services 24/7
- 🎲 Use unrealistic test data that doesn't catch bugs
- 🤷 Require deep knowledge of service dependencies
- 🔥 Shared staging environments have conflicts and stale data

**Previewd makes preview environments:**
- ⚡ **Fast** - Ready in 2 minutes, not 30
- 🧠 **Smart** - AI determines what services you need
- 💰 **Cheap** - 70% cost reduction through intelligent resource sizing
- 🎯 **Realistic** - AI-generated test data that looks like production
- 🧪 **Tested** - Automatically runs relevant integration tests
- 🗑️ **Clean** - Auto-destroys when PR is merged/closed

## Quick Demo

```bash
# Developer opens a PR
git push origin feature/new-auth

# Previewd automatically:
# 1. Analyzes code changes (AI detects: auth-service + user-service needed)
# 2. Spins up minimal preview environment (2 services, not all 20)
# 3. Generates 100 realistic test users with AI
# 4. Runs integration tests (AI selects 42 relevant tests, not all 500)
# 5. Posts preview URL to PR: https://pr-1234.preview.myapp.com
# 6. Reports: "✅ Ready in 2m 15s • Cost: $2.34/day • 41/42 tests passed"

# Developer reviews, merges PR
# Previewd automatically destroys environment → $0/day
```

## Current Status

🚧 **Project Phase:** Planning & Design
📅 **Started:** November 2025
🎯 **Target v0.1.0:** February 2026 (Basic operator, no AI)
🎯 **Target v0.2.0:** April 2026 (AI-powered features)

**Current progress:**
- [x] Project vision defined
- [x] Architecture designed
- [x] Documentation created
- [x] Go environment setup
- [x] Kubebuilder operator scaffold
- [x] Custom Resource Definitions (v1alpha1)
- [ ] Basic reconciliation loop
- [ ] GitHub webhook integration
- [ ] ArgoCD integration
- [ ] AI code analysis
- [ ] Synthetic data generation
- [ ] Cost optimization

See [ROADMAP.md](ROADMAP.md) for detailed timeline.

## Architecture Overview

```mermaid
graph TB
    subgraph "Previewd Operator"
        GH[GitHub Webhook Handler]
        AI[AI Engine<br/>Code Analysis]
        DG[Data Generator<br/>Synthetic Data]

        GH --> RL[Reconciliation Loop]
        AI --> RL
        DG --> RL

        RL --> EC[Environment Controller]
        RL --> ARGO[ArgoCD Integration]
        RL --> CO[Cost Optimizer]
    end

    EC --> K8S[Kubernetes Cluster]
    ARGO --> K8S
    CO --> K8S

    style GH fill:#4CAF50
    style AI fill:#2196F3
    style DG fill:#FF9800
    style RL fill:#9C27B0
    style K8S fill:#326CE5
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed design.

## Key Features

### Phase 1: Core Operator (v0.1.0)
- ✅ **PR Integration** - Webhook from GitHub creates preview environments
- ✅ **Namespace Isolation** - Each PR gets isolated namespace
- ✅ **ArgoCD Deployment** - GitOps-based service deployment
- ✅ **DNS Routing** - Automatic URLs like `pr-123.preview.example.com`
- ✅ **Auto Cleanup** - TTL-based environment destruction
- ✅ **Cost Tracking** - Estimate daily costs per environment

### Phase 2: AI Features (v0.2.0)
- 🤖 **Smart Dependencies** - AI analyzes code to determine which services needed
- 🤖 **Synthetic Data** - AI generates realistic test data from production schemas
- 🤖 **Cost Optimization** - AI predicts lifespan and sizes resources optimally
- 🤖 **Intelligent Tests** - AI selects which tests to run based on changes

### Phase 3: Advanced Features (v0.3.0+)
- 🔮 **Visual Diff** - Screenshot comparison for UI changes
- 🔮 **Performance Testing** - Automated load tests on preview environments
- 🔮 **Database Migrations** - Test migrations safely in preview
- 🔮 **Slack/Teams Integration** - Notifications and slash commands

## Technology Stack

### Languages & Frameworks
- **Go 1.21+** - Operator implementation
- **Kubebuilder** - Operator SDK framework
- **Python** - AI integration, data generation scripts

### Kubernetes Ecosystem
- **Kubernetes 1.28+** - Orchestration platform
- **ArgoCD** - GitOps deployment
- **Cert-Manager** - TLS certificate management
- **External-DNS** - Automatic DNS record creation

### AI/ML
- **OpenAI API** - Code analysis, data generation (v0.2.0)
- **Ollama** - Local LLM option for on-prem deployments
- **LangChain** - AI orchestration (optional)

### Infrastructure
- **AWS** - Primary cloud platform (EKS)
- **AWS CDK** - Infrastructure as code (for examples)
- **PostgreSQL/MySQL** - Schema metadata storage

## Installation

> ⚠️ **Not yet available** - Project is in early development

Once v0.1.0 is released, installation will be:

```bash
# Install Previewd operator
kubectl apply -f https://github.com/mikelane/previewd/releases/latest/install.yaml

# Configure GitHub webhook
kubectl apply -f config/samples/github-webhook.yaml

# Create your first preview environment
kubectl apply -f config/samples/preview-environment.yaml
```

## Development Setup

See [DEVELOPMENT.md](docs/DEVELOPMENT.md) for detailed setup instructions.

**Quick start:**

```bash
# Prerequisites: Go 1.21+, Docker, kubectl, kind (for local cluster)

# Clone repository
git clone https://github.com/mikelane/previewd.git
cd previewd

# Install dependencies
make install

# Run operator locally
make run

# Run tests
make test

# Build and push image
make docker-build docker-push IMG=your-registry/previewd:latest
```

## Custom Resource Example

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

### Complete Example

```yaml
apiVersion: preview.previewd.io/v1alpha1
kind: PreviewEnvironment
metadata:
  name: pr-1234-feature-auth
spec:
  # Required fields
  repository: myorg/myapp
  prNumber: 1234
  headSHA: abcdef1234567890abcdef1234567890abcdef12

  # Optional fields
  baseBranch: main
  headBranch: feature/new-auth
  ttl: "8h"
  services:
    - api
    - web
    - worker

status:
  phase: Ready
  url: https://pr-1234.preview.example.com
  namespace: preview-pr-1234
  services:
    - name: api
      ready: true
      url: https://api-pr-1234.preview.example.com
    - name: web
      ready: true
      url: https://pr-1234.preview.example.com
  costEstimate:
    currency: USD
    hourlyCost: "0.15"
    totalCost: "1.20"
  conditions:
    - type: Ready
      status: "True"
      reason: ServicesReady
      message: All services are healthy
  createdAt: "2025-11-09T08:00:00Z"
  expiresAt: "2025-11-09T16:00:00Z"
```

For detailed API documentation, see [docs/api-reference.md](docs/api-reference.md).

## Why Previewd?

### For Developers
- ⚡ **Speed** - Get preview environments in 2 minutes, not 30
- 🎯 **Confidence** - Test with realistic data that catches bugs
- 🧹 **Clean** - No more fighting over shared staging environments
- 🤝 **Collaboration** - Share preview URLs with designers, PMs, QA

### For Platform Teams
- 💰 **Cost Savings** - 70% reduction through smart resource sizing
- 🤖 **Automation** - No manual setup, no tickets for preview envs
- 📊 **Visibility** - Track costs, usage, and environment health
- 🔒 **Security** - Isolated namespaces, automatic cleanup

### For Engineering Leaders
- 📈 **Faster Delivery** - Reduce cycle time from PR to production
- 💵 **Lower Cloud Bills** - Optimize preview environment costs
- ✅ **Higher Quality** - Catch integration bugs before merge
- 😊 **Developer Happiness** - Remove friction from deployment

## Project Goals

### Technical Goals
1. **Learn Go** - Achieve production-level Go proficiency
2. **Master K8s Operators** - Deep understanding of operator pattern
3. **AI Integration** - Practical LLM integration in infrastructure
4. **Cloud-Native Expertise** - ArgoCD, GitOps, service mesh

### Career Goals
1. **Resume Impact** - "Built AI-powered K8s operator used by X companies"
2. **Thought Leadership** - Speak at KubeCon, write blog posts
3. **Community Building** - 500+ GitHub stars, active contributors
4. **Consulting Opportunities** - Platform engineering expertise

### Success Metrics
- ⭐ **500 GitHub stars** in first 6 months
- 📦 **10 companies** using in production
- 🗣️ **Invited to speak** at K8s meetups or conferences
- 💼 **Recruiter outreach** for Staff+ platform engineering roles

## Roadmap

- **Q4 2025** - Planning, architecture, Go learning
- **Q1 2026** - v0.1.0 (Basic operator, no AI)
- **Q2 2026** - v0.2.0 (AI-powered features)
- **Q3 2026** - v0.3.0 (Advanced features, production adoption)
- **Q4 2026** - v1.0.0 (Stable API, enterprise-ready)

See [ROADMAP.md](ROADMAP.md) for detailed timeline and milestones.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for complete guidelines.

**Quick start for contributors:**

### Code Owners (@mikelane)
Code owners can develop directly from clones:
```bash
git clone git@github.com:mikelane/previewd.git
cd previewd
git checkout -b feat/your-feature
# Make changes, then push and create PR
```

### External Contributors
External contributors must use fork-based workflow:
```bash
# Fork on GitHub first, then:
git clone git@github.com:YOUR-USERNAME/previewd.git
cd previewd
git remote add upstream git@github.com:mikelane/previewd.git
git checkout -b feat/your-feature
# Make changes, push to your fork, then create PR from fork
```

**Development standards:**
- ✅ **Test-Driven Development** - Write tests first, watch fail, make pass
- ✅ **100% Coverage Required** - All new code must be fully tested
- ✅ **Issue-Based Development** - Every change requires a GitHub issue
- ✅ **Documentation Updates Mandatory** - Update docs in same PR as code
- ✅ **Conventional Commits** - Structured commit messages

See [CONTRIBUTING.md](CONTRIBUTING.md) for complete workflow, testing requirements, and quality standards.

## License

MIT License - see [LICENSE](LICENSE) for details.

Copyright (c) 2025 Mike Lane

## Acknowledgments

- Inspired by [Okteto](https://okteto.com/), [Argo Workflows](https://argoproj.github.io/), and [Telepresence](https://www.telepresence.io/)
- Built with [Kubebuilder](https://book.kubebuilder.io/)
- AI integration patterns from [LangChain](https://www.langchain.com/)

---

**Status:** 🚧 Early Development
**Maintainer:** Mike Lane ([@mikelane](https://github.com/mikelane))
**Contact:** mikelane@gmail.com

*Making preview environments fast, smart, and cheap.*
