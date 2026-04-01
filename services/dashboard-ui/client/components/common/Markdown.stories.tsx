import type { ReactNode } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { OrgContext } from '@/providers/org-provider'
import { InstallContext } from '@/providers/install-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import type { TOrg, TInstall } from '@/types'
import { Markdown } from './Markdown'

const mockOrg = { id: 'org-mock', name: 'Mock Org' } as TOrg
const mockInstall = { id: 'install-mock', name: 'Mock Install' } as TInstall
const storyQueryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
})

const MockInstallProviders = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={storyQueryClient}>
    <OrgContext.Provider value={{ org: mockOrg, refresh: () => {} }}>
      <InstallContext.Provider value={{ install: mockInstall, refresh: () => {} }}>
        <SurfacesProvider>
          {children}
        </SurfacesProvider>
      </InstallContext.Provider>
    </OrgContext.Provider>
  </QueryClientProvider>
)

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Enhanced Markdown Component</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The enhanced Markdown component uses react-markdown with custom
        component overrides to provide CodeBlock integration, JSONViewer
        support, and design system consistency.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Basic Features</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`# Enhanced Markdown
          
This component uses react-markdown with custom component overrides.

## Features

- **CodeBlock integration** for syntax highlighting
- **JSONViewer integration** for interactive JSON
- **Mermaid diagrams** with custom rendering
- **External link handling** (opens in new tabs)
- **Table wrapping** for responsive design
- **Collapsible content** support

<details>
<summary>Click to test enhanced collapsible content</summary>

This collapsible content now uses Expand component styling with:

- **Enhanced styling** - Matches the design system
- **Smooth animations** - Rotate arrow icon on expand/collapse  
- **Proper spacing** - Consistent padding and borders
- **Hover effects** - Interactive states like Expand component
- **Accessibility** - Screen reader friendly

\`\`\`javascript
// Code blocks work inside details too
function expandExample() {
  console.log('Details content with code!')
}
\`\`\`

</details>

Check out [this external link](https://example.com) and [this internal link](/dashboard) using the custom Link component.`}
        />
      </div>
    </div>
  </div>
)

export const LinkIntegration = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">
        Custom Link Component Integration
      </h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        All links within markdown now use the custom Link component, providing
        consistent styling and behavior throughout the application.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Link Examples</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`## Link Types

### External Links
These links automatically open in new tabs:
- [GitHub](https://github.com) - External repository
- [Documentation](https://docs.example.com) - External docs
- [API Reference](https://api.example.com) - External API

### Internal Links  
These use Next.js routing:
- [Dashboard](/dashboard) - Internal dashboard
- [Settings](/settings) - Application settings
- [Profile](/profile) - User profile

### Anchor Links
These scroll to sections:
- [Go to top](#top) - Anchor link
- [Features section](#features) - Another anchor

All links use the custom Link component with proper styling, hover states, and focus management.`}
        />
      </div>
    </div>
  </div>
)

