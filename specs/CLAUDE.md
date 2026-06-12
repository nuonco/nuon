# Specs Directory

This directory contains design specifications for Nuon platform features. Each spec lives in a numbered subdirectory.

## For AI Assistants

When asked to create a spec:

1. Create the next numbered directory (`NNN-kebab-case-name`).
2. Write a `prompt.md` with the problem statement, requirements, design decisions, and key files.
3. Generate a `spec.html` — an interactive review document with:
   - Architecture diagram (Mermaid, rendered client-side)
   - Code listings with diff highlighting and line numbers
   - Metrics and event specifications (tables)
   - Inline commenting (click `+` on any line to comment; comments download as `.md` files)
4. Review comments saved into the spec directory can be read back for follow-up implementation.

When referencing a spec, use its number and name (e.g., "spec 001 runner-health-check").

## Existing Specs

| # | Name | Status |
|---|------|--------|
| 001 | runner-health-check | Implemented |
