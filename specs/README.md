# Specs

Design specifications for features and systems in the Nuon platform. Each spec is a numbered directory containing an
interactive HTML review document and supporting materials.

## Structure

```
specs/
  README.md              # This file
  CLAUDE.md              # Instructions for AI assistants
  001-runner-health-check/
    spec.html            # Interactive review document (open in browser)
    prompt.md            # High-level description and requirements
  002-next-feature/
    ...
```

## Naming Convention

Specs are numbered sequentially: `NNN-kebab-case-name`. The number provides ordering; the name provides context.

## Workflow

1. A spec starts as a `prompt.md` capturing the problem, requirements, and design decisions.
2. An interactive `spec.html` is generated for visual review — includes architecture diagrams, code listings, metrics
   tables, and inline commenting.
3. Comments from the HTML review are saved as markdown files in the spec directory for follow-up.
4. Once approved, the spec serves as a reference for the implementation.
