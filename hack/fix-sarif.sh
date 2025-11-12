#!/bin/bash
# fix-sarif.sh - Post-process gosec SARIF output to fix GitHub compatibility issues
#
# This script fixes the SARIF format issues that prevent GitHub CodeQL from accepting
# gosec output. Specifically, it removes the 'fixes' field which contains 'artifactChanges: null'
# that violates the SARIF 2.1.0 schema (which expects an array).
#
# Usage: ./hack/fix-sarif.sh input.sarif output.sarif

set -euo pipefail

if [ $# -ne 2 ]; then
    echo "Usage: $0 <input-sarif> <output-sarif>"
    exit 1
fi

INPUT_FILE="$1"
OUTPUT_FILE="$2"

if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: Input file '$INPUT_FILE' not found"
    exit 1
fi

# Use jq to remove the 'fixes' field from all results
# The 'fixes' field contains 'artifactChanges: null' which violates SARIF schema
# GitHub expects 'artifactChanges' to be an array, not null
jq 'del(.runs[].results[]?.fixes)' "$INPUT_FILE" > "$OUTPUT_FILE"

echo "Fixed SARIF file written to: $OUTPUT_FILE"
