Export policy validation reports in various formats.

This endpoint retrieves the policy evaluation results for a specific policy validation and returns them in the requested format.

## Supported Formats

- **opa** (default): Raw OPA JSON evaluation results including violations, counts, and evaluation metadata
- **sarif**: SARIF 2.1.0 format for integration with code scanning tools and CI/CD systems  
- **html**: Human-readable HTML report suitable for viewing in browsers or sharing

## Usage Examples

```bash
# Get OPA JSON report (default)
curl -H "Authorization: Bearer $TOKEN" \
     -H "X-Nuon-Org-ID: $ORG_ID" \
     "https://api.nuon.co/v1/policy-validations/{validation_id}/reports"

# Get SARIF format for CI integration
curl -H "Authorization: Bearer $TOKEN" \
     -H "X-Nuon-Org-ID: $ORG_ID" \
     "https://api.nuon.co/v1/policy-validations/{validation_id}/reports?format=sarif"

# Get HTML report for human review
curl -H "Authorization: Bearer $TOKEN" \
     -H "X-Nuon-Org-ID: $ORG_ID" \
     "https://api.nuon.co/v1/policy-validations/{validation_id}/reports?format=html"
```

## Response Headers

All formats include a `Content-Disposition` header for file download:
- `policy-report-{validation_id}.opa.json`
- `policy-report-{validation_id}.sarif.json`
- `policy-report-{validation_id}.html`
