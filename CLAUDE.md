# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working on the Previewd project.

## Project Overview

**Previewd** is an AI-powered Kubernetes operator that automatically creates, manages, and destroys preview environments for pull requests. This is a strategic career project for Mike Lane to:

1. **Learn Go** - Achieve production-level proficiency in Go
2. **Master K8s Operators** - Deep expertise in Kubernetes operator pattern
3. **AI Integration** - Practical experience integrating LLMs with infrastructure
4. **Resume Impact** - Showcase multi-language expertise (Python + Rust + Go) and platform engineering

## Strategic Context

### Why This Project Matters

**Career positioning:**
- Mike is a Principal Software Engineer & Dev Lead at GDIT
- Current expertise: Python (expert), Rust (PyO3 via dioxide), TypeScript, Java
- Gap: No production Go experience, limited K8s operator experience
- Goal: Position as **polyglot platform engineer** for Staff+ roles

**Resume multiplier effect:**
- ✅ Python expert (dioxide, valid8r, pytest plugins)
- ✅ Rust proficiency (dioxide with PyO3)
- ✅ TypeScript (second language)
- 🎯 **Add Go + K8s operators** ← This project
- 🎯 **Add AI/ML integration in infrastructure** ← This project

**Target outcomes:**
- 500+ GitHub stars in 6 months
- 10 companies using in production
- Speaking opportunities at K8s meetups/conferences
- Recruiter outreach for Staff+ platform engineering roles

### Project Philosophy

**Engineering principles:**
- **Test-Driven Development** - Write tests first, watch fail, make pass
- **Clean Code** - Readable, maintainable, well-documented
- **AI for value, not hype** - Only use AI where it provides measurable benefit
- **Incremental delivery** - Ship v0.1.0 without AI, add AI in v0.2.0
- **Developer experience first** - Optimize for developer happiness

**Documentation as code:**
- Treat docs with same rigor as production code
- Keep docs updated with every feature
- Examples must work (test them in CI)

## Repository Structure

```
previewd/
├── README.md                   # Project overview, quick start
├── CLAUDE.md                   # This file - context for Claude
├── ARCHITECTURE.md             # Technical architecture, design decisions
├── ROADMAP.md                  # Development timeline, milestones
├── CONTRIBUTING.md             # Contribution guidelines
├── LICENSE                     # MIT License
├── Makefile                    # Build, test, deploy commands
├── go.mod                      # Go module definition
├── go.sum                      # Go dependency checksums
├── main.go                     # Operator entrypoint
├── api/                        # Custom Resource Definitions
│   └── v1alpha1/
│       ├── previewenvironment_types.go
│       └── groupversion_info.go
├── controllers/                # Reconciliation logic
│   └── previewenvironment_controller.go
├── internal/                   # Private application code
│   ├── github/                 # GitHub webhook integration
│   ├── ai/                     # AI engine (code analysis, data gen)
│   ├── cost/                   # Cost optimization
│   └── argocd/                 # ArgoCD integration
├── config/                     # Kubernetes manifests
│   ├── crd/                    # Custom Resource Definitions
│   ├── rbac/                   # RBAC permissions
│   ├── manager/                # Operator deployment
│   └── samples/                # Example CRs
├── pkg/                        # Public libraries (if any)
├── docs/                       # Documentation
│   ├── DEVELOPMENT.md          # Dev setup, contributing
│   ├── ARCHITECTURE.md         # Architecture deep dive
│   └── examples/               # Usage examples
└── hack/                       # Scripts for development
    └── local-setup.sh          # Local K8s cluster setup
```

## Development Phases

### Phase 0: Setup & Learning (Weeks 1-2)

**Goals:**
- Set up Go development environment
- Learn Go basics and idioms
- Understand K8s client-go library
- Scaffold operator with Kubebuilder

**Key tasks:**
- [ ] Complete "A Tour of Go" (tour.golang.org)
- [ ] Build simple CLI tool in Go
- [ ] Learn Go testing with `go test`
- [ ] Install Kubebuilder and create scaffold
- [ ] Create PreviewEnvironment CRD
- [ ] Run operator locally against kind cluster

**Success criteria:**
- Can write idiomatic Go code
- Understand Go modules, interfaces, error handling
- Operator deploys to local cluster
- CRD can be created/deleted