export const CodeBlockIntegration = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Enhanced Code Integration</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Both inline code and code blocks now use the design system components.
        Inline code uses the Code component with inline variant, while code
        blocks use CodeBlock for syntax highlighting and JSONViewer for
        interactive JSON.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Inline Code Examples</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`## Inline Code Usage

Use inline code for:

- **Variables**: The \`userId\` parameter is required for authentication
- **Functions**: Call \`fetchUserData()\` to retrieve user information  
- **File paths**: Update the \`src/components/App.tsx\` file
- **Commands**: Run \`npm install\` to install dependencies
- **API endpoints**: Make a GET request to \`/api/v1/users\`
- **CSS classes**: Add the \`bg-primary\` class for styling
- **Environment variables**: Set \`NODE_ENV=production\` for the build

You can also use inline code within **bold text like \`this\`** or *italic text like \`this\`*.

Mix inline code with links: Check the [\`useState\`](https://react.dev) hook documentation.`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">
        JavaScript with Syntax Highlighting
      </h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`## React Component Example

\`\`\`javascript
import React, { useState, useEffect } from 'react'

function UserDashboard({ userId }) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)
  
  useEffect(() => {
    fetchUserData(userId).then(data => {
      setUser(data)
      setLoading(false)
    })
  }, [userId])
  
  if (loading) return <div>Loading user...</div>
  
  return (
    <div className="dashboard">
      <h1>Welcome, {user.name}!</h1>
      <p>Email: {user.email}</p>
    </div>
  )
}
\`\`\`

This code block uses the CodeBlock component for proper syntax highlighting.`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Directory tree</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`## App directory structure

\`\`\`
.
├── runner.toml                    # Runner config (AWS)
├── stack.toml                     # CloudFormation stack
├── sandbox.toml                   # EKS sandbox with operation_roles
├── sandbox.tfvars                 # Cluster vars + custom role access entries
├── metadata.toml                  # App metadata
├── inputs.toml                    # User-facing inputs (domain)
├── operation_roles.toml           # Matrix rules (app-wide fallback)
│
├── components/
│   ├── whoami.toml                # Helm chart with deploy/teardown roles
│   └── certificate.toml           # Terraform module with deploy/teardown roles
│
├── actions/
│   ├── deployment_status.toml     # Read-only action (view role)
│   └── deployment_restart.toml    # Write action (edit role)
│
├── permissions/
│   ├── provision.toml             # Default provision role
│   ├── maintenance.toml           # Default maintenance role
│   ├── deprovision.toml           # Default deprovision role
│   ├── sandbox-provision.toml     # Custom: sandbox provision
│   ├── sandbox-maintenance.toml   # Custom: sandbox reprovision
│   ├── sandbox-deprovision.toml   # Custom: sandbox deprovision
│   ├── whoami-deploy.toml         # Custom: whoami deploy
│   ├── whoami-teardown.toml       # Custom: whoami teardown
│   ├── certificate-deploy.toml    # Custom: certificate deploy
│   ├── certificate-teardown.toml  # Custom: certificate teardown
│   ├── deployments-status-trigger.toml
│   ├── deployment-restart-trigger.toml
│   ├── provision_boundary.json    # Boundary for provision/deploy ops
│   ├── deprovision_boundary.json  # Boundary for teardown/deprovision ops
│   └── maintenance_boundary.json  # Boundary for action triggers
│
└── src/components/                # Source code for components
    ├── certificate/               # Terraform (ACM + Route53)
    └── whoami/                    # Helm chart (deployment + service)
\`\`\`
`}
        />
      </div>
    </div>
  </div>
)

export const JSONViewerIntegration = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Interactive JSON Viewing</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        JSON code blocks automatically render with the JSONViewer component,
        providing expandable/collapsible tree view and better readability.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Configuration JSON</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`## Application Configuration

\`\`\`json
{
  "app": {
    "name": "dashboard-ui",
    "version": "2.1.0",
    "environment": "production"
  },
  "api": {
    "baseUrl": "https://api.example.com",
    "timeout": 5000,
    "retries": 3,
    "endpoints": {
      "users": "/v1/users",
      "auth": "/v1/auth",
      "analytics": "/v1/analytics"
    }
  },
  "features": {
    "analytics": true,
    "darkMode": true,
    "notifications": false,
    "multiTenant": true
  },
  "theme": {
    "primary": "#8040BF",
    "secondary": "#61AFEF",
    "success": "#98C379",
    "warning": "#D19A66",
    "error": "#E06C75"
  }
}
\`\`\`

The JSON above is rendered with the interactive JSONViewer component.`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">API Response JSON</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`## User API Response

\`\`\`json
{
  "success": true,
  "data": {
    "user": {
      "id": "user_12345",
      "name": "Alice Johnson",
      "email": "alice@company.com",
      "role": "admin",
      "permissions": ["read", "write", "delete", "admin"],
      "profile": {
        "avatar": "https://avatars.example.com/alice.jpg",
        "bio": "Senior Software Engineer",
        "location": "San Francisco, CA",
        "joinDate": "2023-01-15T00:00:00Z"
      },
      "preferences": {
        "theme": "dark",
        "language": "en",
        "notifications": {
          "email": true,
          "push": false,
          "sms": false
        }
      },
      "stats": {
        "loginCount": 247,
        "lastLogin": "2024-12-12T10:30:00Z",
        "projectsCreated": 15,
        "deploymentsManaged": 89
      }
    },
    "organization": {
      "id": "org_789",
      "name": "TechCorp Inc",
      "plan": "enterprise",
      "members": 42
    }
  },
  "meta": {
    "requestId": "req_abc123xyz",
    "timestamp": "2024-12-12T14:30:00.789Z",
    "version": "2.1",
    "cached": false
  }
}
\`\`\`

Complex nested JSON structures are much easier to explore with the interactive viewer.`}
        />
      </div>
    </div>
  </div>
)

