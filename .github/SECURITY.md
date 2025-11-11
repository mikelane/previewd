# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          |
| ------- | ------------------ |
| 0.2.x   | :white_check_mark: |
| 0.1.x   | :white_check_mark: |
| < 0.1.0 | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to:

**mikelane@gmail.com**

Please include the following information:

- Type of issue (e.g., buffer overflow, SQL injection, cross-site scripting, etc.)
- Full paths of source file(s) related to the manifestation of the issue
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit the issue

This information will help us triage your report more quickly.

## Response Timeline

- **Initial Response**: Within 48 hours of report submission
- **Triage**: Within 1 week of report submission
- **Fix Development**: Depends on severity and complexity
- **Release**: Security patches are released as soon as possible after verification

## Disclosure Policy

- Security issues are disclosed once a fix is available
- We will credit reporters (unless they prefer to remain anonymous)
- We follow responsible disclosure practices

## Security Update Process

1. Security issue is reported privately
2. Issue is confirmed and assessed for severity
3. Fix is developed in private
4. Fix is tested thoroughly
5. Security advisory is published
6. Patch is released
7. Public disclosure of the issue

## Security-Related Configuration

### RBAC Permissions

Previewd requires the following Kubernetes permissions:

- **Namespaces**: create, delete, list, watch
- **Deployments**: create, delete, list, watch, update
- **Services**: create, delete, list, watch
- **Ingresses**: create, delete, list, watch
- **ConfigMaps**: create, delete, list, watch
- **Secrets**: create, delete, list, watch (for AI API keys)

### Network Policies

We recommend deploying Previewd with appropriate network policies to restrict:

- Operator access to Kubernetes API
- AI service access (if using external LLM APIs)
- Preview environment egress/ingress

### Secrets Management

- AI API keys should be stored in Kubernetes Secrets
- Use RBAC to restrict access to secrets
- Consider using external secret management (e.g., Vault, AWS Secrets Manager)

## Security Best Practices

### For Users

1. **Keep Previewd Updated**: Always use the latest stable version
2. **Review RBAC**: Ensure the operator has minimal required permissions
3. **Network Isolation**: Use network policies to isolate preview environments
4. **Resource Limits**: Set resource quotas to prevent DoS
5. **Audit Logs**: Enable Kubernetes audit logging
6. **Secret Rotation**: Regularly rotate API keys and credentials

### For Contributors

1. **Dependency Scanning**: All dependencies are scanned for vulnerabilities
2. **Static Analysis**: Code is analyzed with gosec
3. **Code Review**: All changes require review before merge
4. **Tests**: Security-critical code requires comprehensive tests
5. **Least Privilege**: Follow principle of least privilege

## Known Security Considerations

### AI API Keys

Previewd may store OpenAI or other LLM API keys in Kubernetes Secrets. Ensure:

- Secrets are encrypted at rest
- RBAC restricts access to secrets
- API keys are rotated regularly
- Rate limiting is configured

### Preview Environment Isolation

Preview environments may execute untrusted code. Ensure:

- Network policies restrict environment communication
- Resource limits prevent resource exhaustion
- Namespaces provide logical isolation
- Pod security policies/standards are enforced

### GitHub Webhooks

Previewd receives GitHub webhook events. Ensure:

- Webhook secrets are validated
- Input is sanitized and validated
- Rate limiting is configured
- Only necessary repositories send webhooks

## Security Tooling

Our CI/CD pipeline includes:

- **gosec**: Go security scanner
- **govulncheck**: Go vulnerability scanner
- **Trivy**: Container image scanner
- **Nancy**: Dependency vulnerability scanner
- **CodeQL**: Semantic code analysis
- **Dependabot**: Automated dependency updates

## Bug Bounty Program

We do not currently have a bug bounty program. However, we deeply appreciate security researchers who responsibly disclose vulnerabilities and will provide public credit (unless you prefer to remain anonymous).

## Contact

For security-related questions or concerns:

- Email: mikelane@gmail.com
- GitHub: @mikelane

## Attribution

We would like to thank the following individuals and organizations for responsibly disclosing security issues:

<!-- List will be maintained here -->
