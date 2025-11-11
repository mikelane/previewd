# Workflow Quick Reference

Quick reference guide for the Previewd development workflow.

## Creating Issues

### Bug Report
```bash
gh issue create --template bug_report.yml --title "[Bug]: Description"
```

### Feature Request
```bash
gh issue create --template feature_request.yml --title "[Feature]: Description"
```

### Epic
```bash
gh issue create --template epic.yml --title "[Epic]: Description"
```

### Spike/Research
```bash
gh issue create --template spike.yml --title "[Spike]: Description"
```

### Task
```bash
gh issue create --template task.yml --title "[Task]: Description"
```

## Common Label Combinations

### New Feature (User-Facing)
```bash
gh issue create --label "enhancement,priority: P2,size: M,phase: v0.1.0,status: triage"
```

### Bug Fix (Critical)
```bash
gh issue create --label "bug,priority: P0,size: S,status: triage"
```

### Technical Task
```bash
gh issue create --label "type: task,priority: P3,size: S,status: triage"
```

### Research Spike
```bash
gh issue create --label "type: spike,priority: P1,size: XS,status: triage"
```

## Issue Workflow

### 1. Claim an Issue
```bash
gh issue edit <issue-number> --add-assignee @me
gh issue edit <issue-number> --add-label "status: in-progress"
gh issue edit <issue-number> --remove-label "status: ready"
```

### 2. Create Feature Branch
```bash
git checkout main
git pull origin main
git checkout -b feat/description  # or fix/, docs/, test/, refactor/, chore/
```

### 3. Write Tests First (TDD)
```bash
# Write failing tests
go test ./...  # Should fail

# Implement feature
# ...

# Tests should pass
go test ./...
```

### 4. Commit Changes
```bash
# Stage files explicitly (never git add -A)
git add file1.go file2_test.go

# Commit with conventional commit format
git commit -m "feat(scope): description

Detailed description of changes.

- Bullet point 1
- Bullet point 2

Closes #<issue-number>"
```

### 5. Push and Create PR
```bash
git push origin feat/description
gh pr create --title "feat(scope): description" --body "Closes #<issue-number>

## Summary
Description of changes...

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing performed

## Documentation
- [ ] Code comments updated
- [ ] README updated (if applicable)
- [ ] CHANGELOG updated"
```

## PR Workflow

### Open Draft PR (Work in Progress)
```bash
gh pr create --draft --title "WIP: feat(scope): description"
```

### Mark PR Ready for Review
```bash
gh pr ready
```

### Request Specific Reviewer
```bash
gh pr edit <pr-number> --add-reviewer mikelane
```

### Check PR Status
```bash
gh pr status
gh pr view <pr-number>
gh pr checks <pr-number>
```

### Update PR After Feedback
```bash
# Make changes
git add file.go
git commit -m "fix: address review feedback"
git push origin feat/description

# PR automatically updates
```

### Merge PR (Maintainer Only)
```bash
# Squash and merge (default)
gh pr merge <pr-number> --squash --delete-branch

# Merge without squash (for well-structured commits)
gh pr merge <pr-number> --merge --delete-branch
```

## Project Board

### View Project
```bash
gh project view 3 --owner mikelane
```

### Add Issue to Project
```bash
# Issues are auto-added when created, but can be added manually:
gh project item-add 3 --owner mikelane --url https://github.com/mikelane/previewd/issues/<number>
```

## Milestones

### View Milestones
```bash
gh api repos/mikelane/previewd/milestones | jq '.[] | {number, title, open_issues, closed_issues, due_on}'
```

### Assign Issue to Milestone
```bash
gh issue edit <issue-number> --milestone "v0.1.0: Core Operator (No AI)"
```

### View Milestone Progress
```bash
gh issue list --milestone "v0.1.0: Core Operator (No AI)"
```

## Labels

### View All Labels
```bash
gh label list
```

### Add Label to Issue
```bash
gh issue edit <issue-number> --add-label "priority: P1"
```

### Remove Label from Issue
```bash
gh issue edit <issue-number> --remove-label "status: triage"
```

### Search Issues by Label
```bash
gh issue list --label "priority: P0"
gh issue list --label "status: in-progress"
gh issue list --label "phase: v0.1.0"
```