export const Tables = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Table styling</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Markdown tables match the dashboard data table styling with rounded
        borders and header backgrounds.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple table</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`| Name | Status | Region |
|------|--------|--------|
| prod-cluster | Active | us-west-2 |
| staging-cluster | Active | us-east-1 |
| dev-cluster | Inactive | eu-west-1 |`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Wide table (horizontal scroll)</h4>
      <div className="p-4 border rounded-lg max-w-xl">
        <Markdown
          content={`| ID | Name | Status | Region | Provider | Version | CPU | Memory | Created |
|-----|------|--------|--------|----------|---------|-----|--------|---------|
| r-001 | prod-1 | Active | us-west-2 | AWS | 1.28 | 16 | 64GB | 2024-01-15 |
| r-002 | prod-2 | Active | us-east-1 | AWS | 1.28 | 16 | 64GB | 2024-02-20 |
| r-003 | staging | Updating | eu-west-1 | Azure | 1.27 | 8 | 32GB | 2024-03-10 |`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Mixed content</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`| # | Policy | Type | Enforcement | Trigger | Expected Result |
|---|--------|------|-------------|---------|-----------------|
| 1 | Public EKS Endpoint | \`sandbox\` / OPA | **warn** | \`cluster_endpoint_public_access = true\` in sandbox | Warning in UI — continue anyway |
| 2 | S3 Bucket Creation | \`terraform_module\` / OPA | **deny** | \`s3_bucket\` component creates an S3 bucket | Deploy blocked |
| 3 | Database Modification | \`terraform_module\` / OPA | **deny** | Change \`billing_mode\` input after first deploy | First deploy passes, redeploy blocked |
| 4 | Restricted Namespaces | \`helm_chart\` / OPA | **deny** | \`whoami_kube_system\` deploys to \`kube-system\` | Deploy blocked |
| 5a | Runner-Only Access | \`sandbox\` / OPA | **deny** | Non-runner IAM principals in EKS access entries | Deploy blocked (latent) |
| 5b | ECR Images Only | \`helm_chart\` / OPA | **deny** | \`traefik/whoami:latest\` is not from ECR | Deploy blocked |`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">HTML table</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`<table style="width:100%">
<thead>
<tr>
<th></th>
<th>Monitor</th>
<th>Status</th>
<th>Outputs</th>
</tr>
</thead>
<tbody>
<tr>
<td style="width: 1rem">🟢</td>
<td>api-healthcheck</td>
<td>finished</td>
<td><pre style="margin-top: 0; margin-bottom: 0">{"status": "ok", "latency_ms": 42}</pre></td>
</tr>
<tr>
<td style="width: 1rem">🔴</td>
<td>db-healthcheck</td>
<td>error</td>
<td><pre style="margin-top: 0; margin-bottom: 0">{"status": "error", "message": "connection timeout"}</pre></td>
</tr>
<tr>
<td style="width: 1rem">🟡</td>
<td>cache-healthcheck</td>
<td>running</td>
<td><pre style="margin-top: 0; margin-bottom: 0">{"status": "pending"}</pre></td>
</tr>
</tbody>
</table>`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Markdown table (equivalent)</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`| | Monitor | Status | Outputs |
|---|---------|--------|---------|
| 🟢 | api-healthcheck | finished | \`{"status": "ok", "latency_ms": 42}\` |
| 🔴 | db-healthcheck | error | \`{"status": "error", "message": "connection timeout"}\` |
| 🟡 | cache-healthcheck | running | \`{"status": "pending"}\` |`}
        />
      </div>
    </div>
  </div>
)

