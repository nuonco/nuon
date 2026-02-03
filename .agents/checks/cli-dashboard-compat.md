---
name: CLI-Dashboard-compat
description: Flags if a functionality is added to dashboard/CLI and missing in the other
severity-default: low
tools: [Grep, Read]
---

Look for these patterns:

- Look if any of the functionality that's been added to Dashboard/CLI and could potentially be added to the other.
- If there are any plans generated in markdown files,etc - check if it has accounted for both dashboard & CLI.  

Report the functionality, why it could matter and how it could be added. 