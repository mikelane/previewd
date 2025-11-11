# Contributing to Previewd

Thank you for your interest in contributing to Previewd! This guide will help you understand our development workflow and how to contribute effectively.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Issue Lifecycle](#issue-lifecycle)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Requirements](#testing-requirements)
- [Documentation Requirements](#documentation-requirements)
- [Community](#community)

## Code of Conduct

This project adheres to a simple code of conduct:

- Be respectful and inclusive
- Focus on constructive feedback
- Assume good intentions
- Help others learn and grow

## Getting Started

### Prerequisites

- Go 1.25+ (latest stable)
- Docker and Docker Compose
- kubectl
- kind (Kubernetes in Docker)
- Kubebuilder 4.x+
- make

### Local Development Setup

```bash
# Clone the repository
git clone https://github.com/mikelane/previewd.git
cd previewd

# Create a local Kubernetes cluster
make kind-create

# Install CRDs to the cluster
make install

# Run the operator locally (watches the cluster)
make run
```

For detailed development setup and workflow guides, see the `docs/` directory.

## Development Workflow

Previewd uses a **GitHub-native, issue-based workflow** with strict requirements:

### Core Principles

1. **Every change requires an issue** - No exceptions
2. **All commits go through Pull Requests** - Never commit directly to `main`
3. **Documentation updates are mandatory** - Part of Definition of Done
4. **Tests are written first** - Test-Driven Development (TDD)
5. **Code review is required** - Quality gate before merge

### Workflow Overview

```
Issue Created → BDD Tests → Implementation → Code Review → QA Review → SRE Review → Merge → Deploy
     ↓              ↓             ↓              ↓            ↓            ↓         ↓        ↓
  Triage       BDD Ready    In Progress    Code Review   QA Review   SRE Review  Done   Production
```

### Agent-Based Workflow

Previewd uses a multi-agent SDLC workflow (when working with the project maintainer):

1. **product-technical-lead**: Clarifies requirements, writes Gherkin/BDD scenarios, creates issues
2. **qa-security-engineer**: Writes failing BDD tests, labels `workflow: bdd-ready`
3. **senior-developer**: Implements with strict TDD, opens PR
4. **code-reviewer**: Reviews for design, maintainability, SOLID principles
5. **qa-security-engineer**: QA validation, security audit
6. **sre-platform**: Reviews observability, infrastructure, deployment
7. **github-ops**: Merges, CI/CD deploys

**For external contributors**, the workflow is simplified but maintains quality gates:

1. Create/claim an issue
2. Write tests first (TDD)
3. Implement the solution
4. Open PR with comprehensive description
5. Address review feedback
6. Maintainer merges after approval

## Issue Lifecycle

### Issue Types

We use different issue types for different kinds of work:

| Type | Label | When to Use |
|------|-------|-------------|
| **Epic** | `type: epic` | Large initiative spanning multiple issues/milestones |
| **Story** | `type: story` | User-facing feature with business value |
| **Bug** | `bug` | Something isn't working as expected |
| **Task** | `type: task` | Technical work without direct user value (refactoring, CI/CD) |
| **Spike** | `type: spike` | Time-boxed research or investigation |

### Issue States

Issues move through defined workflow stages:

| Stage | Label | Description |
|-------|-------|-------------|
| **Backlog** | - | Not yet ready for work (epics being broken down) |
| **Triage** | `status: triage` | Needs review and prioritization |
| **Ready** | `status: ready` | Ready for development (Definition of Ready met) |
| **BDD Ready** | `workflow: bdd-ready` | BDD tests written, ready for implementation |
| **In Progress** | `status: in-progress` | Currently being worked on |
| **Code Review** | `workflow: ready-for-review` | PR open, awaiting code review |
| **QA Review** | `workflow: needs-qa` | Code approved, awaiting QA validation |
| **SRE Review** | `workflow: needs-sre` | QA passed, awaiting SRE review |
| **Done** | - | Merged to main, deployed |

### Priority Levels

| Priority | Label | Description | SLA |
|----------|-------|-------------|-----|
| **P0** | `priority: P0` | Critical: Production down, security vulnerability | Fix immediately |
| **P1** | `priority: P1` | High: Blocking feature, severe bug | Fix within 48 hours |
| **P2** | `priority: P2` | Medium: Important but not blocking | Fix within 1 week |
| **P3** | `priority: P3` | Low: Nice to have, minor issue | Fix when capacity available |

### Size Estimation

We use T-shirt sizing (considering AI assistance):

| Size | Label | Effort | Description |
|------|-------|--------|-------------|
| **XS** | `size: XS` | 1-2 hours | Trivial change |
| **S** | `size: S` | 2-4 hours | Small feature or bug fix |
| **M** | `size: M` | 1-2 days | Medium feature |
| **L** | `size: L` | 3-5 days | Large feature |
| **XL** | `size: XL` | 1-2 weeks | Very large - consider breaking down |

### Definition of Ready (Issues)

Before an issue can move to "Ready" status, it must have:

- [ ] Clear title and description
- [ ] Acceptance criteria defined (testable)
- [ ] Priority assigned
- [ ] Size estimated
- [ ] Milestone assigned (if applicable)
- [ ] Related issues linked (dependencies)
- [ ] Technical approach discussed (for complex issues)
- [ ] No blockers

## Pull Request Process

### Creating a Pull Request

1. **Create a branch** from `main`:
   ```bash
   git checkout -b feat/your-feature-name
   ```

   Branch naming conventions:
   - `feat/description` - New features
   - `fix/description` - Bug fixes
   - `docs/description` - Documentation only
   - `test/description` - Test improvements
   - `refactor/description` - Code refactoring
   - `chore/description` - Build, CI/CD, dependencies

2. **Write tests first** (TDD):
   ```bash
   # Write failing tests
   go test ./... # Should fail

   # Implement feature
   # ...

   # Tests should now pass
   go test ./...
   ```

3. **Commit your changes**:
   ```bash
   # Stage files explicitly (never use git add -A or git add .)
   git add controllers/previewenvironment_controller.go
   git add controllers/previewenvironment_controller_test.go

   git commit -m "feat(controller): add namespace creation logic

   Implement namespace-per-PR creation in reconciliation loop.

   - Add CreateNamespace function
   - Add unit tests for namespace creation
   - Update documentation

   Closes #42"
   ```

   **Commit message format**:
   ```
   <type>(<scope>): <subject>

   <body>

   Closes #<issue-number>
   ```

   Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

4. **Push your branch**:
   ```bash
   git push origin feat/your-feature-name
   ```

5. **Open a Pull Request**:
   - Use the PR template (auto-populated)
   - Reference the issue: `Closes #42`
   - Provide clear summary of changes
   - Describe testing performed
   - List any breaking changes
   - Add screenshots/demos if applicable

### PR Requirements (Definition of Done)

Before a PR can be merged, it must meet all these criteria:

#### Code Quality
- [ ] All tests pass (`make test`)
- [ ] Test coverage meets threshold (>80% for new code)
- [ ] Code follows Go best practices (gofmt, golangci-lint)
- [ ] No new security vulnerabilities introduced
- [ ] SOLID principles followed
- [ ] No code smells or technical debt

#### Testing
- [ ] Unit tests written (TDD)
- [ ] Integration tests added (if applicable)
- [ ] BDD tests pass (if applicable)
- [ ] Manual testing performed

#### Documentation
- [ ] Code comments updated (godoc)
- [ ] README updated (if user-facing changes)
- [ ] API docs updated (if API changes)
- [ ] CHANGELOG updated (unreleased section)
- [ ] Examples updated (if behavior changes)

#### Review
- [ ] Code review approved by maintainer
- [ ] QA review passed (if applicable)
- [ ] SRE review passed (if infrastructure changes)
- [ ] All comments addressed

#### Process
- [ ] Linked to issue (Closes #N)
- [ ] Milestone assigned
- [ ] Labels applied
- [ ] CI/CD passes
- [ ] No merge conflicts

### PR Review Process

1. **Self-review**: Review your own PR before requesting review
2. **Automated checks**: Wait for CI to pass
3. **Code review**: Maintainer reviews code quality and design
4. **QA review**: QA validates functionality and security
5. **SRE review**: SRE reviews infrastructure/observability (if applicable)
6. **Address feedback**: Make requested changes
7. **Merge**: Maintainer merges after all approvals

### Merge Strategy

- **Squash and merge** is the default for feature branches
- **Rebase and merge** for small, well-structured commits
- **Never force push** to `main` branch
- **Delete branch** after merge

## Coding Standards

### Go Best Practices

We follow [Effective Go](https://go.dev/doc/effective_go) and these additional standards:

- **Package naming**: Short, lowercase, no underscores (e.g., `github`, `argocd`)
- **Interface naming**: End with `-er` (e.g., `Reader`, `Reconciler`)
- **Error handling**: Always check errors immediately, wrap with context
- **Testing**: Table-driven tests with `t.Run()` for subtests
- **Comments**: Godoc comments for all exported functions/types
- **Dependency injection**: Use interfaces for testability

### Kubernetes Operator Patterns

- **Idempotent reconciliation**: Same input → same output (always)
- **Owner references**: Set for automatic garbage collection
- **Status updates**: Separate from spec updates
- **Use informers**: Never poll the API server
- **Finalizers**: For cleanup before deletion

### Project Structure

```
previewd/
├── api/v1alpha1/          # CRD types (Kubebuilder generated)
├── controllers/           # Reconciliation logic
├── internal/              # Private application code
│   ├── github/           # GitHub webhook integration
│   ├── ai/               # AI engine
│   ├── cost/             # Cost optimization
│   └── argocd/           # ArgoCD integration
├── config/                # Kubernetes manifests
├── docs/                  # Documentation
└── hack/                  # Development scripts
```

## Testing Requirements

### Test-Driven Development (TDD)

Previewd follows strict TDD:

1. **Red**: Write a failing test
2. **Green**: Write minimal code to pass
3. **Refactor**: Clean up code while keeping tests passing

### Test Levels

1. **Unit tests**: Individual functions, mocked dependencies
   ```bash
   go test ./controllers/...
   ```

2. **Integration tests**: Real Kubernetes cluster (kind)
   ```bash
   make test-integration
   ```

3. **E2E tests**: Full workflow (GitHub PR → environment → cleanup)
   ```bash
   make test-e2e
   ```

### Test Guidelines

- **No branching in tests**: Use table-driven tests for multiple cases
- **One assertion per test**: Keep tests simple and focused
- **Descriptive names**: `TestReconciler_CreatesNamespace_WhenPROpened`
- **Avoid "should"**: Use "It returns X" not "It should return X"
- **Test behavior, not implementation**: Test what, not how

### Coverage Requirements

- **New code**: >80% coverage
- **Critical paths**: 100% coverage
- **Overall project**: >90% for v1.0.0

## Documentation Requirements

### When to Update Documentation

Documentation must be updated in the **same PR** that changes functionality:

- **README.md**: User-facing features, installation, usage
- **API docs**: Interface changes, new endpoints
- **Architecture docs**: Component changes, design decisions
- **Examples**: Behavior changes, new features
- **Inline comments**: Function signatures, complex logic

### Documentation Standards

- **Use examples**: Show, don't just tell
- **Keep it current**: Outdated docs are worse than no docs
- **Test examples**: Examples must work (CI tests them)
- **Architecture Decision Records (ADRs)**: For significant decisions

## Community

### Communication Channels

- **GitHub Issues**: Bug reports, feature requests, tasks
- **GitHub Discussions**: Questions, ideas, general discussion
- **Pull Requests**: Code review, implementation discussion
- **Project Board**: Track progress, see what's being worked on

### Getting Help

- **Check documentation** first (README, docs/)
- **Search existing issues** for similar problems
- **Ask in Discussions** for general questions
- **Create an issue** for specific bugs or features

### Recognition

Contributors are recognized in:

- **CHANGELOG**: All contributions credited
- **README**: Major contributors highlighted
- **Release notes**: Feature contributors mentioned

## Project Board

Track project progress on our GitHub Issues and Projects boards.

The board has multiple views:

- **Board View**: Kanban-style workflow stages
- **Table View**: All issues with metadata
- **Roadmap View**: Timeline of milestones

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to Previewd!** Your contributions help make Kubernetes preview environments better for everyone.

For questions, reach out to [@mikelane](https://github.com/mikelane) or open an issue.