export const CollapsibleContent = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Enhanced Collapsible Content</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Details/summary elements now use styling inspired by the Expand
        component, providing consistent visual design and interactive behavior.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Collapsible Sections</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`## FAQ Section

<details>
<summary>What is the difference between CodeBlock and JSONViewer?</summary>

**CodeBlock** is used for syntax highlighting of various programming languages:
- JavaScript, TypeScript, Python, Go, etc.
- Syntax highlighting with themes
- Copy-to-clipboard functionality
- Language detection

**JSONViewer** is specifically for JSON data:
- Interactive tree view
- Expandable/collapsible nodes
- Data type indicators
- Object/array size information

</details>

<details>
<summary>How do external links work?</summary>

External links are automatically detected and:
- Use the custom Link component
- Open in new tabs
- Include proper security attributes
- Follow design system styling

Internal links use Next.js routing for better performance.

</details>

<details>
<summary>Can I use code blocks inside collapsible content?</summary>

Absolutely! Here's an example:

\`\`\`typescript
interface MarkdownConfig {
  enableCodeHighlighting: boolean
  enableJSONViewer: boolean
  enableMermaidDiagrams: boolean
}

const config: MarkdownConfig = {
  enableCodeHighlighting: true,
  enableJSONViewer: true,
  enableMermaidDiagrams: true
}
\`\`\`

And here's some JSON data:

\`\`\`json
{
  "features": {
    "codeBlocks": true,
    "jsonViewer": true,
    "collapsible": true,
    "links": "enhanced"
  }
}
\`\`\`

</details>

<details>
<summary>Advanced Content Example</summary>

This section demonstrates complex nested content:

### Tables Work Too

| Feature | Status | Notes |
|---------|--------|-------|
| Syntax Highlighting | ✅ Complete | Uses Prism.js |
| JSON Viewing | ✅ Complete | Interactive tree |
| Mermaid Diagrams | ✅ Complete | SVG rendering |
| Collapsible Content | ✅ Complete | Enhanced styling |

### Lists and Text Formatting

1. **First item** with emphasis
2. *Second item* with different emphasis  
3. ~~Crossed out item~~ for completeness
4. \`Inline code\` for technical terms

- Bullet point one
- Bullet point two with [a link](https://example.com)
- Bullet point three with more content

</details>`}
        />
      </div>
    </div>
  </div>
)

export const MermaidDiagrams = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Mermaid Diagrams</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Mermaid diagrams are rendered with a custom component that handles
        dynamic imports and error handling.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">System Architecture</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`## Component Dependencies

\`\`\`mermaid
graph TD
  cluster[cluster<br/>0-tf-cluster]
  repository[repository<br/>1-tf-repository]
  certificate[certificate<br/>1-tf-certificate]
  img[img_dashboard<br/>0-img-dashboard]
  builder[builder<br/>2-tf-builder]

  cluster --> builder
  repository --> builder
  certificate --> img
  img --> builder

  style builder fill:#D6B0FC,stroke:#8040BF,color:#000
  style cluster fill:#D6B0FC,stroke:#8040BF,color:#000
  style repository fill:#D6B0FC,stroke:#8040BF,color:#000
  style certificate fill:#D6B0FC,stroke:#8040BF,color:#000
  style img fill:#FCA04A,stroke:#FCA04A,color:#000
\`\`\`

This shows the component dependency graph with custom styling.`}
        />
      </div>
    </div>
  </div>
)

export const ComprehensiveExample = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Complete Feature Demonstration</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        This example demonstrates all the enhanced features working together:
        syntax highlighting, interactive JSON, mermaid diagrams, tables, and
        collapsible content.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Full Documentation Example</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`# API Documentation

## Overview

Our API provides comprehensive access to user management and analytics.

### Authentication

Include your bearer token in all requests:

\`\`\`bash
curl -H "Authorization: Bearer your-token-here" \\
     https://api.example.com/v1/users
\`\`\`

### Response Format

All API responses follow this structure:

\`\`\`json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "user_123",
        "name": "John Doe",
        "email": "john@example.com",
        "role": "admin",
        "lastLogin": "2024-12-12T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "perPage": 10,
      "total": 25
    }
  },
  "meta": {
    "requestId": "req_abc123",
    "timestamp": "2024-12-12T14:30:00.789Z"
  }
}
\`\`\`

### System Architecture

\`\`\`mermaid
sequenceDiagram
    participant Client
    participant API Gateway
    participant Auth Service
    participant User Service
    participant Database

    Client->>API Gateway: Request with token
    API Gateway->>Auth Service: Validate token
    Auth Service-->>API Gateway: Valid user
    API Gateway->>User Service: Get user data
    User Service->>Database: Query users
    Database-->>User Service: User data
    User Service-->>API Gateway: Response
    API Gateway-->>Client: JSON response
\`\`\`

## Rate Limits

| Plan | Requests/minute | Burst limit |
|------|----------------|-------------|
| Free | 100 | 120 |
| Pro | 1,000 | 1,200 |
| Enterprise | 10,000 | 12,000 |

## Advanced Usage

<details>
<summary>Click to see TypeScript example</summary>

### TypeScript Integration

\`\`\`typescript
interface User {
  id: string
  name: string
  email: string
  role: 'admin' | 'user' | 'moderator'
  metadata: {
    lastLogin: string
    loginCount: number
  }
}

class APIClient {
  private baseUrl: string
  private token: string

  constructor(baseUrl: string, token: string) {
    this.baseUrl = baseUrl
    this.token = token
  }

  async getUser(id: string): Promise<User | null> {
    const response = await fetch(\`\${this.baseUrl}/users/\${id}\`, {
      headers: {
        'Authorization': \`Bearer \${this.token}\`,
        'Content-Type': 'application/json'
      }
    })
    
    if (!response.ok) {
      throw new Error(\`HTTP error! status: \${response.status}\`)
    }
    
    const data = await response.json()
    return data.success ? data.data.user : null
  }
}
\`\`\`

</details>

<details>
<summary>Error Handling</summary>

### Common Error Codes

| Code | Description | Solution |
|------|-------------|----------|
| 401 | Unauthorized | Check your API token |
| 403 | Forbidden | Insufficient permissions |
| 429 | Rate Limited | Slow down requests |
| 500 | Server Error | Contact support |

</details>

Visit our [developer portal](https://developers.example.com) for more information.`}
        />
      </div>
    </div>
  </div>
)

