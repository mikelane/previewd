# Development Guide

This guide covers local development setup, workflows, and best practices for contributing to Previewd.

## Prerequisites

- **Go 1.25+** - [Install](https://go.dev/dl/)
- **Docker** - [Install](https://docs.docker.com/get-docker/)
- **kind** - [Install](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- **kubectl** - [Install](https://kubernetes.io/docs/tasks/tools/)
- **pre-commit** - [Install](#pre-commit-hooks) (recommended)
- **goimports** - `go install golang.org/x/tools/cmd/goimports@latest`
- **golangci-lint** - [Install](https://golangci-lint.run/welcome/install/)

## Quick Start

```bash
# Clone the repository
git clone https://github.com/mikelane/previewd.git
cd previewd

# Install dependencies
go mod download

# Install pre-commit hooks (recommended)
pre-commit install

# Run tests
make test

# Build the operator
make build
```

## Pre-commit Hooks

This project uses [pre-commit](https://pre-commit.com/) to catch issues before CI.

### Installation

```bash
# Install pre-commit (choose one)
pip install pre-commit          # via pip
brew install pre-commit         # via homebrew
conda install -c conda-forge pre-commit  # via conda

# Install the git hooks
pre-commit install

# Test hooks on all files
pre-commit run --all-files
```

### What Gets Checked

- **gofmt** - Go code formatting
- **goimports** - Import organization
- **go vet** - Go static analysis
- **golangci-lint** - Comprehensive linting
- **trailing-whitespace** - No trailing whitespace
- **end-of-file-fixer** - Newline at end of files
- **check-yaml** - YAML syntax validation
- **check-json** - JSON syntax validation

### Bypassing Hooks

For emergencies only:
```bash
git commit --no-verify
```

**Note:** CI will still enforce these checks, so bypassing locally will cause CI failures.

## Development Workflow

### 1. Create Feature Branch

```bash
git checkout -b feat/your-feature-name
```

Branch naming conventions:
- `feat/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation only
- `test/` - Test improvements
- `refactor/` - Code refactoring

### 2. Make Changes

Follow TDD approach:
1. Write failing test
2. Implement minimal code to pass
3. Refactor while keeping tests green

### 3. Run Tests Locally

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test -v -run TestYourTest ./pkg/...
```

### 4. Format and Lint

Pre-commit hooks handle this automatically, but you can run manually:

```bash
# Format code
gofmt -w .
goimports -w .

# Run linters
golangci-lint run

# Or run all quality checks
make lint
```

### 5. Commit Changes

```bash
# Pre-commit hooks run automatically
git add .
git commit -m "feat(scope): brief description"
```

Commit message format:
```
<type>(<scope>): <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

### 6. Push and Create PR

```bash
git push origin feat/your-feature-name
```

Then create a Pull Request on GitHub.

## Testing

### Unit Tests

```bash
# Run all unit tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./internal/controller/...

# Run with race detector
go test -race ./...
```

### Integration Tests

```bash
# Run integration tests (requires kind cluster)
make test-integration
```

### E2E Tests

```bash
# Run end-to-end tests
make test-e2e
```

## Local Kubernetes Development

### Create Local Cluster

```bash
# Create kind cluster
make kind-create

# Verify cluster is running
kubectl cluster-info --context kind-kind
```

### Install CRDs

```bash
# Install Custom Resource Definitions
make install
```

### Run Operator Locally

```bash
# Run operator against local cluster
make run
```

### Deploy to Cluster

```bash
# Build and deploy operator to cluster
make deploy

# View logs
kubectl logs -n previewd-system deployment/previewd-controller-manager -f
```

### Clean Up

```bash
# Undeploy operator
make undeploy

# Delete kind cluster
make kind-delete
```

## Code Organization

```
previewd/
├── api/v1alpha1/           # CRD definitions
├── cmd/                    # Main application entry point
├── config/                 # Kubernetes manifests
│   ├── crd/               # CRD manifests
│   ├── manager/           # Operator deployment
│   └── rbac/              # RBAC permissions
├── internal/              # Private application code
│   ├── controller/        # Reconciliation logic
│   ├── github/            # GitHub integration
│   └── cost/              # Cost estimation
├── test/                  # Integration/E2E tests
└── hack/                  # Development scripts
```

## Troubleshooting

### Pre-commit Hooks Fail

```bash
# Run hooks manually to see errors
pre-commit run --all-files

# Update hooks to latest version
pre-commit autoupdate

# Clean and reinstall hooks
pre-commit clean
pre-commit install
```

### Tests Fail Locally

```bash
# Clean test cache
go clean -testcache

# Ensure dependencies are up to date
go mod download
go mod tidy

# Rebuild generated code
make generate
```

### CRD Changes Not Reflected

```bash
# Regenerate CRDs and manifests
make manifests

# Reinstall CRDs
make install
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## Resources

- [Kubebuilder Book](https://book.kubebuilder.io/) - Operator development guide
- [Go Best Practices](https://go.dev/doc/effective_go) - Effective Go
- [Pre-commit Documentation](https://pre-commit.com/) - Pre-commit hooks
