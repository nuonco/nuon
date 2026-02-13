---
name: Readability
description: Flags opportunities to simplify control flow and improve code readability
severity-default: low
tools: [Grep, Read]
---

Look for these patterns:

- Deeply nested `if/else` blocks or long guard chains that obscure the main flow.
- Small, repeated blocks of logic that could be extracted into well-named helper functions.
- Mixed concerns in a single function, such as side-effect logging intertwined with data shaping.

Report the line, why it matters, and suggest a refactor that makes the primary flow easier to read (e.g., helper extraction, early returns).
