# Repository Setup - Quick Reference

This document provides a quick reference for setting up the Previewd repository with all necessary configurations.

## One-Time Setup Script

You can run these commands sequentially to complete the entire repository setup:

```bash
#!/bin/bash

# GitHub repository setup script for Previewd
# Run this after pushing the CI/CD configuration files to the repository

set -e

REPO="mikelane/previewd"

echo "🚀 Setting up Previewd repository configuration..."

# 1. Enable repository features
echo "📋 Enabling repository features..."
gh repo edit $REPO --enable-discussions
gh repo edit $REPO --enable-vulnerability-alerts
gh repo edit $REPO --enable-automated-security-fixes
gh repo edit $REPO --enable-auto-merge
gh repo edit $REPO --enable-issues
gh repo edit $REPO --enable-projects
gh repo edit $REPO --disable-wiki

# 2. Set default branch and merge settings
echo "🌿 Configuring branch and merge settings..."
gh repo edit $REPO --default-branch main
gh api repos/$REPO \
  --method PATCH \
  --field allow_squash_merge=true \
  --field allow_merge_commit=false \
  --field allow_rebase_merge=false \
  --field delete_branch_on_merge=true

# 3. Set up branch protection for main
echo "🛡️  Setting up branch protection..."
gh api repos/$REPO/branches/main/protection \
  --method PUT \
  --input - <<EOF
{
  "required_status_checks": {
    "strict": true,
    "contexts": [
      "Lint",
      "Format Check",
      "Static Analysis",
      "Security Scan",
      "Vulnerability Scan",
      "Unit Tests",
      "Branch Coverage",
      "Mutation Testing",
      "Race Condition Detection",
      "E2E Tests",
      "Build",
      "Docker Build",
      "License Compliance",
      "Documentation Check"
    ]
  },
  "enforce_admins": false,
  "required_pull_request_reviews": {
    "dismissal_restrictions": {},
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": true,
    "required_approving_review_count": 0,
    "require_last_push_approval": false
  },
  "restrictions": null,
  "required_linear_history": true,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "block_creations": false,
  "required_conversation_resolution": true,
  "lock_branch": false,
  "allow_fork_syncing": true
}
EOF

# 4. Configure workflow permissions
echo "🔐 Setting workflow permissions..."
gh api repos/$REPO \
  --method PATCH \
  --field default_workflow_permissions="write" \
  --field can_approve_pull_request_reviews=true

# 5. Create v0.1.0 milestone
echo "🎯 Creating v0.1.0 milestone..."
gh api repos/$REPO/milestones \
  --method POST \
  --field title="v0.1.0" \
  --field description="Initial release - Core operator functionality" \
  --field state="open" || echo "Milestone already exists"

# 6. Set up repository labels
echo "🏷️  Creating repository labels..."
gh label create "bug" --description "Something isn't working" --color "d73a4a" --repo $REPO || true
gh label create "enhancement" --description "New feature or request" --color "a2eeef" --repo $REPO || true
gh label create "documentation" --description "Improvements or additions to documentation" --color "0075ca" --repo $REPO || true
gh label create "dependencies" --description "Pull requests that update a dependency file" --color "0366d6" --repo $REPO || true
gh label create "github-actions" --description "Pull requests that update GitHub Actions code" --color "000000" --repo $REPO || true
gh label create "go" --description "Pull requests that update Go code" --color "00ADD8" --repo $REPO || true
gh label create "docker" --description "Pull requests that update Docker code" --color "0db7ed" --repo $REPO || true
gh label create "security" --description "Security-related issues or PRs" --color "ee0701" --repo $REPO || true
gh label create "triage" --description "Needs triage" --color "fbca04" --repo $REPO || true
gh label create "good first issue" --description "Good for newcomers" --color "7057ff" --repo $REPO || true
gh label create "help wanted" --description "Extra attention is needed" --color "008672" --repo $REPO || true
gh label create "question" --description "Further information is requested" --color "d876e3" --repo $REPO || true

echo "✅ Repository setup complete!"
echo ""
echo "Next steps:"
echo "1. Enable Codecov: Visit https://codecov.io, sign in with GitHub, add $REPO (OIDC is configured automatically)"
echo "2. Review and customize issue/PR templates as needed"
echo "3. Create your first issue to start development workflow"
echo "4. Test the pipeline with a small PR to verify auto-merge works"
```

