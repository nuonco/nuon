---
name: DRY
description: Flags common optimisations for Do Not Repeat Yourself
severity-default: medium
tools: [Grep, Read]
---

Look for these patterns:

- Look if any of the processing/utility functions that were generated already exists. 

Report the line, why it matters, and suggest a refactor based on existing references on where we could place it.