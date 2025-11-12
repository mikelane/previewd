# Known Vulnerabilities

This document tracks known security vulnerabilities in Previewd's dependencies that have been evaluated and accepted with documented risk mitigation.

## Status: None

As of 2025-11-11, there are no known vulnerabilities with CRITICAL or HIGH severity in Previewd's dependencies.

## How to Use This Document

When a vulnerability is discovered that cannot be immediately patched:

1. Create an entry below using the template
2. Document the risk assessment and mitigation plan
3. Set a review date (maximum 30 days for MEDIUM, 90 days for LOW)
4. Update this document when the vulnerability is resolved

## Template

```markdown
## CVE-YYYY-XXXXX: [Vulnerability Title]

**Severity:** [CRITICAL/HIGH/MEDIUM/LOW]
**Package:** [package-name] [version]
**Status:** [Accepted/In Progress/Resolved]
**CVSS Score:** [X.X]

**Description:**
[Brief description of the vulnerability]

**Affected Versions:**
- [package-name] < [patched-version]

**Impact Assessment:**
[Detailed explanation of why this vulnerability affects or doesn't affect Previewd]
- Is the vulnerable code path used?
- What is the attack surface?
- What is the potential impact?

**Mitigation:**
[Actions taken to reduce risk]
- Workarounds implemented
- Configuration changes
- Monitoring added
- Upgrade plan

**Justification for Acceptance:**
[Why we're accepting this risk]
- No patch available yet
- Vulnerable code path not used
- Impact is minimal in our context
- Alternative dependency has worse issues

**Timeline:**
- **Discovered:** [YYYY-MM-DD]
- **Evaluated:** [YYYY-MM-DD]
- **Accepted By:** [Name/Role]
- **Review Date:** [YYYY-MM-DD]
- **Target Resolution:** [YYYY-MM-DD or "When patch available"]

**References:**
- [CVE Link]
- [Vendor Advisory]
- [GitHub Issue/PR if applicable]
```

## Review Process

All accepted vulnerabilities must be reviewed on their review date:

1. Check if a patch is now available
2. Re-evaluate risk based on current context
3. Update mitigation if needed
4. Extend review date if still unpatched (max 30 days)
5. Escalate to HIGH priority if patch remains unavailable after 90 days

## Historical Vulnerabilities

### [Example - Remove this section once you have real entries]

#### CVE-2024-12345: Example SQL Injection

**Severity:** MEDIUM
**Package:** github.com/example/database v1.0.0
**Status:** Resolved (2025-11-05)

**Description:**
SQL injection vulnerability in query builder.

**Resolution:**
Upgraded to v1.1.0 which includes the patch.

**Resolved By:** Mike Lane
**Resolved Date:** 2025-11-05

---

*Last Updated: 2025-11-11*
*Next Review: 2025-12-11*
