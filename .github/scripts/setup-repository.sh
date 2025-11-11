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
