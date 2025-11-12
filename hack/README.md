# Development Scripts

This directory contains utility scripts for local development and CI/CD pipelines.

## fix-sarif.sh

Post-processes gosec SARIF output to fix GitHub CodeQL compatibility issues.

### Problem

Gosec v2.22+ generates SARIF output with a `fixes` field containing `artifactChanges: null`. However, the SARIF 2.1.0 schema requires `artifactChanges` to be an array, not null. This causes GitHub's CodeQL action to reject the SARIF file with the error:

```
instance.runs[0].results[0].fixes[0].artifactChanges is not of a type(s) array
```

### Solution

This script removes the `fixes` field entirely from the SARIF output, which is acceptable since:
1. The `fixes` field is optional in SARIF
2. gosec's fixes are informational suggestions, not required for security scanning
3. The security findings themselves remain intact

### Usage

```bash
./hack/fix-sarif.sh input.sarif output.sarif
```

### Example

```bash
# Run gosec
gosec -no-fail -fmt sarif -out gosec-raw.sarif -conf .gosec.json ./...

# Fix SARIF format
./hack/fix-sarif.sh gosec-raw.sarif gosec.sarif

# Upload to GitHub (in CI)
# github/codeql-action/upload-sarif@v3 will now accept this file
```

### Dependencies

- `jq` - JSON processor (pre-installed on GitHub Actions Ubuntu runners)

### Technical Details

The script uses `jq` to remove the `fixes` field from all results in the SARIF file:

```bash
jq 'del(.runs[].results[]?.fixes)' "$INPUT_FILE" > "$OUTPUT_FILE"
```

This preserves all other SARIF content including:
- Security findings (results)
- Rule definitions
- Taxonomies (CWE mappings)
- Tool metadata

### Related Issues

- GitHub Issue #22: Configure gosec security scanner and SARIF upload
- gosec Issue: https://github.com/securego/gosec/issues/1037 (SARIF format compatibility)
