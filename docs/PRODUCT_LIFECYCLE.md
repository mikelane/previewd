# Product Lifecycle Management

This document describes the complete product lifecycle management process for the Previewd project, from idea to production deployment.

## Table of Contents

- [Overview](#overview)
- [Issue Lifecycle](#issue-lifecycle)
- [Pull Request Lifecycle](#pull-request-lifecycle)
- [Release Lifecycle](#release-lifecycle)
- [Workflow Stages](#workflow-stages)
- [Quality Gates](#quality-gates)
- [Metrics and Tracking](#metrics-and-tracking)
- [Automation](#automation)

## Overview

Previewd uses a **GitHub-native, issue-based product lifecycle** that ensures:

- ✅ Complete audit trail of all changes
- ✅ Documentation stays synchronized with code
- ✅ Project management has accurate visibility
- ✅ Future maintainers can understand historical context
- ✅ Quality gates are enforced at every stage

### Core Principles

1. **Issue-First Development**: Every change requires a GitHub issue
2. **PR-Based Integration**: All commits go through Pull Requests
3. **Continuous Documentation**: Docs updated in same PR as code
4. **Test-Driven Development**: Tests written before implementation
5. **Multi-Stage Review**: Code → QA → SRE reviews before merge

## Issue Lifecycle

### Issue States and Transitions

```
New Issue Created
      ↓
   Triage ─────────→ Closed (duplicate/invalid/wontfix)
      ↓
   Ready
      ↓
  BDD Ready (if applicable)
      ↓
 In Progress ←────┐
      ↓           │
 Code Review      │ (feedback loop)
      ↓           │
  QA Review  ─────┘
      ↓
 SRE Review (if applicable)
      ↓
    Done
```

### State Definitions

#### 1. Triage (`status: triage`)

**Description**: Issue has been created but not yet reviewed or prioritized.

**Activities**:
- Product lead reviews issue
- Clarifies requirements if needed
- Assigns priority (P0-P3)
- Assigns size estimate (XS-XL)
- Assigns milestone/phase
- Identifies dependencies
- Links to related issues/epics

**Exit Criteria**:
- Priority assigned
- Size estimated
- Clear acceptance criteria
- No blockers
- Moves to "Ready" or closed

#### 2. Ready (`status: ready`)

**Description**: Issue meets Definition of Ready and can be worked on.

**Definition of Ready**:
- [ ] Clear title and description
- [ ] Acceptance criteria defined (testable)
- [ ] Priority assigned (P0-P3)
- [ ] Size estimated (XS-XL)
- [ ] Milestone assigned (if applicable)
- [ ] Dependencies identified and resolved
- [ ] Technical approach discussed (for complex issues)
- [ ] Assigned to appropriate milestone/phase

**Activities**:
- Waits in backlog for capacity
- May have BDD scenarios written (moves to BDD Ready)
- Developer claims issue when ready to work

**Exit Criteria**:
- Developer assigns themselves
- Moves to "BDD Ready" or "In Progress"

#### 3. BDD Ready (`workflow: bdd-ready`)

**Description**: BDD/Gherkin tests have been written and committed to a feature branch.

**Activities**:
- QA/Product lead writes Gherkin scenarios
- BDD tests committed to feature branch
- Tests are failing (red state)
- Developer ready to make tests pass

**Exit Criteria**:
- Failing BDD tests exist
- Developer starts implementation
- Moves to "In Progress"

#### 4. In Progress (`status: in-progress`)

**Description**: Actively being worked on by a developer.

**Activities**:
- Developer writes unit tests (TDD)
- Developer implements feature
- Developer runs tests locally
- Developer commits to feature branch
- Regular updates on issue (blockers, progress)

**Exit Criteria**:
- Implementation complete
- All tests passing (unit, integration, BDD)
- Code self-reviewed
- Documentation updated
- Pull Request opened
- Moves to "Code Review"

#### 5. Code Review (`workflow: ready-for-review`)

**Description**: PR is open and awaiting code review.

**Activities**:
- Code reviewer examines code quality
- Checks for SOLID principles
- Reviews test coverage
- Suggests improvements
- Approves or requests changes

**Quality Checks**:
- Code follows Go best practices
- Tests have >80% coverage
- No security vulnerabilities
- Documentation updated
- No code smells

**Exit Criteria**:
- Code review approved
- All feedback addressed
- Moves to "QA Review"

#### 6. QA Review (`workflow: needs-qa`)

**Description**: Code approved, awaiting QA validation.

**Activities**:
- QA runs full test suite
- QA performs security audit (OWASP Top 10)
- QA validates acceptance criteria
- QA tests edge cases
- QA documents findings

**Quality Checks**:
- All tests pass
- No security vulnerabilities
- Acceptance criteria met
- Performance acceptable
- Error handling robust

**Exit Criteria**:
- QA approval
- All issues resolved
- Moves to "SRE Review" or "Done"

#### 7. SRE Review (`workflow: needs-sre`)

**Description**: QA passed, awaiting SRE review (for infrastructure changes).

**When Required**:
- Infrastructure as Code changes
- Deployment configuration changes
- Observability/monitoring changes
- Performance-critical changes
- Resource management changes

**Activities**:
- SRE reviews observability (logs/metrics/traces)
- SRE validates infrastructure changes
- SRE checks cost implications
- SRE verifies rollback plan
- SRE approves for production

**Quality Checks**:
- Sufficient observability
- Infrastructure as code validated
- Cost impact understood
- Rollback plan exists
- Production-ready

**Exit Criteria**:
- SRE approval
- All concerns addressed
- Moves to "Done"

#### 8. Done

**Description**: PR merged to main, deployed (or ready to deploy).

**Activities**:
- PR merged to main branch
- CI/CD pipeline runs
- Tests pass in pipeline
- (Optional) Deployed to staging
- (Optional) Deployed to production
- Issue closed automatically

**Verification**:
- Feature works in production
- Monitoring shows healthy metrics
- No rollbacks needed

## Pull Request Lifecycle

### PR States and Transitions

```
Branch Created
      ↓
   Draft PR (optional)
      ↓
   Ready for Review
      ↓
 Automated Checks ────→ Failed (fix and retry)
      ↓
   Code Review ←───────┐
      ↓                │ (feedback loop)
  Changes Requested ───┘
      ↓
   Approved
      ↓
   QA Review ←─────────┐
      ↓                │ (feedback loop)
  Changes Requested ───┘
      ↓
  SRE Review (if applicable)
      ↓
   Merged
      ↓
  Deployed
```

### PR Requirements

#### Before Opening PR
- [ ] Feature branch created from main
- [ ] Tests written first (TDD)
- [ ] Implementation complete
- [ ] All tests passing locally
- [ ] Documentation updated
- [ ] Self-review performed

#### PR Description Must Include
- [ ] Summary of changes
- [ ] Why this change is needed
- [ ] Testing performed
- [ ] Reference to issue: `Closes #N`
- [ ] Breaking changes (if any)
- [ ] Screenshots/demos (if applicable)

#### Automated Checks
- [ ] Tests pass (`make test`)
- [ ] Linting passes (`golangci-lint`)
- [ ] Code formatted (`gofmt`)
- [ ] Coverage threshold met (>80%)
- [ ] No security vulnerabilities
- [ ] Build succeeds

#### Code Review Approval
- [ ] Code quality reviewed
- [ ] Design patterns validated
- [ ] Test coverage adequate
- [ ] Documentation sufficient
- [ ] All comments addressed

#### QA Approval
- [ ] Acceptance criteria met
- [ ] Security audit passed
- [ ] Performance acceptable
- [ ] Edge cases tested

#### SRE Approval (if applicable)
- [ ] Observability sufficient
- [ ] Infrastructure validated
- [ ] Cost impact acceptable
- [ ] Rollback plan exists

### Merge Process

1. **All approvals obtained** (Code + QA + SRE if needed)
2. **Automated checks passing**
3. **No merge conflicts**
4. **Maintainer merges PR** (squash and merge)
5. **Feature branch deleted**
6. **CI/CD pipeline runs**
7. **Issue automatically closed**

## Release Lifecycle

### Release Cadence

- **v0.1.0**: 6 weeks (Core operator, no AI)
- **v0.2.0**: 4 weeks (AI integration)
- **v0.3.0**: 2 weeks (Production polish)
- **Post-launch**: 2-week sprints with monthly releases

### Release Process

#### 1. Planning (Week before release)

- [ ] Review milestone progress
- [ ] Identify issues at risk
- [ ] Determine what's in/out of release
- [ ] Create release notes draft
- [ ] Plan release testing

#### 2. Code Freeze (3 days before release)

- [ ] No new features merged
- [ ] Only bug fixes allowed
- [ ] Regression testing performed
- [ ] Documentation finalized
- [ ] CHANGELOG updated

#### 3. Release Candidate (2 days before release)

- [ ] Create release branch: `release/v0.X.0`
- [ ] Tag release candidate: `v0.X.0-rc1`
- [ ] Deploy to staging environment
- [ ] Smoke testing
- [ ] Performance testing
- [ ] Security scan

#### 4. Release (Release day)

- [ ] Final testing complete
- [ ] Tag release: `v0.X.0`
- [ ] Build and publish artifacts
- [ ] Update documentation site
- [ ] Publish release notes
- [ ] Announce release
- [ ] Monitor for issues

#### 5. Post-Release (Week after release)

- [ ] Monitor metrics
- [ ] Address critical issues (P0/P1)
- [ ] Gather user feedback
- [ ] Update roadmap
- [ ] Retrospective meeting

## Workflow Stages

### Visual Workflow

```
┌─────────────┐
│   Backlog   │  Ideas, future work
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Triage    │  Review, prioritize, estimate
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    Ready    │  Definition of Ready met
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  BDD Ready  │  BDD tests written (failing)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ In Progress │  Developer implementing
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Code Review │  Reviewer checking quality
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  QA Review  │  QA validating functionality
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ SRE Review  │  SRE checking infrastructure
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    Done     │  Merged and deployed
└─────────────┘
```

### Stage Owners

| Stage | Owner | Responsibilities |
|-------|-------|------------------|
| Triage | Product Lead | Prioritize, estimate, clarify requirements |
| BDD Ready | QA Engineer | Write Gherkin scenarios, create failing tests |
| In Progress | Developer | Implement with TDD, update docs |
| Code Review | Code Reviewer | Review quality, design, best practices |
| QA Review | QA Engineer | Validate functionality, security, performance |
| SRE Review | SRE | Review observability, infrastructure, cost |
| Merge | Maintainer | Final approval, merge, monitor deployment |

## Quality Gates

### Definition of Ready (Issues)

Before an issue can move from "Triage" to "Ready":

- [ ] Clear, testable acceptance criteria
- [ ] Priority assigned (P0-P3)
- [ ] Size estimated (XS-XL)
- [ ] Milestone/phase assigned
- [ ] Dependencies identified and resolved
- [ ] Technical approach discussed
- [ ] Labels applied (type, area, priority, size)

### Definition of Done (Pull Requests)

Before a PR can be merged to main:

#### Code Quality
- [ ] All tests pass (unit, integration, E2E)
- [ ] Test coverage >80% for new code
- [ ] Code follows Go best practices
- [ ] No security vulnerabilities
- [ ] SOLID principles followed
- [ ] No code smells

#### Documentation
- [ ] Godoc comments updated
- [ ] README updated (if user-facing)
- [ ] API docs updated (if API changes)
- [ ] CHANGELOG updated (unreleased section)
- [ ] Examples updated (if applicable)

#### Review
- [ ] Code review approved
- [ ] QA review approved
- [ ] SRE review approved (if applicable)
- [ ] All feedback addressed

#### Process
- [ ] Linked to issue (Closes #N)
- [ ] Milestone assigned
- [ ] Labels applied
- [ ] CI/CD passes
- [ ] No merge conflicts

## Metrics and Tracking

### Issue Metrics

| Metric | Target | Tracking |
|--------|--------|----------|
| Time in Triage | <3 days | GitHub Project automation |
| Time to Ready | <5 days | Manual tracking |
| Cycle Time (Ready → Done) | <1 week (avg) | GitHub Insights |
| Issues Completed per Sprint | 5-10 | Milestone burndown |

### Pull Request Metrics

| Metric | Target | Tracking |
|--------|--------|----------|
| Time to First Review | <24 hours | GitHub PR insights |
| Time to Merge | <3 days | GitHub PR insights |
| Review Iterations | <3 (avg) | Manual tracking |
| PR Size | <500 lines | Automated labeling |

### Release Metrics

| Metric | Target | Tracking |
|--------|--------|----------|
| Release Frequency | Every 2 weeks (post v1.0) | Manual |
| Defect Rate | <5% of issues | Issue labels |
| Velocity | Increasing trend | Sprint reports |
| Coverage | >90% (v1.0+) | CI reports |

### Success Metrics (Project Goals)

| Metric | Target | Timeline |
|--------|--------|----------|
| GitHub Stars | 500 | 6 months post-launch |
| Production Users | 10 companies | 6 months post-launch |
| Contributors | 5+ external | 6 months post-launch |
| Test Coverage | >90% | v1.0.0 |
| Performance | <2 min to ready state | v0.1.0 |

## Automation

### GitHub Actions Workflows

#### 1. Auto-Label PRs
**Trigger**: PR opened or edited
**Actions**:
- Label based on file paths (area: ai, area: kubernetes)
- Label based on size (<100 lines = small)
- Label based on PR title (feat → enhancement)

#### 2. Auto-Move Issues on Project Board
**Trigger**: Issue/PR state change
**Actions**:
- PR opened → Move issue to "Code Review"
- PR approved → Move to "QA Review"
- PR merged → Move to "Done", close issue

#### 3. Auto-Assign Reviewers
**Trigger**: PR opened
**Actions**:
- Assign code reviewer based on CODEOWNERS
- Request QA review when code approved
- Request SRE review when QA approved (if needed)

#### 4. Auto-Close Stale Issues
**Trigger**: Scheduled (weekly)
**Actions**:
- Label issues inactive >30 days as "stale"
- Close stale issues after 7 more days (if no activity)
- Skip issues with labels: P0, P1, type: epic

#### 5. Auto-Update Project Fields
**Trigger**: Issue/PR labeled
**Actions**:
- Sync Priority label → Priority field
- Sync Size label → Size field
- Sync Phase label → Phase field
- Sync Workflow label → Workflow Stage field

#### 6. Auto-Link Issues to PRs
**Trigger**: PR opened with "Closes #N" in description
**Actions**:
- Link issue to PR
- Add PR link to issue
- Move issue to "Code Review" stage

### Manual Processes (Until Automated)

Some processes require manual intervention (for now):

- **Milestone assignment**: Manually assign during triage
- **Epic breakdown**: Manually create child issues from epics
- **Sprint planning**: Manually select issues for sprint
- **Retrospectives**: Manually run after each sprint
- **Release notes**: Manually compile from CHANGELOG

## References

- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines
- [CLAUDE.md](../CLAUDE.md) - Project context and workflow
- [GitHub Issues](https://github.com/mikelane/previewd/issues) - Issue tracking
- [Milestones](https://github.com/mikelane/previewd/milestones) - Release planning

---

**Questions?** Open an issue or contact [@mikelane](https://github.com/mikelane).