## Common Searches

### My Open Issues
```bash
gh issue list --assignee @me
```

### Issues Ready for Work
```bash
gh issue list --label "status: ready"
```

### Issues in Code Review
```bash
gh issue list --label "workflow: ready-for-review"
```

### High Priority Issues
```bash
gh issue list --label "priority: P0,priority: P1"
```

### Issues by Milestone
```bash
gh issue list --milestone "v0.1.0: Core Operator (No AI)"
```

### Issues by Area
```bash
gh issue list --label "area: kubernetes"
```

## Testing

### Run All Tests
```bash
make test
```

### Run Tests with Coverage
```bash
make test-coverage
```

### Run Integration Tests
```bash
make test-integration
```

### Run Specific Test
```bash
go test -v -run TestControllerReconcile ./controllers/...
```

## Quality Checks

### Format Code
```bash
gofmt -w .
goimports -w .
```

### Lint Code
```bash
golangci-lint run
```

### Run All Checks (before PR)
```bash
make verify  # formats, lints, tests
```

## Documentation

### Update Godoc Comments
```go
// Package controllers implements Kubernetes controllers for preview environments.
package controllers

// Reconcile implements the reconciliation loop for PreviewEnvironment resources.
// It ensures the desired state matches the actual state in the cluster.
func (r *PreviewEnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ...
}
```

### Generate Godoc HTML
```bash
godoc -http=:6060
# Visit http://localhost:6060/pkg/github.com/mikelane/previewd/
```

## Common Workflows

### Start New Feature
```bash
# 1. Find or create issue
gh issue create --template feature_request.yml

# 2. Assign to yourself
gh issue edit <number> --add-assignee @me

# 3. Update labels
gh issue edit <number> --add-label "status: in-progress"
gh issue edit <number> --remove-label "status: ready"

# 4. Create branch
git checkout -b feat/description

# 5. Write tests (TDD)
# 6. Implement feature
# 7. Commit and push
# 8. Open PR
gh pr create
```

### Fix a Bug
```bash
# 1. Create bug report
gh issue create --template bug_report.yml --label "bug,priority: P1"

# 2. Create branch
git checkout -b fix/description

# 3. Write failing test that reproduces bug
# 4. Fix bug (test should pass)
# 5. Commit and push
git commit -m "fix: description

Fixes bug where...

Closes #<issue-number>"
gh pr create
```

### Research/Spike
```bash
# 1. Create spike
gh issue create --template spike.yml --label "type: spike,priority: P2"

# 2. Time-box research (e.g., 4 hours)

# 3. Document findings in issue comment
gh issue comment <number> --body "## Research Findings

**Question:** Can we integrate with X?

**Answer:** Yes, using approach Y.

**Recommendation:** Use library Z.

**Proof of Concept:** [link to code]

**Next Steps:**
- Create issue for implementation
- Update ADR
"

# 4. Close spike
gh issue close <number>
```

## Shortcuts

### Create Issue + PR in One Go
```bash
# Create issue
ISSUE=$(gh issue create --template task.yml --title "Task: description" --label "type: task" --json number --jq .number)

# Create branch and PR
git checkout -b task/description
# ... make changes ...
git commit -m "chore: description

Closes #$ISSUE"
gh pr create --title "chore: description" --body "Closes #$ISSUE"
```

### Bulk Label Issues
```bash
# Label all open issues in milestone with phase
gh issue list --milestone "v0.1.0" --state open --json number --jq '.[].number' | xargs -I {} gh issue edit {} --add-label "phase: v0.1.0"
```

### Find Stale Issues
```bash
gh issue list --label "status: in-progress" --json number,title,updatedAt --jq '.[] | select(.updatedAt < (now - 604800 | strftime("%Y-%m-%dT%H:%M:%SZ")))'
```

## Resources

- [CONTRIBUTING.md](../CONTRIBUTING.md) - Full workflow documentation
- [PRODUCT_LIFECYCLE.md](PRODUCT_LIFECYCLE.md) - Detailed lifecycle
- [GitHub Project Board](https://github.com/mikelane/previewd/issues) - Live tracking
- [Milestones](https://github.com/mikelane/previewd/milestones) - Release planning

---

**Tip:** Bookmark this page for quick reference during daily development!
