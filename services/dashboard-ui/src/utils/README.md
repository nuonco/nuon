# Utility Functions

This directory contains pure utility functions and helper modules that provide common functionality across the application.

## Overview

Utility functions are stateless, reusable pieces of logic that can be imported and used throughout the application. They help maintain consistency and reduce code duplication by centralizing common operations.

## Modules

### Data Processing

- **data-utils.ts** - General data manipulation and transformation helpers
- **string-utils.ts** - String manipulation, formatting, and validation utilities
- **time-utils.ts** - Date and time formatting, parsing, and calculation helpers
- **classnames.ts** - CSS class name management and conditional styling

### Domain-Specific Utilities

- **terraform-utils.ts** - Terraform plan parsing and resource management utilities
- **kubernetes-utils.ts** - Kubernetes resource handling and formatting
- **helm-utils.ts** - Helm chart and deployment utilities
- **workflow-utils.ts** - Workflow state management and processing helpers
- **approval-utils.ts** - Approval process and workflow step utilities
- **install-utils.ts** - Installation management and provisioning helpers

### Infrastructure & API

- **action-utils.ts** - Nuon action response handling and utilities
- **build-query-params.ts** - URL query parameter construction and management
- **code-utils.ts** - Code formatting, syntax highlighting, and parsing
- **status-utils.ts** - Status checking, health monitoring, and state management
- **runner-utils.ts** - Runner and execution environment utilities

### UI & Display

- **timeline-utils.ts** - Timeline visualization and event processing helpers

## Usage

Import specific utilities directly from their files:

```typescript
// Data processing utilities
import { toSentenceCase, formatBytes } from '@/utils/string-utils'
import { formatDuration, timeAgo } from '@/utils/time-utils'
import { classnames, cx } from '@/utils/classnames'

// Domain-specific utilities
import { getResourceChanges, parseTerraformPlan } from '@/utils/terraform-utils'
import { parseKubernetesManifest } from '@/utils/kubernetes-utils'
import { getWorkflowStatus } from '@/utils/workflow-utils'
import { getInstallHealth } from '@/utils/install-utils'

// Infrastructure & API utilities
import { buildQueryParams } from '@/utils/build-query-params'
import { formatCode, highlightSyntax } from '@/utils/code-utils'
import { getStatusColor, isHealthy } from '@/utils/status-utils'
import { getRunnerInfo } from '@/utils/runner-utils'

// UI & Display utilities
import { createTimelineEvents } from '@/utils/timeline-utils'
```

## Guidelines

When adding new utilities:

1. Keep functions pure (no side effects)
2. Make them testable and well-documented
3. Group related functions in the same file
4. Export individual functions rather than default exports
5. Use TypeScript for type safety
6. Consider performance implications for frequently used utilities