export const NuonComponents = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Nuon embedded components</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Custom nuon-* tags render real React components from the dashboard.
        These can be used in user-authored markdown like runbooks.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Badge</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`Badges inline with text: <nuon-badge theme="success">Healthy</nuon-badge> <nuon-badge theme="error">Degraded</nuon-badge> <nuon-badge theme="warn">Pending</nuon-badge>

With sizes: <nuon-badge theme="info" size="sm">small</nuon-badge> <nuon-badge theme="info" size="md">medium</nuon-badge> <nuon-badge theme="info" size="lg">large</nuon-badge>`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Status</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`Status indicators:

<nuon-status status="active"></nuon-status>

<nuon-status status="error"></nuon-status>

<nuon-status status="provisioning"></nuon-status>

With badge variant: <nuon-status status="active" variant="badge"></nuon-status>`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Banner</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`## Runbook: restart procedure

<nuon-banner theme="warn">Only run this during a maintenance window.</nuon-banner>

Follow these steps:

1. Verify the install is in a stable state
2. Notify the team in Slack

<nuon-banner theme="info">This process typically takes 5-10 minutes to complete.</nuon-banner>

3. Proceed with the restart
4. Verify health checks pass

<nuon-banner theme="success">If all checks pass, the restart is complete.</nuon-banner>`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Combined example</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`# Environment status

| Environment | Status |
|---|---|
| Production | <nuon-badge theme="success">Healthy</nuon-badge> |
| Staging | <nuon-badge theme="warn">Deploying</nuon-badge> |
| Development | <nuon-badge theme="error">Down</nuon-badge> |

<nuon-banner theme="warn">Staging deployment in progress — avoid merging to main.</nuon-banner>

<details>
<summary>Troubleshooting steps for development</summary>

Check the current status: <nuon-status status="error"></nuon-status>

1. Check runner connectivity
2. Review recent deploy logs
3. Escalate if unresolved after 15 minutes

</details>`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Tabs</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`## Deployment guide

<nuon-tabs>
<nuon-tab name="AWS">

### AWS setup

1. Create an IAM role
2. Configure VPC networking
3. Deploy the stack

\`\`\`bash
nuon installs create --cloud aws
\`\`\`

</nuon-tab>
<nuon-tab name="Azure">

### Azure setup

1. Create a service principal
2. Configure VNet
3. Deploy the resources

\`\`\`bash
nuon installs create --cloud azure
\`\`\`

</nuon-tab>
</nuon-tabs>`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Tabs with embedded components</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`<nuon-tabs>
<nuon-tab name="status">

Current environment status:

<nuon-badge theme="success">Healthy</nuon-badge> Production

<nuon-badge theme="warn">Deploying</nuon-badge> Staging

</nuon-tab>
<nuon-tab name="runbook">

<nuon-banner theme="info">Follow these steps during a maintenance window.</nuon-banner>

1. Pause traffic
2. Run migrations
3. Resume traffic

</nuon-tab>
</nuon-tabs>`}
        />
      </div>
    </div>
  </div>
)

