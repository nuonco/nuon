import { Markdown } from './Markdown'

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

/* eslint-disable no-irregular-whitespace */
const BYOC_README = `


<center>
  <img class="mt-0 block dark:hidden" src="https://mintlify.s3-us-west-1.amazonaws.com/nuoninc/logo/light.svg"/>
  <img class="mt-0 hidden dark:block" src="https://mintlify.s3-us-west-1.amazonaws.com/nuoninc/logo/dark.svg"/>
  <small>

AWS | 873046390432 | us-west-2 | vpc-0abf81c9ba0a13384

  </small>

<small>[DataDog](https://us5.datadoghq.com/logs?query=env%3Abyoc%20install.id%3Ainl2a3gxirvbecgjxq5mz0avrb)</small> | <small>[Dashboard](https://app.byoc.retool.com)</small> | <small>[API](https://api.byoc.retool.com/docs/index.html)</small>

</center>

<div>
    <table style="width:100%">
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
                    <td>alb_healthcheck_app</td>
                    <td>finished</td>
                    <td><pre style="margin-top: 0; margin-bottom: 0">map[certificate_arn:arn:aws:acm:us-west-2:873046390432:certificate/5cf98917-9ba1-452b-ae8b-a603f791af13 hostname:app.byoc.retool.com indicator:🟢]</pre></td>
                </tr>
                <tr>
                    <td style="width: 1rem">🟢</td>
                    <td>alb_healthcheck_private</td>
                    <td>finished</td>
                    <td><pre style="margin-top: 0; margin-bottom: 0">map[certificate_arn:null hostname:admin.internal.byoc.retool.com indicator:🟢]</pre></td>
                </tr>
                <tr>
                    <td style="width: 1rem">🟢</td>
                    <td>alb_healthcheck_public</td>
                    <td>finished</td>
                    <td><pre style="margin-top: 0; margin-bottom: 0">map[certificate_arn:arn:aws:acm:us-west-2:873046390432:certificate/5cf98917 hostname:api.byoc.retool.com indicator:🟢]</pre></td>
                </tr>
                <tr>
                    <td style="width: 1rem">🟢</td>
                    <td>alb_healthcheck_runner</td>
                    <td>finished</td>
                    <td><pre style="margin-top: 0; margin-bottom: 0">map[certificate_arn:arn:aws:acm:us-west-2:873046390432:certificate/5cf98917 hostname:runner.byoc.retool.com indicator:🟢]</pre></td>
                </tr>
                <tr>
                    <td style="width: 1rem">🟢</td>
                    <td>healthcheck_temporal</td>
                    <td>finished</td>
                    <td><pre style="margin-top: 0; margin-bottom: 0">map[indicator:🟢 steps:map[tctl cluster health:map[indicator:🟢]]]</pre></td>
                </tr>
        </tbody>
    </table>
</div>


<details>
<summary><strong>Installing Nuon</strong></summary>

Nuon has a few dependencies you must configure ahead of time.

- Custom DNS (Optional)
- Github App
- Google OAuth

You will need an install ID to configure these. For this reason, the first step in the installation process is to create
your Nuon install -- don't bother updating any of the inputs -- and then cancel the provision. You will use the install
ID to configure the dependencies as detailed below. Once the dependencies are ready, update your install's inputs, then
click on "Reprovision Install" in the "Manage" menu.

### Configure DNS

There are two domains at play with a BYOC Nuon deployment. The first is the \`root_domain\` under which all of the
services (e.g. APIs & Frontend) are served. The second, \`nuon_dns_domain\`, is a domain which you can use to
automate the provisioning of Route53 zones for installs. This second feature is optional but, at the time of writing, a
default value must be provided.

|                 | Input                  | Description                                                                          |
| --------------- | ---------------------- | ------------------------------------------------------------------------------------ |
| Root Domain     | \`root_domain\`          | The root domain from which the nuon services are served.                             |
| Nuon DNS Domain | \`nuon_dns_domain\` | The domain used to provision domains for installs managed by this BYOC Nuon Install. |

BYOC Nuon should be hosted under a custom domain of your choice, for example:

- \`byoc.organization.com\`

Nuon DNS should be hosted under a separate domain or a dedicated subdomain, such as:

- \`installs.organization.com\`
- \`hosted.organization.io\`


> [!NOTE]
> We strongly suggest you choose your domains so there is NO overlap between the two.


Nuon DNS is optional, but a valid domain should be provided during installation nonetheless. It will remain inactive so
long as app's set \`enable_nuon_dns\` to \`false\` in the sandbox configs
([link](https://github.com/nuonco/aws-eks-sandbox?tab=readme-ov-file#input_enable_nuon_dns)).


> [!NOTE]
> All Nuon-authored sandboxes implement a nuon_dns module whose outputs nuon knows how to read.


#### Current DNS Configurations

When an install is created, a Route53 zone will be created for each of the domains. When these are ready, you can use
those details to configure your domain in your registrar to use the AWS nameservers.



##### Root Domain

| Attribute   | Value                                                                                          |
| ----------- | ---------------------------------------------------------------------------------------------- |
| Domain Name | byoc.retool.com                                                                           |
| Zone ID     | Z00820512O0HS8UQ4RYVH |


| Value     | Record Type | priority |
| --------- | ----------- | -------- |
| ns-1143.awsdns-14.org | NS          | 0   |
| ns-1584.awsdns-06.co.uk | NS          | 1   |
| ns-473.awsdns-59.com | NS          | 2   |
| ns-523.awsdns-01.net | NS          | 3   |




##### Nuon DNS Root Domain

| Attribute   | Value                                                          |
| ----------- | -------------------------------------------------------------- |
| Domain Name | byoc.retool.com  |
| Zone ID     | Z0953494FT4NDVLRRWD1 |


| Value     | Record Type | priority |
| --------- | ----------- | -------- |
| ns-1205.awsdns-22.org | NS          | 0   |
| ns-1641.awsdns-13.co.uk | NS          | 1   |
| ns-53.awsdns-06.com | NS          | 2   |
| ns-938.awsdns-53.net | NS          | 3   |



Additional Documentation

- [Creating a subdomain that uses Amazon Route 53 as the DNS service without migrating the parent domain](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/CreatingNewSubdomain.html)

### Configure Github App

Create a github app so BYOC Nuon can clone code for components from private repos. (To configure a new App:
https://github.com/settings/apps) Configure it thusly:

- Github app name: (pick any name)
- Homepage URL: [https://app.byoc.retool.com](https://app.byoc.retool.com)
- Post Installation:
  - Setup URL: [https://app.byoc.retool.com/connect](https://app.byoc.retool.com/connect)
  - Redirect on Update: check
- Webhook:
  - Webhook: un-check
- Permissions:
  - Contents: Read-only
  - Where can this GitHub app be installed?: Only on this account. (unless you have repos you need to access in other
    GitHub accounts.)

Once the app has been created, scroll to the bottom and generate a PEM key. You will need to provide this as a secret
later.

### Configure Google OAuth

Nuon uses Google OAuth for authentication. Users will sign in with their Google account.

The user key in the BYOC application is \`email\`. Organizations and apps are associated with the user based on this key.

#### Create Google OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create or select a project
3. Navigate to **APIs & Services** > **Credentials**
4. Click **Create Credentials** > **OAuth client ID**
5. Select **Web application** as the application type
6. Configure the OAuth client (see table below)
7. Note the **Client ID** and **Client Secret** - you'll need these for the install inputs and secrets

| Setting                       | Value                                    |
| ----------------------------- | ---------------------------------------- |
| Name                          | \`BYOC Nuon\` (or any name)                |
| Authorized JavaScript origins | \`https://auth.byoc.retool.com\`      |
| Authorized redirect URIs      | \`https://auth.byoc.retool.com/auth\` |

### Update Inputs

Once the dependencies have been configured, you can update your install inputs. This will trigger a workflow that's
going to fail because the install hasn't been provisioned yet. This won't cause any problems, and you can ignore it.

#### Authentication Configuration

| Input              | Value                                                                                 |
| ------------------ | ------------------------------------------------------------------------------------- |
| Auth Provider Type | \`google\` (default)                                                                    |
| Auth Client ID     | Client ID from Google OAuth credentials                                               |
| Auth Redirect URL  | Defaults to \`https://auth.byoc.retool.com/auth\` |

#### Github

| Input                | Value                              |
| -------------------- | ---------------------------------- |
| Github App Name      | name of your github app            |
| Github App ID        | ID of your github app              |
| Github App client ID | the client ID from your Github app |

#### DNS Configuration

|                 | Input                                                         | Description                                                                          |
| --------------- | ------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| Root Domain     | \`byoc.retool.com\`       | The root domain from which the nuon services are served.                             |
| Nuon DNS Domain | \`byoc.retool.com\` | The domain used to provision domains for installs managed by this BYOC Nuon Install. |

### Update Secrets

When provisioning the install CloudFormation stack, you will need to provide the following secrets.

| Secret                    | Value                               |
| ------------------------- | ----------------------------------- |
| \`github_app_key\`          | your base64 encoded PEM key         |
| \`nuon_auth_client_secret\` | the client secret from Google OAuth |

The github app PEM key must be base64 encoded. AWS CloudFormation does not preserve newlines in text fields. By encoding
the PEM key before pasting it in, and decoding it later when it's read, we can preserve the newlines in the text.

The following secrets are auto-generated and do not need to be provided:

- \`nuon_auth_session_key\` - used for session nonce
- \`nuon_auth_jwt_secret\` - used to sign JWT tokens

</details>

<details>
<summary><strong>Application Links</strong></summary>

Once Nuon is successfully provisioned, you can inspect it at the following URLs.

| Service   | URL                                                       |
| --------- | --------------------------------------------------------- |
| Dashboard | [app.byoc.retool.com](https://app.byoc.retool.com)       |
| CTL API   | [api.byoc.retool.com](https://api.byoc.retool.com)       |
| Runner API| [runner.byoc.retool.com](https://runner.byoc.retool.com)  |

</details>

<details>
<summary><strong>Accessing the EKS Cluster</strong></summary>

1. Add an access entry for the relevant role.
2. Grant the following perms: AWSEKSAdmin, AWSClusterAdmin.gtg
3. Add the cluster kubeconfig w/ the following command.

<pre>
aws --region us-west-2 \\
    --profile your.Profile eks update-kubeconfig      \\
    --name n-inl2a3gxirvbecgjxq5mz0avrb \\
    --alias n-inl2a3gxirvbecgjxq5mz0avrb
</pre>

</details>

<details>
<summary><strong>Secrets</strong></summary>

The following secrets are created in the CloudFormation stack and then synced into the cluster.

| Secret                   | Key(s)            | Namespace  | name                       | Source                    | Description                                 |
| ------------------------ | ----------------- | ---------- | -------------------------- | ------------------------- | ------------------------------------------- |
| clickhouse-operator-pw   | value             | clickhouse | clickhouse-operator-pw     | secrets-sync              | clickhouse operator password                |
| clickhouse-cluster-ro-pw | value             | clickhouse | clickhouse-cluster-ro-pw   | secrets-sync              | clickhouse cluster readonly user password   |
| clickhouse-cluster-pw    | value             | clickhouse | clickhouse-cluster-pw      | secrets-sync              | clickhouse cluster read/write user password |
| github-app-key           | value             | ctl-api    | github-app-key             | secrets-sync              | github app key                              |
| nuon_auth_client_secret  | value             | ctl-api    | ctl-api-auth-client-secret | secrets-sync              | OIDC client secret                          |
| nuon_auth_session_key    | value             | ctl-api    | ctl-api-auth-session-key   | secrets-sync              | Auto-generated session key                  |
| nuon_auth_jwt_secret     | value             | ctl-api    | ctl-api-auth-jwt-secret    | secrets-sync              | Auto-generated JWT signing secret           |

### Updating Secrets

Secrets can be updated by re-provisioning the stack and updating the secret values.

1. Re-provision Install
2. Wait for the Install Stack to be updated.
3. Open the CF link and copy the template url.
4. Navigate to your stack and Click "Update Stack" then click on "Create a changeset".
5. Select "Replace Existing Template" and paste the newly generated S3 URL.
6. Click Next
7. Review the changes and update the secrets as necessary.
8. Click Next
9. Click the toggles and click Next.
10. Review changes and click Submit.
11. Wait for changes to be calculated then click "Execute change set".
12. Accept settings, click "Execute change set."

</details>

<details>
<summary><strong>Runners</strong></summary>

<table>
    <thead>
        <tr>
            <th></th>
            <th>ID</th>
            <th>Org ID</th>
            <th>Tag</th>
            <th>Created At</th>
            <th>Updated At</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>🔴</td>
            <td><code>run25a0bg7cpwi3tsacvyj90nb</code></td>
            <td><code>org4oic0k3423h74okoxaspm0n</code></td>
            <td><code>0.19.810</code></td>
            <td>2026-03-10 21:07</td>
            <td>2026-03-11 21:32</td>
        </tr>
        <tr>
            <td>🟢</td>
            <td><code>run2av86175xai9t3n1120kuk8</code></td>
            <td><code>org4oic0k3423h74okoxaspm0n</code></td>
            <td><code>0.19.810</code></td>
            <td>2026-03-10 23:22</td>
            <td>2026-03-11 00:23</td>
        </tr>
        <tr>
            <td>🟢</td>
            <td><code>run2f0i4yoj1d8bxc7rzu3n6yc</code></td>
            <td><code>org4oic0k3423h74okoxaspm0n</code></td>
            <td><code>0.19.810</code></td>
            <td>2025-06-10 18:44</td>
            <td>2026-03-21 00:18</td>
        </tr>
        <tr>
            <td>🟢</td>
            <td><code>run2gc86sjo4n7odqg7rf796yj</code></td>
            <td><code>org4oic0k3423h74okoxaspm0n</code></td>
            <td><code>cloud</code></td>
            <td>2025-06-24 18:36</td>
            <td>2026-02-13 20:15</td>
        </tr>
    </tbody>
</table>

</details>

<details>
<summary><strong>Components</strong></summary>

### RDS Clusters

The nuon cluster is created w/ an admin user and a \`nuon\` db. This admin user is responsible for creating the \`ctl_api\`
user and db. This is done in an [action](/actions/).

</details>

<details>
<summary><strong>CLI</strong></summary>

Install the latest version of the nuon cli ([docs](https://docs.nuon.co/cli#cli)).

\`\`\`bash
brew install nuonco/tap/nuon
\`\`\`

Update your \`~/.nuon\` config or create one specifically for this byoc install (e.g. \`~/.nuon.byoc\`).

Configure as follows:

\`\`\`yaml
api_url: https://api.byoc.retool.com
\`\`\`

Log in:

\`\`\`yaml
nuon -f ~/.nuon.byoc login
\`\`\`

</details>
`

export const BYOCReadme = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">BYOC Install Readme</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        A real-world rendered readme from a BYOC Nuon install, containing mixed
        HTML tables, collapsible sections, markdown tables, code blocks, and
        inline code.
      </p>
    </div>
    <div className="p-4 border rounded-lg">
      <Markdown content={BYOC_README} />
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
