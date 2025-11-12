# Security and Vulnerability Scanning

This document describes Previewd's vulnerability scanning strategy, tooling, and policies for managing security risks in dependencies and code.

## Overview

Previewd uses a defense-in-depth approach to vulnerability management with multiple scanning tools:

1. **govulncheck** - Official Go vulnerability database scanner
2. **Trivy** - Multi-purpose security scanner for dependencies and containers
3. **gosec** - Go source code security analyzer (SAST)

## Vulnerability Scanning Tools

### govulncheck

**Purpose:** Scans Go code for known vulnerabilities in dependencies using the official Go vulnerability database.

**What it checks:**
- Direct and indirect dependencies
- Standard library vulnerabilities
- Known CVEs in Go packages

**Configuration:**
- No configuration required
- Runs on all Go code: `govulncheck ./...`
- Fails immediately on any detected vulnerability

**Why we use it:**
- Official Go tooling, maintained by the Go team
- Integrated with Go's vulnerability database
- Zero false positives (only reports actually reachable vulnerabilities)
- No authentication required

**When it runs:**
- On every PR via CI
- Before merging to main branch

### Trivy (Filesystem Mode)

**Purpose:** Comprehensive security scanner that checks Go modules, container images, and other dependencies.

**What it checks:**
- Go module vulnerabilities (go.mod)
- License compliance issues
- Configuration issues
- Secret detection

**Configuration:**
```yaml
# Scan filesystem for vulnerabilities
scan-type: 'fs'
scan-ref: '.'
severity: 'CRITICAL,HIGH'
exit-code: '1'  # Fail CI on CRITICAL/HIGH
```

