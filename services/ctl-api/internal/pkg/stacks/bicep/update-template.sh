#!/bin/bash

# Script to convert bicep template to JSON and update Go template variable
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BICEP_FILE="$SCRIPT_DIR/stack.bicep"
OUTPUT_JSON="$SCRIPT_DIR/output.json"
TMPL_GO_FILE="$SCRIPT_DIR/tmpl.go"

echo "Converting bicep template to JSON..."

# Convert bicep to JSON using Azure CLI
if ! command -v az &> /dev/null; then
    echo "ERROR: Azure CLI is not installed. Please install it first."
    echo "Visit: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"
    exit 1
fi

# Convert bicep to JSON
az bicep build --file "$BICEP_FILE" --outfile "$OUTPUT_JSON"

if [ ! -f "$OUTPUT_JSON" ]; then
    echo "ERROR: Failed to generate JSON output"
    exit 1
fi

echo "Generated JSON template: $OUTPUT_JSON"

# Read the JSON content and escape it for Go string
JSON_CONTENT=$(cat "$OUTPUT_JSON")

# Escape backticks and backslashes for Go raw string literal
ESCAPED_JSON=$(echo "$JSON_CONTENT" | sed 's/`/` + "`" + `/g')

# Create the new tmpl.go file content
cat > "$TMPL_GO_FILE" << 'EOF'
package bicep

// Generated file. DO NOT EDIT
const tmpl = `
EOF

echo "$ESCAPED_JSON" >> "$TMPL_GO_FILE"

cat >> "$TMPL_GO_FILE" << 'EOF'
`
EOF

rm $OUTPUT_JSON

echo "Updated Go template file: $TMPL_GO_FILE"
echo "Done! The bicep template has been converted to JSON and embedded in the Go string variable."