## Manual Setup Steps

If you prefer to set up manually or need to troubleshoot, follow these steps:

### 1. Codecov Setup

```bash
# Go to https://codecov.io and add your repository
# OIDC authentication is configured automatically in the CI workflow
# No token required - just add the repository to Codecov!
```

### 2. Branch Protection

```bash
gh api repos/mikelane/previewd/branches/main/protection \
  --method PUT \
  --input .github/scripts/branch-protection.json
```

### 3. Workflow Permissions

```bash
gh api repos/mikelane/previewd \
  --method PATCH \
  --field default_workflow_permissions="write" \
  --field can_approve_pull_request_reviews=true
```

### 4. Enable Features

```bash
gh repo edit mikelane/previewd --enable-auto-merge
gh repo edit mikelane/previewd --enable-discussions
gh repo edit mikelane/previewd --enable-vulnerability-alerts
```

## Verification Checklist

After running the setup script, verify:

- [ ] Branch protection is active on `main`
- [ ] Auto-merge is enabled
- [ ] Codecov repository is added (no token required with OIDC)
- [ ] Workflow permissions are write
- [ ] All required status checks are configured
- [ ] Dependabot is enabled
- [ ] Security policies are in place

## Testing the Pipeline

Create a test PR to verify the pipeline:

```bash
# Create a feature branch
git checkout -b test/ci-pipeline

# Make a small change
echo "# Test" >> README.md

# Commit and push
git add README.md
git commit -m "test: verify CI pipeline"
git push origin test/ci-pipeline

# Create PR
gh pr create --title "test: CI pipeline verification" \
  --body "Testing the CI/CD pipeline setup"

# Watch the checks run
gh pr checks --watch
```

## Troubleshooting

### Issue: Branch protection not applying

**Solution**: Ensure you have admin access and the branch name is exactly `main`

```bash
gh api repos/mikelane/previewd/branches/main/protection
```

### Issue: Auto-merge not working

**Solution**: Check workflow permissions and auto-merge is enabled

```bash
gh repo view mikelane/previewd --json autoMergeAllowed
```

### Issue: Codecov not uploading

**Solution**: Verify repository is added to Codecov and OIDC is enabled

- Visit https://codecov.io and ensure `mikelane/previewd` is added
- Check that the CI workflow has `use_oidc: true` in the Codecov upload step
- Verify workflow has `id-token: write` permission

## Repository Structure

After setup, your repository should have:

```
previewd/
├── .github/
│   ├── CODEOWNERS
│   ├── FUNDING.yml
│   ├── SECURITY.md
│   ├── dependabot.yml
│   ├── markdown-link-check-config.json
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.yml
│   │   ├── feature_request.yml
│   │   └── config.yml
│   ├── pull_request_template.md
│   └── workflows/
│       ├── ci.yml
│       ├── auto-merge.yml
│       ├── test.yml
│       ├── lint.yml
│       └── test-e2e.yml
├── docs/
│   ├── CI_CD_SETUP.md
│   └── REPOSITORY_SETUP.md
├── codecov.yml
└── ... (other project files)
```

## Next Steps

1. ✅ Complete repository setup
2. ✅ Verify CI/CD pipeline
3. 📝 Create first issue for development work
4. 🔧 Start development workflow (issue → branch → PR → merge)
5. 📊 Monitor CI/CD metrics
6. 🎉 Celebrate world-class engineering!

## Support

Questions? Contact @mikelane or open an issue in the repository.