### Phase 1: Core Operator - No AI (Weeks 3-6)

**Goals:**
- Build functional operator (MVP)
- GitHub webhook integration
- ArgoCD deployment integration
- Basic environment lifecycle

**Key tasks:**
- [ ] Implement reconciliation loop
- [ ] GitHub webhook handler (PR opened/closed events)
- [ ] Create namespace-per-PR
- [ ] Deploy services via ArgoCD ApplicationSet
- [ ] Ingress/DNS routing (pr-123.preview.example.com)
- [ ] TTL-based cleanup
- [ ] Cost estimation (sum of pod resources)
- [ ] Unit tests for controllers
- [ ] Integration tests with kind

**Success criteria:**
- Open PR → preview environment created
- Services accessible via URL
- Close PR → environment destroyed
- Tests pass, >80% coverage

**Milestone:** v0.1.0 release - Functional operator (no AI)

### Phase 2: AI Integration (Weeks 7-10)

**Goals:**
- Add AI-powered features
- Prove AI provides value over static config
- Measure impact (cost, time, accuracy)

**Key tasks:**
- [ ] Integrate OpenAI API (with caching)
- [ ] Code diff analysis → service dependencies
- [ ] Synthetic test data generation from schemas
- [ ] Cost prediction model (lifespan estimation)
- [ ] Smart test selection
- [ ] A/B testing: AI vs non-AI
- [ ] Cost tracking for AI API calls

**Success criteria:**
- AI correctly detects service dependencies (>90% accuracy)
- Synthetic data looks realistic (manual review)
- Cost optimization saves >50% vs always-on
- AI features are toggle-able (feature flags)

**Milestone:** v0.2.0 release - AI-powered operator

### Phase 3: Polish & Launch (Weeks 11-12)

**Goals:**
- Production-ready operator
- Documentation and examples
- Community launch

**Key tasks:**
- [ ] Security review (RBAC, secrets management)
- [ ] Performance testing (1000s of PRs)
- [ ] Helm chart for installation
- [ ] Documentation site (GitHub Pages)
- [ ] Example applications (Node.js, Python, Java)
- [ ] Demo video (2 min)
- [ ] Blog post on dev.to/Medium
- [ ] Launch on HN, Reddit, X

**Success criteria:**
- Can deploy to production cluster safely
- Documentation is comprehensive
- 5 people can install without asking questions
- Blog post gets >1000 views

**Milestone:** v0.2.0 public launch

## Technical Decisions

### Language & Framework Choices

**Go for operator:**
- ✅ Dominant language for K8s ecosystem
- ✅ Native K8s client (client-go)
- ✅ Excellent concurrency model
- ✅ Fast compile times, single binary
- ❌ Less familiar than Python (learning curve)

