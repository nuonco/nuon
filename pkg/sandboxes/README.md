# Sandboxes

Common parsing logic / types for sandboxes and their outputs.

## Background

We export outputs from our [sandboxes](../../sandboxes) and need to use _some_ values in code. For instance, to look up an ECR repo, deploy a runner etc. We tried to avoid managing types for these, but ultimately it led to too many places where output logic was leaking / dependencies were not known.

This package enforces that we model the outputs from our sandboxes as actual go types, so we can safely rely on them throughout our code.