**Severity levels:**
- **CRITICAL**: Immediate action required, blocks merge
- **HIGH**: Must be addressed before release, blocks merge
- **MEDIUM**: Should be fixed, informational only (doesn't block)
- **LOW**: Informational only, tracked but doesn't block

**Why we use it:**
- Already used for Docker image scanning
- Supports multiple vulnerability databases (NVD, OSV, Go vulndb)
- No authentication required
- Active community and frequent updates
- SARIF output for GitHub Security integration

**When it runs:**
- On every PR via CI
- Before merging to main branch
- During Docker image builds

### gosec (SAST)

**Purpose:** Static analysis security testing for Go source code.

**What it checks:**
- Hardcoded credentials
- SQL injection vulnerabilities
- Unsafe cryptographic practices
- File path traversal issues
- Integer overflow risks
- And more (see [gosec rules](https://github.com/securego/gosec#available-rules))

**Configuration:**
```bash
gosec -no-fail -fmt sarif -out gosec.sarif ./...
```

**When it runs:**
- On every PR via CI
- Before merging to main branch

## Vulnerability Management Policy

### Severity-Based Response

| Severity | Response Time | Action | CI Blocking |
|----------|--------------|--------|-------------|
| **CRITICAL** | Immediate (24h) | Must fix or remove dependency | ✅ Yes |
| **HIGH** | 1 week | Must fix before next release | ✅ Yes |
| **MEDIUM** | 1 month | Fix in regular development cycle | ❌ No |
| **LOW** | Best effort | Track and fix when convenient | ❌ No |

### Version Policy

**v0.1.0 - v0.5.0 (Alpha):**
- Block: CRITICAL and HIGH
- Allow: MEDIUM and LOW (documented)
- Goal: Functional development with security awareness

**v0.6.0 - v0.9.0 (Beta):**
- Block: CRITICAL and HIGH
- Reduce: MEDIUM (must have mitigation plan)
- Allow: LOW (documented)
- Goal: Production-ready security posture

**v1.0.0+ (Production):**
- Block: CRITICAL, HIGH, and unmitigated MEDIUM
- Allow: LOW with documented acceptance
- Goal: Zero critical/high vulnerabilities

### Handling Vulnerabilities

#### 1. Update Dependencies

First, try to update to a patched version:

```bash
# Update specific dependency
go get -u github.com/example/package@latest

# Update all dependencies (carefully)
go get -u ./...

# Verify tests still pass
go test ./...
```

#### 2. Evaluate Impact

If no patch is available, assess:
- Is the vulnerable code path actually used?
- What's the attack surface?
- Is there a workaround?
- Can we replace the dependency?

#### 3. Document Known Issues

Create an entry in `docs/KNOWN_VULNERABILITIES.md`:

```markdown
## CVE-2024-XXXXX: Example Vulnerability

**Severity:** MEDIUM
**Package:** github.com/example/package v1.2.3
**Status:** Accepted (no patch available)

**Description:**
Brief description of the vulnerability.

**Impact Assessment:**
Why this is acceptable in our context.

**Mitigation:**
- What we're doing to reduce risk
- Monitoring plan
- Upgrade plan when patch available

**Accepted By:** [Name]
**Date:** 2025-11-11
**Review Date:** 2025-12-11
```

#### 4. Use Ignore Files (Sparingly)

Only for confirmed false positives:

```bash
# .trivyignore
# False positive: we don't use the vulnerable function
CVE-2024-XXXXX

# Accepted risk: documented in KNOWN_VULNERABILITIES.md
CVE-2024-YYYYY
```

## CI/CD Integration

### GitHub Actions Workflow

The `vulnerability-scan` job runs in CI:

```yaml
vulnerability-scan:
  name: Vulnerability Scan
  runs-on: ubuntu-latest
  steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Go
      uses: actions/setup-go@v5
      with:
        go-version-file: go.mod
        cache: true

    - name: Run govulncheck
      run: |
        go install golang.org/x/vuln/cmd/govulncheck@latest
        govulncheck ./...

    - name: Run Trivy (fail on CRITICAL/HIGH)
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
        severity: 'CRITICAL,HIGH'
        exit-code: '1'

    - name: Run Trivy (full report)
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
        format: 'table'
        severity: 'CRITICAL,HIGH,MEDIUM,LOW'
```

### GitHub Security Integration

Vulnerability findings are uploaded to GitHub Security tab via SARIF:

- **gosec results**: Code scanning alerts
- **Trivy filesystem results**: Dependency alerts
- **Trivy image results**: Container image alerts

View at: `https://github.com/mikelane/previewd/security`

## Local Development

### Running Scans Locally

**Check for Go vulnerabilities:**
```bash
make vulncheck
# or
govulncheck ./...
```

**Run Trivy filesystem scan:**
```bash
trivy fs . --severity CRITICAL,HIGH
```

**Run gosec:**
```bash
make security-scan
# or
gosec ./...
```

### Before Committing

Always run locally:
```bash
# Quick security check
make vulncheck

# Full security suite
make security-scan
```

## Dependency Management Best Practices

### 1. Minimal Dependencies

Only add dependencies when necessary:
- Evaluate alternatives
- Consider implementing small functionality yourself
- Prefer standard library when possible

### 2. Pin Versions

Always use specific versions in `go.mod`:
```go
require (
    github.com/example/package v1.2.3  // Good: specific version
)
```

Avoid:
```go
require (
    github.com/example/package latest  // Bad: unpredictable
)
```

### 3. Regular Updates

Update dependencies regularly:
```bash
# Check for updates
go list -u -m all

# Update with care
go get -u ./...
go mod tidy

# Test thoroughly
go test ./...
```

### 4. Audit New Dependencies

Before adding a new dependency:
- Check GitHub stars, activity, and maintenance
- Review open security issues
- Scan for vulnerabilities: `govulncheck ./...`
- Check license compatibility: `go-licenses check ./...`

## Historical Context

### Why Not Nancy?

Previously, we used [Nancy](https://github.com/sonatype-nexus-community/nancy) for dependency scanning. We removed it because:

1. **Authentication Required**: Nancy requires Sonatype OSS Index credentials
   - Free tier has rate limits
   - CI requires secret management
   - Additional operational complexity

2. **Overlapping Coverage**: govulncheck + Trivy provide comprehensive coverage
   - govulncheck: Official Go vulnerability database
   - Trivy: Multiple databases (NVD, OSV, Go vulndb)
   - No gaps in coverage

3. **Operational Simplicity**: Fewer tools, less configuration
   - No credentials to manage
   - No rate limit concerns
   - Easier to maintain

### Alternative Considered

**OSV-Scanner**: Google's Open Source Vulnerabilities scanner
- Pros: Aggregates multiple databases, no auth required
- Cons: Overlaps with Trivy, added complexity
- Decision: Trivy provides equivalent coverage with better GitHub integration

## Resources

- [Go Vulnerability Database](https://vuln.go.dev/)
- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
- [gosec Rules](https://github.com/securego/gosec#available-rules)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [GitHub Security Best Practices](https://docs.github.com/en/code-security)

## Support

For security concerns or questions:
- Open an issue with `area: security` label
- Email: mikelane@gmail.com
- Security vulnerabilities: See [SECURITY.md](../SECURITY.md)
