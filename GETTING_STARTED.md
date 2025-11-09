# Getting Started with Previewd Development

This guide will help you set up your development environment and start building Previewd.

## Prerequisites

### Required Software

1. **Go 1.21 or later**
   ```bash
   # Install via official installer: https://go.dev/dl/
   # Or use Homebrew on macOS:
   brew install go

   # Verify installation
   go version  # Should show 1.21 or later
   ```

2. **Docker Desktop**
   ```bash
   # Install from: https://www.docker.com/products/docker-desktop
   # Or Homebrew on macOS:
   brew install --cask docker

   # Verify
   docker version
   ```

3. **kubectl**
   ```bash
   # Homebrew on macOS:
   brew install kubectl

   # Or download binary: https://kubernetes.io/docs/tasks/tools/

   # Verify
   kubectl version --client
   ```

4. **kind (Kubernetes in Docker)**
   ```bash
   # Homebrew on macOS:
   brew install kind

   # Or Go install:
   go install sigs.k8s.io/kind@latest

   # Verify
   kind version
   ```

5. **Kubebuilder**
   ```bash
   # Install Kubebuilder 3.x
   curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)
   chmod +x kubebuilder
   sudo mv kubebuilder /usr/local/bin/

   # Verify
   kubebuilder version
   ```

### Optional but Recommended

- **Visual Studio Code** with Go extension
- **golangci-lint** for linting
- **git** for version control

## Development Setup

### Step 1: Clone Repository

```bash
# If you haven't already
cd ~/dev
git clone https://github.com/mikelane/previewd.git
cd previewd
```

### Step 2: Initialize Go Module

```bash
# Initialize Go module (if not already done)
go mod init github.com/mikelane/previewd

# Download dependencies
go mod tidy
```

### Step 3: Set Up Local Kubernetes Cluster

```bash
# Create kind cluster
kind create cluster --name previewd-dev

# Verify cluster is running
kubectl cluster-info --context kind-previewd-dev

# Set context
kubectl config use-context kind-previewd-dev
```

### Step 4: Scaffold Operator with Kubebuilder

```bash
# Initialize operator project
kubebuilder init --domain previewd.io --repo github.com/mikelane/previewd

# Create API and controller
kubebuilder create api \
  --group preview \
  --version v1alpha1 \
  --kind PreviewEnvironment \
  --resource \
  --controller

# This creates:
# - api/v1alpha1/previewenvironment_types.go (CRD definition)
# - controllers/previewenvironment_controller.go (reconciliation logic)
# - config/ directory (K8s manifests)
```

### Step 5: Install CRDs to Cluster

```bash
# Install Custom Resource Definitions
make install

# Verify CRD is installed
kubectl get crds | grep previewenvironment
```

### Step 6: Run Operator Locally

```bash
# Run operator (watches cluster)
make run

# In another terminal, create a sample PreviewEnvironment
kubectl apply -f config/samples/preview_v1alpha1_previewenvironment.yaml

# Check logs in the first terminal to see reconciliation
```

## Project Structure

After scaffolding with Kubebuilder, your project should look like:

```
previewd/
├── api/
│   └── v1alpha1/
│       ├── previewenvironment_types.go    # CRD definition
│       ├── groupversion_info.go           # API version info
│       └── zz_generated.deepcopy.go       # Auto-generated
├── config/
│   ├── crd/                               # CRD manifests
│   ├── default/                           # Default kustomization
│   ├── manager/                           # Operator deployment
│   ├── rbac/                              # RBAC permissions
│   └── samples/                           # Example CRs
├── controllers/
│   ├── previewenvironment_controller.go   # Reconciliation logic
│   └── suite_test.go                      # Test setup
├── go.mod                                 # Go dependencies
├── go.sum                                 # Dependency checksums
├── main.go                                # Operator entrypoint
├── Makefile                               # Build commands
├── PROJECT                                # Kubebuilder metadata
├── README.md                              # Project README
├── CLAUDE.md                              # Context for Claude
├── ARCHITECTURE.md                        # Architecture docs
└── ROADMAP.md                             # Development timeline
```

## Common Development Commands

### Building & Running

```bash
# Build operator binary
make build

# Run tests
make test

# Run linter
make lint

# Run operator locally
make run

# Build Docker image
make docker-build IMG=your-registry/previewd:latest

# Push Docker image
make docker-push IMG=your-registry/previewd:latest

# Deploy to cluster
make deploy IMG=your-registry/previewd:latest
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v -run TestPreviewEnvironmentController ./controllers/...

# Run tests with race detector
go test -race ./...
```

### Kubernetes Operations

```bash
# Install CRDs
make install

# Uninstall CRDs
make uninstall

# Deploy operator
make deploy

# Undeploy operator
make undeploy

# View logs (when deployed)
kubectl logs -n previewd-system deployment/previewd-controller-manager -f
```

### Kind Cluster Management