**Kubebuilder for operator SDK:**
- ✅ Official K8s operator SDK
- ✅ Best practices baked in
- ✅ Active community, good docs
- ❌ Opinionated structure (but that's good)

**ArgoCD for deployments:**
- ✅ GitOps standard
- ✅ Declarative, auditable
- ✅ Multi-tenancy support
- ❌ Requires cluster setup (but worth it)

**OpenAI for AI (initially):**
- ✅ Best-in-class for code analysis
- ✅ Fast to integrate
- ✅ Can swap for local LLM later (Ollama)
- ❌ Costs money (but can cache)

### AI Integration Strategy

**Phase 1 (v0.1.0): No AI**
- Prove core operator works
- Static configuration (user specifies services)
- Establishes baseline for comparison

**Phase 2 (v0.2.0): Add AI**
- AI augments, doesn't replace static config
- Feature flags to toggle AI on/off
- Measure: does AI save time/money/improve accuracy?

**Phase 3 (v0.3.0): AI-first**
- AI is default, static config is fallback
- Local LLM support (Ollama) for on-prem
- Fine-tune models on user's codebase

**Principle:** AI must provide measurable value, not just marketing

### Testing Strategy

**Test levels:**
1. **Unit tests** - Individual functions, mocked K8s API
2. **Integration tests** - Real K8s cluster (kind), real CRDs
3. **E2E tests** - Full workflow (GitHub PR → environment → cleanup)
4. **Load tests** - 1000s of PRs simultaneously

**Coverage target:** >80% for v0.1.0, >90% for v1.0.0

**Test data:**
- Use pytest-style parametrization (Go has similar with table-driven tests)
- No loops/branching in tests (keep tests simple)
- One assertion per test when possible

### Cost Optimization Principles

**Resource sizing:**
- Default: minimal (0.1 CPU, 256Mi RAM per service)
- AI predicts: based on PR size, user history, time of day
- User can override via annotations

**Cleanup strategy:**
- Default TTL: 4 hours
- AI extends TTL if PR still active
- Force destroy after 7 days (even if active)

**Spot instances:**
- Use spot for short-lived environments (<4h predicted)
- Use on-demand for longer environments

## Commands & Workflows

### Local Development

```bash
# Setup local K8s cluster (kind)
make kind-create

# Install CRDs to cluster
make install

# Run operator locally (watches cluster)
make run

# Run tests
make test

# Run tests with coverage
make test-coverage

# Build operator binary
make build

# Build and load Docker image to kind
make docker-build-kind

# Deploy operator to cluster
make deploy

# Undeploy operator
make undeploy

# Delete kind cluster
make kind-delete
```

### Testing Workflows

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Run specific test
go test -v -run TestPreviewEnvironmentController ./controllers/...

# Update golden files (if using golden file testing)
go test ./... -update

# Benchmark tests
go test -bench=. ./...
```

### Git Workflow

**Branch naming:**
- `feat/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation only
- `test/description` - Test improvements
- `refactor/description` - Code refactoring

**Commit messages:**
```
<type>(<scope>): <subject>

<body>

<footer>
```

Types: feat, fix, docs, style, refactor, test, chore

Example:
```
feat(ai): add code diff analysis for service detection

Integrate OpenAI API to analyze PR diffs and automatically
detect which services are affected by code changes.

- Add github package for PR diff fetching
- Add ai package with OpenAI client
- Cache AI responses to reduce API costs
- Add feature flag to enable/disable AI analysis

Closes #42
```

**No attribution lines:**
- Do NOT add "Co-Authored-By: Claude" or similar
- Commits should be attributed to Mike Lane only

## Development Workflow Requirements

**CRITICAL: All agents (including Claude Code) MUST follow these workflow rules without exception.**

### Issue-Based Development

**MUST have an issue for all work:**
- Every task, feature, bug fix, or improvement MUST have a corresponding GitHub issue
- If no issue exists for the work you're about to do, **STOP** and ask the user if you should create one
- Issues provide traceability, context, and documentation for all changes
- Issues enable proper project management and milestone tracking

**Issue lifecycle:**
- Issues MUST be updated as work progresses (status changes, blockers, progress notes)
- Reference the issue number in all related commits and PRs using `#issue-number`
- Close issues via PR merge using "Closes #issue-number" in PR description

### Pull Request Workflow

**MUST use PRs for ALL commits:**
- Never commit directly to `main` or `master` branch
- ALL work MUST go through a Pull Request, no exceptions
- PRs provide documentation, review history, and archaeology (understanding why changes were made)
- Even single-commit changes require a PR

**PR requirements:**
- Create feature branch from `main` using branch naming conventions above
- Open PR early (can be draft) to signal work in progress
- PR title MUST follow conventional commit format: `<type>(<scope>): <subject>`
- PR description MUST include:
  - Summary of changes
  - Why this change is needed
  - Testing performed
  - Reference to issue(s) being addressed: `Closes #42`
  - Any breaking changes or migration notes

**PR updates:**
- Keep PR description current as implementation evolves
- Update PR with test results, benchmark data, or performance metrics
- Add comments to explain complex changes or design decisions
- Respond to review feedback promptly

### Continuous Documentation

**MUST keep documentation up-to-date as you work:**
- Documentation changes are NOT optional - they are part of the definition of done
- Update docs in the SAME PR that changes functionality
- Never defer documentation updates to "later" or separate issues

**Internal documentation (in `/docs`):**
- Update architecture docs when adding/changing components
- Update development docs when changing dev workflow or tooling
- Update decision records (ADRs) when making architectural choices

**Customer-facing documentation:**
- Update README.md when adding features or changing usage
- Update API docs when changing interfaces
- Update examples when changing behavior
- Update troubleshooting guides when fixing common issues

**Code-level documentation:**
- Update godoc comments when changing function signatures or behavior
- Update inline comments when refactoring complex logic
- Update examples in code comments when changing APIs

### Product Lifecycle Management

**MUST maintain strict product lifecycle:**
- Issues track all planned work (backlog, in progress, done)
- PRs track all code changes (open, in review, merged)
- Milestones track releases and major deliverables
- Labels categorize work (bug, feature, documentation, etc.)

**Status tracking:**
- Update issue status when starting work (`in progress`)
- Update PR status when ready for review (remove `draft` status)
- Link related issues and PRs bidirectionally
- Close completed work promptly (don't leave stale PRs/issues open)

**Communication:**
- Use issue/PR comments for questions, blockers, or status updates
- Tag relevant people when input is needed
- Document decisions made during implementation
- Provide context for future maintainers

### Enforcement

These rules are **mandatory** for all development work on Previewd. They ensure:
- ✅ Complete audit trail of all changes
- ✅ Documentation stays synchronized with code
- ✅ Project management has accurate visibility
- ✅ Future maintainers can understand historical context
- ✅ Quality gates are enforced (review, testing, documentation)

**If you're unsure whether an issue exists or what to do, ASK. Never proceed without proper issue tracking.**

## AI Integration Guidelines

### When to Use AI

**Good use cases:**
1. **Code analysis** - Detect service dependencies from diffs
2. **Data generation** - Create realistic test data
3. **Cost prediction** - Estimate environment lifespan
4. **Test selection** - Pick relevant tests to run

**Bad use cases:**
1. **Critical path logic** - Don't rely on AI for control flow
2. **Security decisions** - AI shouldn't determine access control
3. **Billing** - Don't let AI make actual billing decisions

### AI Response Handling

**Always:**
- Validate AI responses (schema, constraints)
- Have fallback for AI failures
- Cache responses (identical diffs → same services)
- Rate limit AI API calls
- Track costs per API call

**Never:**
- Trust AI blindly
- Block operations waiting for AI (async preferred)
- Expose raw AI responses to users (sanitize first)

### Prompt Engineering

**Template:**
```
You are analyzing a code diff to determine service dependencies.

Context:
- Repository: {repo}
- PR: {pr_number}
- Files changed: {file_count}

Diff:
{diff}

Task: List all services that are directly modified or depend on changed code.

Return JSON array of service names only.
Example: ["auth-service", "user-service"]
```

**Principles:**
- Be specific about output format (JSON, schema)
- Provide context (repo, PR metadata)
- Give examples of expected output
- Keep prompts under 4000 tokens (cost optimization)

## Go Best Practices (2025)

**Version:** Go 1.25.4 (November 2025) - Current stable release

**Go 1.25 Major Features:**
- **Container-aware GOMAXPROCS** - Runtime auto-adjusts to cgroup CPU limits (critical for K8s!)
- **testing/synctest** - Stable package for deterministic concurrent testing
- **Experimental JSON v2** - `encoding/json/v2` with better performance (`GOEXPERIMENT=jsonv2`)
- **DWARF v5** - Smaller binaries, faster linking
- **FlightRecorder** - Continuous runtime tracing for debugging intermittent issues
- **Core types removed** - Generics simplified in language spec

**Essential Patterns:**
- **TDD with table-driven tests** - Use `t.Run()` for subtests, `b.Loop()` for benchmarks
- **Error handling** - Always check errors immediately, wrap with `%w` for context
- **Dependency management** - Pin versions, commit `go.sum`, run `go mod tidy` regularly
- **Use `internal/` packages** - Protect implementation details (enforced by toolchain)
- **Container-aware** - Leverage GOMAXPROCS auto-adjustment for Kubernetes deployments
- **Idiomatic naming** - Interfaces: `Reader`, Methods: `Read()`, Packages: `github` (lowercase, no underscores)

**Project Structure:**
```
previewd/
├── cmd/previewd/main.go      # Entrypoint
├── api/v1alpha1/             # CRD types (Kubebuilder)
├── controllers/              # Reconcilers
├── internal/                 # Private packages
│   ├── github/
│   ├── ai/
│   └── argocd/
└── config/                   # K8s manifests
```

**Code Quality:**
- Run `golangci-lint` v2 with `standard` preset
- Coverage: >80% for v0.1.0, >90% for v1.0.0
- Use `gofmt` and `goimports` (enforced in CI)

**Operator-Specific (Kubebuilder):**
- **Idempotent reconciliation** - Same input → same output (always)
- **One controller per API** - Single reconciler per CRD
- **Owner references** - Enable automatic garbage collection
- **Status updates** - Update status separately from spec

**See GO_BEST_PRACTICES.md for comprehensive guidelines**

## Common Gotchas

### Go-specific

1. **Goroutine leaks** - Always ensure goroutines can terminate
2. **Channel deadlocks** - Close channels when done writing
3. **Nil maps** - Initialize maps before use (`make(map[string]string)`)
4. **Pointer to loop variable** - Create new variable in loop
5. **Defer in loops** - Be careful with resource cleanup

### Kubernetes-specific

1. **API rate limiting** - Use informers, not polling
2. **Namespace deletion** - Finalizers prevent stuck namespaces
3. **RBAC** - Operator needs permissions for all resources it manages
4. **CRD versioning** - Use conversion webhooks for v1alpha1 → v1beta1

### Operator-specific

1. **Reconciliation idempotency** - Same input → same output (always)
2. **Ownership** - Set owner references for garbage collection
3. **Status updates** - Update status in separate call from spec
4. **Requeue vs error** - Requeue for transient failures, error for permanent

## Learning Resources

### Go
- [A Tour of Go](https://tour.golang.org/) - Interactive tutorial
- [Effective Go](https://go.dev/doc/effective_go) - Best practices
- [Go by Example](https://gobyexample.com/) - Practical examples

### Kubernetes
- [Kubernetes API Concepts](https://kubernetes.io/docs/reference/using-api/api-concepts/)
- [client-go Examples](https://github.com/kubernetes/client-go/tree/master/examples)
- [Programming Kubernetes](https://www.oreilly.com/library/view/programming-kubernetes/9781492047094/) - Book

### Operators
- [Kubebuilder Book](https://book.kubebuilder.io/) - Official guide
- [Operator SDK](https://sdk.operatorframework.io/)
- [Best Practices](https://sdk.operatorframework.io/docs/best-practices/)

### AI Integration
- [OpenAI API Docs](https://platform.openai.com/docs)
- [Ollama](https://ollama.ai/) - Local LLM
- [LangChain Go](https://github.com/tmc/langchaingo)

## Success Metrics

### Technical Metrics
- ⭐ **GitHub stars:** 500 in 6 months
- 🔧 **Contributors:** 5+ external contributors
- 📦 **Production users:** 10 companies
- 🧪 **Test coverage:** >90%
- 🚀 **Performance:** <2 min to ready state

### Career Metrics
- 💼 **Recruiter outreach:** Staff+ platform engineering roles
- 🎤 **Speaking:** Invited to K8s meetup or conference
- 📝 **Thought leadership:** Blog posts >1000 views
- 💰 **Consulting:** Opportunities at $200-400/hr

### Learning Metrics
- ✅ **Go proficiency:** Can write production Go code
- ✅ **K8s expertise:** Deep understanding of operators
- ✅ **AI integration:** Practical LLM integration patterns
- ✅ **Platform engineering:** FinOps, observability, GitOps

## Project Mantras

1. **Ship early, iterate fast** - v0.1.0 in 6 weeks, not 6 months
2. **AI for value, not hype** - Measure before adding AI features
3. **Developer experience first** - Optimize for happiness
4. **Tests are non-negotiable** - TDD always
5. **Documentation is code** - Treat with same rigor
6. **Open source mindset** - Build in public, learn in public

## Contact & Collaboration

**Maintainer:** Mike Lane
- GitHub: [@mikelane](https://github.com/mikelane)
- Email: mikelane@gmail.com
- LinkedIn: [linkedin.com/in/lanemik](https://www.linkedin.com/in/lanemik)

**Current status:** Early development (as of November 2025)
**Looking for:** Early adopters, contributors, feedback

---

This document should give you complete context to work on Previewd effectively. Refer to ARCHITECTURE.md for technical details and ROADMAP.md for timeline.
