---
name: Unit-Testable Functions
description: Flags functions without network calls that should have unit tests or be refactored to be testable
severity-default: medium
tools: [Grep, Read]
---

Look for these patterns:

- Functions that perform pure transformations, parsing, or validation without network calls or side effects, but lack unit tests.
- Functions that are hard to test due to mixed responsibilities, where extracting a helper would enable unit tests.

Report the line, why it matters, and suggest either adding a unit test or extracting a smaller helper to make testing possible.