```bash
# Create cluster
kind create cluster --name previewd-dev

# Delete cluster
kind delete cluster --name previewd-dev

# Load local image to kind
kind load docker-image your-registry/previewd:latest --name previewd-dev

# Get cluster info
kubectl cluster-info --context kind-previewd-dev
```

## First Development Task: Hello World

Let's create a simple "Hello World" reconciliation to verify everything works.

### 1. Edit the Reconcile Function

Edit `controllers/previewenvironment_controller.go`:

```go
func (r *PreviewEnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // Fetch the PreviewEnvironment instance
    env := &previewv1alpha1.PreviewEnvironment{}
    if err := r.Get(ctx, req.NamespacedName, env); err != nil {
        if errors.IsNotFound(err) {
            // Object not found, could have been deleted
            return ctrl.Result{}, nil
        }
        // Error reading the object
        return ctrl.Result{}, err
    }

    // Hello World: Just log the PR number
    log.Info("Reconciling PreviewEnvironment",
        "name", env.Name,
        "prNumber", env.Spec.PRNumber,
        "repository", env.Spec.Repository)

    // Update status (Hello World example)
    env.Status.Phase = "Hello"
    env.Status.URL = fmt.Sprintf("https://pr-%d.preview.example.com", env.Spec.PRNumber)

    if err := r.Status().Update(ctx, env); err != nil {
        log.Error(err, "Failed to update PreviewEnvironment status")
        return ctrl.Result{}, err
    }

    return ctrl.Result{}, nil
}
```

### 2. Define the Spec and Status

Edit `api/v1alpha1/previewenvironment_types.go`:

```go
// PreviewEnvironmentSpec defines the desired state of PreviewEnvironment
type PreviewEnvironmentSpec struct {
    // PR number
    PRNumber int `json:"prNumber"`

    // Repository (e.g., "myorg/myapp")
    Repository string `json:"repository"`

    // Branch name
    Branch string `json:"branch"`
}

// PreviewEnvironmentStatus defines the observed state of PreviewEnvironment
type PreviewEnvironmentStatus struct {
    // Phase (e.g., "Pending", "Ready", "Failed")
    Phase string `json:"phase,omitempty"`

    // URL of the preview environment
    URL string `json:"url,omitempty"`
}
```

### 3. Generate Code and Update CRDs

```bash
# Generate deepcopy code
make generate

# Update CRD manifests
make manifests

# Install updated CRDs
make install
```

### 4. Test It

```bash
# Run operator
make run

# In another terminal, create a sample PreviewEnvironment
cat <<EOF | kubectl apply -f -
apiVersion: preview.previewd.io/v1alpha1
kind: PreviewEnvironment
metadata:
  name: test-pr-123
spec:
  prNumber: 123
  repository: "myorg/myapp"
  branch: "feature/test"
EOF

# Check status
kubectl get previewenvironments test-pr-123 -o yaml

# You should see:
# status:
#   phase: Hello
#   url: https://pr-123.preview.example.com
```

## Learning Resources

### Go
- [A Tour of Go](https://tour.golang.org/) - Interactive tutorial
- [Effective Go](https://go.dev/doc/effective_go) - Best practices
- [Go by Example](https://gobyexample.com/) - Practical examples

### Kubernetes
- [Kubernetes Basics](https://kubernetes.io/docs/tutorials/kubernetes-basics/)
- [client-go Examples](https://github.com/kubernetes/client-go/tree/master/examples)

### Operators
- [Kubebuilder Book](https://book.kubebuilder.io/) - Official guide
- [Programming Kubernetes](https://www.oreilly.com/library/view/programming-kubernetes/9781492047094/) - O'Reilly book
- [Operator Best Practices](https://sdk.operatorframework.io/docs/best-practices/)

## Troubleshooting

### Issue: CRD not found

**Solution:**
```bash
# Reinstall CRDs
make install

# Verify
kubectl get crds | grep previewenvironment
```

### Issue: Operator not reconciling

**Solution:**
```bash
# Check if operator is running
# If running via `make run`, check terminal logs

# If deployed to cluster:
kubectl logs -n previewd-system deployment/previewd-controller-manager -f

# Check if CR exists
kubectl get previewenvironments
```

### Issue: Go dependencies not resolving

**Solution:**
```bash
# Update dependencies
go mod tidy

# If using vendor:
go mod vendor
```

### Issue: Kind cluster not accessible

**Solution:**
```bash
# Delete and recreate cluster
kind delete cluster --name previewd-dev
kind create cluster --name previewd-dev

# Verify
kubectl cluster-info --context kind-previewd-dev
```

## Next Steps

Once you have the Hello World working:

1. **Read ARCHITECTURE.md** - Understand the system design
2. **Read ROADMAP.md** - See the development plan
3. **Start Phase 0, Week 1** - Go learning path
4. **Join discussions** - Ask questions in GitHub Discussions

## Getting Help

- **Documentation:** See README.md, ARCHITECTURE.md, ROADMAP.md
- **GitHub Issues:** https://github.com/mikelane/previewd/issues
- **Contact:** mikelane@gmail.com

---

Happy coding! 🚀
