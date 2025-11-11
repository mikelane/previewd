# Project Management Guide

## Overview

This document describes the GitHub project management structure for Previewd, including project boards, issue workflow, labels, milestones, and automation.

## Project Board Structure

**Project**: [Previewd Development](https://github.com/mikelane/previewd/issues)

### Status Field (Workflow Columns)

Issues flow through these states:

1. **Todo** - Backlog items, not yet ready for work
2. **In Progress** - Actively being worked on
3. **Done** - Completed and merged

### Workflow Stage Field (SDLC Tracking)

Detailed workflow stages for agent coordination:

1. **Backlog** - Needs triage and prioritization
2. **Triage** - Being reviewed and scoped
3. **Ready** - Ready for BDD test creation
4. **BDD Ready** - BDD tests written, ready for implementation
5. **In Progress** - Implementation underway
6. **Code Review** - PR open, awaiting code review
7. **QA Review** - Code approved, awaiting QA validation
8. **SRE Review** - QA passed, awaiting SRE approval
9. **Done** - Merged and deployed

### Custom Fields

- **Priority**: P0-Critical, P1-High, P2-Medium, P3-Low
- **Size**: XS (1-2h), S (half day), M (1-2 days), L (3-5 days), XL (1-2 weeks)
- **Phase**: v0.1.0-Core, v0.2.0-AI, v0.3.0-Polish, Future
- **Workflow Stage**: (see above)

## Label Taxonomy

### Priority Labels (Required)

All issues must have a priority label:

- `priority: P0` - Critical: Production down, security vulnerability (red: #b60205)
- `priority: P1` - High: Blocking feature, severe bug (orange: #d93f0b)
- `priority: P2` - Medium: Important but not blocking (yellow: #fbca04)
- `priority: P3` - Low: Nice to have, minor issue (green: #0e8a16)

### Type Labels (Required)

All issues must have a type label:

- `type: epic` - Large initiative spanning multiple issues (purple: #3e4b9e)
- `type: feature` - New feature implementation (gray: #ededed)
- `type: bug` - Something isn't working (red: #d73a4a)
- `type: chore` - Maintenance tasks and code cleanup (gray: #ededed)
- `type: test` - Test implementation and testing infrastructure (gray: #ededed)
- `type: documentation` - Documentation updates and improvements (gray: #ededed)
- `type: spike` - Research or investigation work (pink: #d876e3)
- `type: story` - User story with business value (purple: #5319e7)
- `type: task` - Technical task without direct user value (blue: #1d76db)

### Area Labels (Optional)

Categorize by functional area:

- `area: api` - API and CRD definitions (gray: #ededed)
- `area: controller` - Controller and reconciliation logic (gray: #ededed)
- `area: kubernetes` - Kubernetes core (blue: #84b6eb)
- `area: webhook` - GitHub webhook handling (gray: #ededed)
- `area: github` - GitHub integration (blue: #84b6eb)
- `area: argocd` - ArgoCD integration (blue: #84b6eb)
- `area: networking` - Ingress, TLS, DNS management (gray: #ededed)
- `area: cleanup` - Resource cleanup and TTL management (gray: #ededed)
- `area: cost` - Cost optimization (blue: #84b6eb)
- `area: testing` - Test infrastructure and E2E tests (gray: #ededed)
- `area: docs` - Documentation and examples (gray: #ededed)
- `area: observability` - Metrics, logging, tracing (blue: #84b6eb)
- `area: security` - Security related (blue: #84b6eb)
- `area: ai` - AI/ML integration (blue: #84b6eb)

### Size Labels (Required for planning)

Estimate effort for capacity planning:

- `size: XS` - 1-2 hours (light green: #c2e0c6)
- `size: S` - Half day (2-4 hours) (light blue: #bfd4f2)
- `size: M` - 1-2 days (light yellow: #fef2c0)
- `size: L` - 3-5 days (light pink: #f9d0c4)
- `size: XL` - 1-2 weeks, consider breaking down (light purple: #d4c5f9)

### Status Labels (Used by automation)

Track workflow state and enable automation:

- `status: triage` - Needs review and prioritization (gray: #ededed)
- `status: ready` - Ready for development (green: #0e8a16)
- `status: in-progress` - Currently being worked on (blue: #1d76db)
- `status: review` - In code review (yellow: #fbca04)
- `status: blocked` - Blocked by dependency or decision (red: #b60205)

### Workflow Labels (Agent coordination)

Enable SDLC workflow tracking:

- `workflow: bdd-ready` - BDD tests written, ready for implementation (light blue: #c5def5)
- `workflow: ready-for-dev` - BDD complete, developer can start (green: #0e8a16)
- `workflow: ready-for-review` - Code complete, awaiting review (yellow: #fbca04)
- `workflow: needs-qa` - Awaiting QA validation (orange: #d93f0b)
- `workflow: needs-sre` - Awaiting SRE review (purple: #5319e7)

### Phase Labels (Milestone tracking)

Track release phases:

- `phase: v0.1.0` - Core operator (no AI) (teal: #006b75)
- `phase: v0.2.0` - AI integration (green: #0e8a16)
- `phase: v0.3.0` - Production polish (purple: #5319e7)

## Milestones

Milestones represent major releases and align with roadmap phases:

1. **v0.1.0: Core Operator (No AI)** - Due: 2025-12-20
   - Functional operator without AI features
   - GitHub webhook integration
   - ArgoCD deployment
   - Basic environment lifecycle
   - 11 open issues, 1 closed

2. **v0.2.0: AI Integration** - Due: 2026-01-17
   - AI-powered code analysis
   - Synthetic test data generation
   - Cost prediction model
   - Smart test selection
   - 0 open issues

3. **v0.3.0: Production Polish & Launch** - Due: 2026-01-31
   - Security review and hardening
   - Performance testing
   - Comprehensive documentation
   - Public launch preparation
   - 0 open issues

## Issue Workflow

### Creating New Issues

1. **Use issue templates** (bug report, feature request, epic, spike, task)
2. **Add required labels**: priority, type, size
3. **Add optional labels**: area, phase, status, workflow
4. **Add to milestone** if related to release
5. **Add to project board** (automatic for new issues via GitHub Actions)
6. **Set initial workflow stage**: Usually "Backlog" or "Triage"

### Issue Lifecycle

```
Create Issue → Triage → Backlog → Ready → BDD Ready → In Progress →
Code Review → QA Review → SRE Review (if needed) → Done
```

**State transitions**:

- **Triage → Backlog**: After review and prioritization
- **Backlog → Ready**: When dependencies resolved, ready for work
- **Ready → BDD Ready**: After BDD tests written (QA agent)
- **BDD Ready → In Progress**: Developer picks up and starts implementation
- **In Progress → Code Review**: PR created (automatic via GitHub Actions)
- **Code Review → QA Review**: Code review approved (automatic)
- **QA Review → SRE Review**: QA validation passed (if infrastructure changes)
- **SRE Review → Done**: PR merged (automatic via GitHub Actions)

### Working on Issues

1. **Assign yourself** before starting work
2. **Update status** to "In Progress"
3. **Update workflow stage** to match current phase
4. **Comment on progress** if blocked or delayed
5. **Link PR** with "Closes #<number>" in PR description
6. **Wait for automation** to transition states

### Closing Issues

Issues close automatically when PR merges with "Closes #<number>" in description.

**Manual close**:
- When work completed without PR (docs-only, etc.)
- Mark as closed with reason: "completed", "not_planned", or "duplicate"

## Epics and Sub-Issues

### When to Create an Epic

Create an epic when a feature requires:
- More than 1 week of work
- Multiple team members or agents
- Coordination across functional areas
- Phased implementation

### Epic Structure

```
Epic (Parent Issue)
├── Sub-issue: Requirements & BDD
├── Sub-issue: Implementation (Core)
├── Sub-issue: Testing (E2E)
├── Sub-issue: Documentation
└── Sub-issue: Deployment
```

### Creating Epics

1. **Create parent issue** with `type: epic` label
2. **Break into sub-issues** by component or agent
3. **Link sub-issues** using GitHub sub-issues API
4. **Add all to project board**
5. **Set dependencies** if sub-issues have ordering

**Example**:
```bash
# Create epic
gh issue create --title "Add AI code analysis" --label "type: epic" --milestone "v0.2.0"

# Create sub-issues
gh issue create --title "[AI Analysis] Requirements & BDD" --label "type: task"
gh issue create --title "[AI Analysis] Core implementation" --label "type: feature"

# Link sub-issues (requires API call or web UI)
```

### Tracking Epic Progress

- Parent issue shows progress: "3/5 sub-issues complete"
- Update parent description with status
- Comment when sub-issues complete
- Close parent when all sub-issues done

## Automation

### GitHub Actions Workflows

#### project-automation.yml

Automatically manages project board transitions:

**Triggers**:
- Issue/PR opened → Add to project board
- PR opened → Move to "Code Review", add `workflow: ready-for-review` label
- PR approved → Move to "QA Review", add `workflow: needs-qa` label
- PR merged → Close linked issues, move to "Done"
- Issue labeled → Update project board fields

**Configuration**: `.github/workflows/project-automation.yml`

### Manual Project Updates

When automation doesn't work:

```bash
# Add issue to project
gh project item-add 3 --owner mikelane --url https://github.com/mikelane/previewd/issues/123

# Update status field
gh project item-edit --project-id <id> --field-id <field-id> --text "In Progress"

# Check current state
gh project item-list 3 --owner mikelane --format json | jq '.items[] | {number, status}'
```

## Sprint Planning

### Before Sprint

1. **Review backlog** for priority and readiness
2. **Identify blockers** and dependencies
3. **Size estimation** for capacity planning
4. **Create milestone** if new sprint/release

### During Sprint Planning

1. **Set sprint goal** in milestone description
2. **Select issues** based on priority and capacity
3. **Assign to milestone**
4. **Update workflow stages** to "Ready"
5. **Identify risks** and mitigation

### During Sprint

1. **Daily standup** (async via issue comments)
2. **Update workflow stages** as work progresses
3. **Communicate blockers** promptly
4. **Track velocity** (completed story points)

### After Sprint

1. **Close completed issues**
2. **Move incomplete to next sprint**
3. **Review velocity** and adjust
4. **Retrospective** learnings

## Best Practices

### DO

- ✅ Add all issues to project board
- ✅ Use required labels (priority, type, size)
- ✅ Link PRs with "Closes #<number>"
- ✅ Update workflow stages as work progresses
- ✅ Comment on blockers and progress
- ✅ Create epics for large features
- ✅ Use milestones for release planning
- ✅ Follow issue-based development (all work has issue)

### DON'T

- ❌ Skip adding issues to project
- ❌ Create issues without labels
- ❌ Merge PRs without linking issues
- ❌ Leave workflow stages outdated
- ❌ Create work without issues
- ❌ Bypass workflow (direct to main)
- ❌ Ignore blocked dependencies
- ❌ Create epics without sub-issues

## Troubleshooting

### Issue Not in Project Board

**Symptom**: Issue exists but not visible in project

**Fix**:
```bash
gh project item-add 3 --owner mikelane --url https://github.com/mikelane/previewd/issues/<number>
```

### Workflow Stage Not Updating

**Symptom**: Automation didn't transition state

**Causes**:
- Label name mismatch (case-sensitive)
- Automation rule not configured
- GitHub Actions permission issue

**Fix**: Manual update or check `.github/workflows/project-automation.yml`

### PR Not Closing Issue

**Symptom**: PR merged but issue still open

**Causes**:
- Missing "Closes #<number>" in PR description
- Wrong keyword (use: close, closes, closed, fix, fixes, fixed, resolve, resolves, resolved)

**Fix**: Manually close issue and add comment linking to PR

### Milestone Not Showing in Project

**Symptom**: Issues have milestone but not reflected in project Phase field

**Fix**: Project Phase field is separate from milestones. Update both:
```bash
# Add to milestone
gh issue edit <number> --milestone "v0.1.0: Core Operator (No AI)"

# Update project field (requires GraphQL API)
```

## Resources

- [GitHub Projects v2 Documentation](https://docs.github.com/en/issues/planning-and-tracking-with-projects)
- [GitHub Issues Documentation](https://docs.github.com/en/issues)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Project Board](https://github.com/mikelane/previewd/issues)
- [Repository Issues](https://github.com/mikelane/previewd/issues)
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution workflow
- [CLAUDE.md](../CLAUDE.md) - Development workflow requirements

## Contact

For questions or issues with project management:
- **Maintainer**: Mike Lane (@mikelane)
- **Email**: mikelane@gmail.com
- **Project Board**: https://github.com/mikelane/previewd/issues