export const RenderModes = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Render modes: app vs install</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Action components like <code>nuon-view-state</code> require install context.
        In app mode they degrade to inline code. Display components with unresolved
        Go templates also degrade.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">App mode (default) — action tags render as code</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          mode="app"
          content={`## Runbook

Check the current install state:

<nuon-view-state></nuon-view-state>

View the dependency graph:

<nuon-config-graph></nuon-config-graph>

Display components still render: <nuon-badge theme="success">Healthy</nuon-badge>`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">App mode — unresolved templates degrade to code</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          mode="app"
          content={`Status with template: <nuon-badge theme="success">{"{{ .install.status }}"}</nuon-badge>

Resolved badge still renders: <nuon-badge theme="info">Ready</nuon-badge>`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Install mode — action tags render real components</h4>
      <p className="text-xs text-gray-500 dark:text-gray-500">
        ViewStateButton renders as a real button (clicking opens the modal with mock context).
      </p>
      <div className="p-4 border rounded-lg">
        <MockInstallProviders>
          <Markdown
            mode="install"
            content={`## Runbook

Check the current install state:

<nuon-view-state></nuon-view-state>

View the dependency graph:

<nuon-config-graph></nuon-config-graph>

Display components also render: <nuon-badge theme="success">Healthy</nuon-badge>`}
          />
        </MockInstallProviders>
      </div>
    </div>
  </div>
)

export const HTMLIntegration = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Raw HTML Integration</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The Markdown component supports raw HTML via the rehype-raw plugin,
        allowing you to mix HTML elements with markdown content seamlessly.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Mixed HTML and Markdown</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`# Mixed Content Example

This is regular **markdown** content with some *formatting*.

<div class="bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 rounded-lg p-4 my-4">
  <h4 class="text-blue-900 dark:text-blue-100 font-semibold mb-2">💡 Custom HTML Alert</h4>
  <p class="text-blue-800 dark:text-blue-200 mb-0">
    This is a custom HTML div with Tailwind classes that renders within markdown content.
    You can use any HTML elements and they will be processed alongside markdown.
  </p>
</div>

Back to regular markdown. Here's a list:

- First markdown item
- Second item with \`inline code\`

<table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700 my-4">
  <thead class="bg-gray-50 dark:bg-gray-800">
    <tr>
      <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
        Feature
      </th>
      <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
        Status
      </th>
    </tr>
  </thead>
  <tbody class="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
    <tr>
      <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
        HTML Integration
      </td>
      <td class="px-6 py-4 whitespace-nowrap">
        <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200">
          ✅ Supported
        </span>
      </td>
    </tr>
    <tr>
      <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
        Markdown Processing
      </td>
      <td class="px-6 py-4 whitespace-nowrap">
        <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200">
          ✅ Active
        </span>
      </td>
    </tr>
  </tbody>
</table>

And we can continue with more markdown content, including **bold text** and [links](https://example.com).

<div class="flex items-center space-x-2 my-4 p-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded">
  <svg class="w-5 h-5 text-yellow-600 dark:text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
    <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"></path>
  </svg>
  <p class="text-sm text-yellow-800 dark:text-yellow-200 mb-0">
    <strong>Note:</strong> HTML elements inherit the design system colors and work seamlessly with dark mode.
  </p>
</div>`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Interactive HTML Elements</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={`## Interactive Components

You can embed interactive HTML elements:

<div class="space-y-4">
  <details class="border border-gray-200 dark:border-gray-700 rounded-lg">
    <summary class="px-4 py-2 bg-gray-50 dark:bg-gray-800 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700">
      Click to reveal custom HTML details
    </summary>
    <div class="p-4">
      <p class="text-sm text-gray-600 dark:text-gray-400">
        This is a custom HTML details/summary element with Tailwind styling.
        It works alongside the markdown-native details elements.
      </p>
    </div>
  </details>

  <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
    <div class="bg-primary-50 dark:bg-primary-900/20 border border-primary-200 dark:border-primary-800 rounded-lg p-4">
      <h5 class="font-semibold text-primary-900 dark:text-primary-100 mb-2">Primary Card</h5>
      <p class="text-sm text-primary-700 dark:text-primary-300">Custom HTML card with design system colors.</p>
    </div>
    <div class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-4">
      <h5 class="font-semibold text-green-900 dark:text-green-100 mb-2">Success Card</h5>
      <p class="text-sm text-green-700 dark:text-green-300">Another custom card using semantic colors.</p>
    </div>
    <div class="bg-orange-50 dark:bg-orange-900/20 border border-orange-200 dark:border-orange-800 rounded-lg p-4">
      <h5 class="font-semibold text-orange-900 dark:text-orange-100 mb-2">Warning Card</h5>
      <p class="text-sm text-orange-700 dark:text-orange-300">Third card demonstrating color consistency.</p>
    </div>
  </div>
</div>

Regular markdown continues to work perfectly alongside the HTML elements.

> **Pro tip:** You can use any Tailwind classes in your HTML elements and they'll integrate seamlessly with the design system.`}
        />
      </div>
    </div>
  </div>
)
