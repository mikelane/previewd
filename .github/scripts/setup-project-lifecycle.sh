#!/usr/bin/env bash
#
# Setup script for Previewd product lifecycle management
#
# This script configures the complete product lifecycle management system:
# - Labels (priority, size, type, status, workflow, phase, area)
# - Milestones (v0.1.0, v0.2.0, v0.3.0)
# - GitHub Project board with custom fields
# - Repository linkage
#
# Usage: ./.github/scripts/setup-project-lifecycle.sh
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
OWNER="mikelane"
REPO="previewd"
PROJECT_NUMBER=3

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ ${NC}$1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

create_label() {
    local name="$1"
    local description="$2"
    local color="$3"

    if gh label create "$name" --description "$description" --color "$color" 2>/dev/null; then
        log_success "Created label: $name"
    else
        log_warning "Label already exists: $name"
    fi
}

# Main setup
main() {
    log_info "Starting Previewd product lifecycle setup..."
    echo ""

    # Verify gh CLI is installed
    if ! command -v gh &> /dev/null; then
        log_error "GitHub CLI (gh) is not installed. Install from: https://cli.github.com/"
        exit 1
    fi

    # Verify authentication
    if ! gh auth status &> /dev/null; then
        log_error "Not authenticated with GitHub. Run: gh auth login"
        exit 1
    fi

    log_info "GitHub CLI authenticated ✓"
    echo ""

    # ====================
    # Create Labels
    # ====================
    log_info "Creating labels..."

    # Priority labels
    create_label "priority: P0" "Critical: Production down, security vulnerability" "b60205"
    create_label "priority: P1" "High: Blocking feature, severe bug" "d93f0b"
    create_label "priority: P2" "Medium: Important but not blocking" "fbca04"
    create_label "priority: P3" "Low: Nice to have, minor issue" "0e8a16"

    # Size labels
    create_label "size: XS" "1-2 hours" "c2e0c6"
    create_label "size: S" "Half day (2-4 hours)" "bfd4f2"
    create_label "size: M" "1-2 days" "fef2c0"
    create_label "size: L" "3-5 days" "f9d0c4"
    create_label "size: XL" "1-2 weeks (consider breaking down)" "d4c5f9"

    # Type labels
    create_label "type: epic" "Large initiative spanning multiple issues" "3e4b9e"
    create_label "type: story" "User story with business value" "5319e7"
    create_label "type: task" "Technical task without direct user value" "1d76db"
    create_label "type: spike" "Research or investigation work" "d876e3"

    # Status labels
    create_label "status: triage" "Needs review and prioritization" "ededed"
    create_label "status: ready" "Ready for development" "0e8a16"
    create_label "status: in-progress" "Currently being worked on" "1d76db"
    create_label "status: blocked" "Blocked by dependency or decision" "b60205"
    create_label "status: review" "In code review" "fbca04"

    # Workflow labels
    create_label "workflow: bdd-ready" "BDD tests written, ready for implementation" "c5def5"
    create_label "workflow: ready-for-dev" "BDD complete, developer can start" "0e8a16"
    create_label "workflow: ready-for-review" "Code complete, awaiting review" "fbca04"
    create_label "workflow: needs-qa" "Awaiting QA validation" "d93f0b"
    create_label "workflow: needs-sre" "Awaiting SRE review" "5319e7"

    # Phase labels
    create_label "phase: v0.1.0" "Core operator (no AI)" "006b75"
    create_label "phase: v0.2.0" "AI integration" "0e8a16"
    create_label "phase: v0.3.0" "Production polish" "5319e7"

    # Area labels
    create_label "area: ai" "AI/ML integration" "84b6eb"
    create_label "area: argocd" "ArgoCD integration" "84b6eb"
    create_label "area: cost" "Cost optimization" "84b6eb"
    create_label "area: github" "GitHub integration" "84b6eb"
    create_label "area: kubernetes" "Kubernetes core" "84b6eb"
    create_label "area: observability" "Metrics, logging, tracing" "84b6eb"
    create_label "area: security" "Security related" "84b6eb"

    echo ""
    log_success "Labels created successfully"
    echo ""

    # ====================
    # Create Milestones
    # ====================
    log_info "Creating milestones..."

    # Note: gh CLI doesn't have milestone commands, using API
    if gh api repos/${OWNER}/${REPO}/milestones \
        -f title="v0.1.0: Core Operator (No AI)" \
        -f description="Functional Kubernetes operator with GitHub webhook integration, ArgoCD deployment, and basic environment lifecycle management. No AI features." \
        -f state="open" \
        -f due_on="2025-12-21T00:00:00Z" \
        &>/dev/null; then
        log_success "Created milestone: v0.1.0"
    else
        log_warning "Milestone already exists or failed: v0.1.0"
    fi

    if gh api repos/${OWNER}/${REPO}/milestones \
        -f title="v0.2.0: AI Integration" \
        -f description="AI-powered features for intelligent service dependency detection, synthetic test data generation, and cost optimization." \
        -f state="open" \
        -f due_on="2026-01-18T00:00:00Z" \
        &>/dev/null; then
        log_success "Created milestone: v0.2.0"
    else
        log_warning "Milestone already exists or failed: v0.2.0"
    fi

    if gh api repos/${OWNER}/${REPO}/milestones \
        -f title="v0.3.0: Production Polish & Launch" \
        -f description="Production-ready operator with comprehensive documentation, security hardening, performance optimization, and public launch." \
        -f state="open" \
        -f due_on="2026-02-01T00:00:00Z" \
        &>/dev/null; then
        log_success "Created milestone: v0.3.0"
    else
        log_warning "Milestone already exists or failed: v0.3.0"
    fi

    echo ""
    log_success "Milestones created successfully"
    echo ""

    # ====================
    # GitHub Project Board
    # ====================
    log_info "Checking GitHub Project board..."

    # Check if project exists
    if gh project view ${PROJECT_NUMBER} --owner ${OWNER} &>/dev/null; then
        log_success "Project #${PROJECT_NUMBER} exists"

        # Link repository to project
        if gh project link ${PROJECT_NUMBER} --owner ${OWNER} --repo ${OWNER}/${REPO} 2>/dev/null; then
            log_success "Repository linked to project"
        else
            log_warning "Repository already linked or failed to link"
        fi
    else
        log_warning "Project #${PROJECT_NUMBER} not found. Please create it manually at:"
        log_warning "https://github.com/users/${OWNER}/projects"
    fi

    echo ""

    # ====================
    # Summary
    # ====================
    log_success "Product lifecycle setup complete!"
    echo ""
    log_info "Next steps:"
    echo "  1. Review labels: gh label list"
    echo "  2. Review milestones: gh api repos/${OWNER}/${REPO}/milestones"
    echo "  3. View project board: https://github.com/users/${OWNER}/projects/${PROJECT_NUMBER}"
    echo "  4. Create your first issue: gh issue create"
    echo ""
    log_info "Documentation:"
    echo "  - CONTRIBUTING.md - Contribution guidelines"
    echo "  - docs/PRODUCT_LIFECYCLE.md - Complete lifecycle documentation"
    echo "  - GitHub Project: https://github.com/users/${OWNER}/projects/${PROJECT_NUMBER}"
    echo ""
}

# Run main function
main "$@"
