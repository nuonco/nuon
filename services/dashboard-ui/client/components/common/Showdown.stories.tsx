/* eslint-disable react/no-unescaped-entities */
import { Markdown as Showdown } from './Showdown'
import { Text } from './Text'
import { Badge } from './Badge'
import { Button } from './Button'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Markdown Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The Markdown component renders markdown content as HTML using the
        showdown library. It supports GitHub-flavored markdown features,
        automatically opens external links in new tabs, and provides proper
        styling for all standard markdown elements.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple Markdown Example</h4>
      <div className="p-4 border rounded-lg">
        <Showdown
          content={`# Hello World

This is a basic markdown example with **bold** and *italic* text.

## Features

- Easy to use
- GitHub-flavored markdown
- Automatic link handling

Check out [this external link](https://example.com) that opens in a new tab.`}
        />
      </div>
      <Text variant="subtext" theme="neutral">
        External links automatically open in new tabs for better user experience
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Renders markdown as semantic HTML with proper styling</li>
        <li>Supports all standard markdown syntax and GitHub extensions</li>
        <li>External links open in new tabs automatically</li>
        <li>Code blocks with syntax highlighting support</li>
        <li>Tables, task lists, and collapsible content</li>
      </ul>
    </div>
  </div>
)

const complexMarkdownContent = `# Markdown Examples

This component renders markdown as HTML using the showdown library.

## Headers

### Sub Header

#### Sub Sub Header

## Text Formatting

This is **bold text** and this is *italic text*.

You can also use ~~strikethrough~~ text.

## Lists

### Unordered List
- Item one
- Item two
- Item three
  - Nested item
  - Another nested item

### Ordered List
1. First item
2. Second item
3. Third item

## Code Blocks

Inline code: \`const x = 42;\`

\`\`\`javascript
function greet(name) {
  return \`Hello, \${name}!\`;
}

console.log(greet('World'));
\`\`\`

## Links and External Links

[Internal link](#section)
[External link](https://example.com) - opens in new tab

## Tables

| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Row 1    | Data     | More data |
| Row 2    | Info     | More info |

## Task Lists

- [x] Completed task
- [ ] Incomplete task
- [ ] Another task

## Blockquotes

> This is a blockquote.
> It can span multiple lines.

## Horizontal Rule

---

## Collapsible Content

<details>
<summary>Click to expand details</summary>

This content is hidden by default and can be expanded by clicking the summary.

You can include any markdown content inside:

- Lists
- **Bold text**
- Code: \`const x = 42;\`
- Even tables:

| Feature | Supported |
|---------|-----------|
| Details | ✅ Yes    |
| Summary | ✅ Yes    |

</details>

<details>
<summary>Another collapsible section</summary>

### Nested content

This shows how you can nest other markdown elements inside details blocks.

\`\`\`javascript
// Even code blocks work
function example() {
  console.log('Inside details block!');
}
\`\`\`

</details>

## HTML

Raw HTML is also supported: <strong>Bold HTML</strong>`

export const TypographyElements = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Typography Elements</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Markdown supports a full range of typography elements including headers,
        text formatting, lists, and more. All elements are properly styled and
        maintain consistency with the design system.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Headers and Text Formatting</h4>
      <div className="p-4 border rounded-lg">
        <Showdown
          content={`# Main Header (H1)

## Section Header (H2)

### Subsection Header (H3)

#### Sub-subsection Header (H4)

Regular paragraph text with **bold**, *italic*, and ~~strikethrough~~ formatting.

You can combine formatting like ***bold and italic*** text.

> This is a blockquote that can contain multiple lines
> and provides emphasis for important information.

---

Horizontal rules create visual separation between content sections.`}
        />
      </div>
    </div>
  </div>
)

export const ListsAndTables = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Lists and Tables</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Markdown provides comprehensive support for both ordered and unordered
        lists, as well as tables with proper alignment and styling. Task lists
        are also supported for interactive content.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Lists and Tables Example</h4>
      <div className="p-4 border rounded-lg">
        <Showdown
          content={`## Lists

### Unordered List
- First item
- Second item with a longer description
- Third item
  - Nested item
  - Another nested item
    - Deeply nested item

### Ordered List
1. First numbered item
2. Second numbered item
3. Third numbered item
   1. Nested numbered item
   2. Another nested numbered item

### Task Lists
- [x] Completed task
- [x] Another completed task
- [ ] Incomplete task
- [ ] Future task to complete

## Tables

| Feature | Status | Priority |
|---------|--------|----------|
| Authentication | ✅ Complete | High |
| Dashboard | 🚧 In Progress | High |
| API Integration | ⏳ Planned | Medium |
| Documentation | ✅ Complete | Low |

### Aligned Table

| Left Aligned | Center Aligned | Right Aligned |
|:-------------|:--------------:|--------------:|
| Text | Text | Text |
| More content | Centered | Right |
| Final row | Middle | End |`}
        />
      </div>
    </div>
  </div>
)

export const CodeBlocks = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Code Blocks with Syntax Highlighting</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The Markdown component now supports advanced code block rendering with our custom
        CodeBlock component for syntax highlighting and interactive JSONViewer for JSON content.
        Code blocks automatically detect the language and render appropriately.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">JavaScript/TypeScript Examples</h4>
      <div className="p-4 border rounded-lg">
        <Showdown content={`## JavaScript

\`\`\`javascript
// React component with hooks
import React, { useState, useEffect } from 'react'

function UserProfile({ userId }) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)
  
  useEffect(() => {
    async function fetchUser() {
      try {
        const response = await fetch(\`/api/users/\${userId}\`)
        const userData = await response.json()
        setUser(userData)
      } catch (error) {
        console.error('Failed to fetch user:', error)
      } finally {
        setLoading(false)
      }
    }
    
    fetchUser()
  }, [userId])
  
  if (loading) return <div>Loading...</div>
  
  return (
    <div className="user-profile">
      <h2>{user?.name}</h2>
      <p>{user?.email}</p>
    </div>
  )
}
\`\`\`

## TypeScript

\`\`\`typescript
interface User {
  id: string
  name: string
  email: string
  role: 'admin' | 'user' | 'moderator'
  preferences: {
    theme: 'dark' | 'light'
    notifications: boolean
  }
}

class UserService {
  private apiUrl: string
  
  constructor(apiUrl: string) {
    this.apiUrl = apiUrl
  }
  
  async getUser(id: string): Promise<User | null> {
    try {
      const response = await fetch(\`\${this.apiUrl}/users/\${id}\`)
      if (!response.ok) {
        throw new Error(\`HTTP error! status: \${response.status}\`)
      }
      return await response.json()
    } catch (error) {
      console.error('Error fetching user:', error)
      return null
    }
  }
  
  async updateUser(id: string, updates: Partial<User>): Promise<boolean> {
    try {
      const response = await fetch(\`\${this.apiUrl}/users/\${id}\`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updates)
      })
      return response.ok
    } catch {
      return false
    }
  }
}
\`\`\``} />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">JSON Data (Interactive JSONViewer)</h4>
      <div className="p-4 border rounded-lg">
        <Showdown content={`## Configuration Object

\`\`\`json
{
  "name": "dashboard-ui",
  "version": "2.1.0",
  "dependencies": {
    "react": "^18.0.0",
    "next": "^14.0.0",
    "typescript": "^5.0.0"
  },
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start"
  },
  "config": {
    "api": {
      "baseUrl": "https://api.example.com",
      "timeout": 5000,
      "retries": 3
    },
    "features": {
      "analytics": true,
      "darkMode": true,
      "notifications": false
    },
    "theme": {
      "colors": {
        "primary": "#8040BF",
        "secondary": "#61AFEF",
        "success": "#98C379",
        "warning": "#D19A66",
        "error": "#E06C75"
      },
      "fonts": {
        "sans": "Inter, sans-serif",
        "mono": "Hack, monospace"
      }
    }
  }
}
\`\`\`

## API Response Example

\`\`\`json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "user_123",
        "name": "Alice Johnson",
        "email": "alice@example.com",
        "role": "admin",
        "lastLogin": "2024-12-12T10:30:00Z",
        "permissions": ["read", "write", "delete"],
        "metadata": {
          "createdAt": "2024-01-15T09:00:00Z",
          "updatedAt": "2024-12-10T15:45:00Z",
          "loginCount": 142
        }
      }
    ],
    "pagination": {
      "page": 1,
      "perPage": 10,
      "total": 25,
      "totalPages": 3
    }
  },
  "meta": {
    "requestId": "req_abc123",
    "timestamp": "2024-12-12T14:30:00.789Z",
    "version": "v2.1"
  }
}
\`\`\``} />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Backend Languages</h4>
      <div className="p-4 border rounded-lg">
        <Showdown content={`## Python

\`\`\`python
from fastapi import FastAPI, HTTPException, Depends
from sqlalchemy.orm import Session
from typing import List, Optional
import uvicorn

app = FastAPI(title="User API", version="1.0.0")

class UserService:
    def __init__(self, db: Session):
        self.db = db
    
    def get_user(self, user_id: str) -> Optional[User]:
        """Retrieve a user by ID"""
        return self.db.query(User).filter(User.id == user_id).first()
    
    def create_user(self, user_data: dict) -> User:
        """Create a new user"""
        user = User(**user_data)
        self.db.add(user)
        self.db.commit()
        self.db.refresh(user)
        return user

@app.get("/users/{user_id}")
async def get_user(user_id: str, db: Session = Depends(get_db)):
    user = UserService(db).get_user(user_id)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return {"success": True, "data": user}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
\`\`\`

## Go

\`\`\`go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
    
    "github.com/gorilla/mux"
    "gorm.io/gorm"
)

type User struct {
    ID        string    \`json:"id" gorm:"primaryKey"\`
    Name      string    \`json:"name"\`
    Email     string    \`json:"email"\`
    Role      string    \`json:"role"\`
    CreatedAt time.Time \`json:"created_at"\`
    UpdatedAt time.Time \`json:"updated_at"\`
}

type UserHandler struct {
    db *gorm.DB
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    userID := vars["id"]
    
    var user User
    if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    
    response := map[string]interface{}{
        "success": true,
        "data":    user,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {
    r := mux.NewRouter()
    handler := &UserHandler{db: setupDB()}
    
    r.HandleFunc("/users/{id}", handler.GetUser).Methods("GET")
    
    fmt.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
\`\`\``} />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Configuration Files</h4>
      <div className="p-4 border rounded-lg">
        <Showdown content={`## Docker Configuration

\`\`\`dockerfile
# Multi-stage build for Node.js application
FROM node:18-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

# Production stage
FROM node:18-alpine AS runner

WORKDIR /app

# Create non-root user
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Copy built application
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000
ENV PORT 3000
ENV NODE_ENV production

CMD ["node", "server.js"]
\`\`\`

## YAML Configuration

\`\`\`yaml
# kubernetes deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dashboard-ui
  namespace: production
  labels:
    app: dashboard-ui
    version: v2.1.0
spec:
  replicas: 3
  selector:
    matchLabels:
      app: dashboard-ui
  template:
    metadata:
      labels:
        app: dashboard-ui
    spec:
      containers:
      - name: dashboard-ui
        image: dashboard-ui:v2.1.0
        ports:
        - containerPort: 3000
        env:
        - name: NODE_ENV
          value: "production"
        - name: API_URL
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: api-url
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 30
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: dashboard-ui-service
spec:
  selector:
    app: dashboard-ui
  ports:
  - port: 80
    targetPort: 3000
  type: ClusterIP
\`\`\`

## Shell Script

\`\`\`bash
#!/bin/bash

# Deployment script for dashboard-ui
set -e

# Configuration
APP_NAME="dashboard-ui"
VERSION=\${1:-"latest"}
REGISTRY="your-registry.com"
NAMESPACE="production"

echo "🚀 Starting deployment of \$APP_NAME:\$VERSION"

# Build and push Docker image
echo "📦 Building Docker image..."
docker build -t \$REGISTRY/\$APP_NAME:\$VERSION .
docker push \$REGISTRY/\$APP_NAME:\$VERSION

# Update Kubernetes deployment
echo "🔄 Updating Kubernetes deployment..."
kubectl set image deployment/\$APP_NAME \$APP_NAME=\$REGISTRY/\$APP_NAME:\$VERSION -n \$NAMESPACE

# Wait for rollout to complete
echo "⏳ Waiting for deployment to complete..."
kubectl rollout status deployment/\$APP_NAME -n \$NAMESPACE --timeout=300s

# Verify deployment
echo "✅ Verifying deployment..."
kubectl get pods -n \$NAMESPACE -l app=\$APP_NAME

echo "🎉 Deployment completed successfully!"
\`\`\``} />
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Code Block Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>**Automatic language detection** - Syntax highlighting based on language specification</li>
        <li>**Interactive JSON viewing** - JSON code blocks render with expandable/collapsible JSONViewer</li>
        <li>**Theme consistency** - Code blocks use the same One Dark/One Light themes as the rest of the UI</li>
        <li>**Fallback handling** - Invalid JSON falls back to syntax-highlighted code block</li>
        <li>**Wide language support** - JavaScript, TypeScript, Python, Go, Docker, YAML, Bash, and more</li>
        <li>**Responsive design** - Code blocks adapt to container width with horizontal scrolling when needed</li>
      </ul>
    </div>
  </div>
)

export const InteractiveContent = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">
        Interactive and Advanced Content
      </h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Markdown supports advanced features like collapsible sections, HTML
        elements, and interactive content. This makes it suitable for
        documentation, help content, and rich text displays.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Advanced Features</h4>
      <div className="p-4 border rounded-lg">
        <Showdown
          content={`## Collapsible Content

<details>
<summary>Click to expand API documentation</summary>

### Authentication

All API requests require a valid bearer token:

\`\`\`bash
curl -H "Authorization: Bearer your-token-here" \
     https://api.example.com/v1/users
\`\`\`

### Response Format

All responses are in JSON format:

\`\`\`json
{
  "status": "success",
  "data": {
    "id": "user_123",
    "name": "John Doe"
  }
}
\`\`\`

</details>

<details>
<summary>Implementation Examples</summary>

### React Component Usage

\`\`\`tsx
import { Markdown } from './Markdown'

function DocumentationPage() {
  const content = \`# Welcome\n\nThis is **markdown** content!\`
  return <Showdown content={content} />
}
\`\`\`

### Features List

- [x] GitHub-flavored markdown support
- [x] Automatic external link handling
- [x] Code syntax highlighting
- [x] Table support with alignment
- [x] Task list support
- [ ] Custom styling options
- [ ] Plugin system

</details>

## HTML Support

Raw HTML elements work within markdown:

<div style="background: #f0f8ff; padding: 16px; border-radius: 8px; border-left: 4px solid #0066cc;">
  <strong>Info:</strong> You can use HTML for custom styling when needed.
</div>`}
        />
      </div>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The Markdown component is commonly used for documentation, help content,
        user-generated content, and any interface where rich text formatting is
        needed.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Documentation Panel</h4>
      <div className="p-4 border rounded-lg space-y-4">
        <div className="flex justify-between items-center">
          <Text variant="h3" weight="stronger">
            API Documentation
          </Text>
          <Badge theme="info">v2.1</Badge>
        </div>
        <Showdown
          content={`## Quick Start

Get started with our API in minutes:

1. **Get your API key** from the dashboard
2. **Make your first request** using curl or your favorite HTTP client
3. **Explore the endpoints** using our interactive documentation

### Authentication

Include your API key in the Authorization header:

\`\`\`bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
     https://api.example.com/v1/endpoint
\`\`\`

### Rate Limits

| Plan | Requests per minute |
|------|--------------------|
| Free | 100 |
| Pro  | 1,000 |
| Enterprise | Unlimited |

> **Tip:** Use pagination to efficiently handle large datasets.`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Help Content</h4>
      <div className="p-4 border rounded-lg space-y-4">
        <div className="flex justify-between items-center">
          <Text variant="h3" weight="stronger">
            Troubleshooting Guide
          </Text>
          <Button variant="ghost" size="sm">
            Contact Support
          </Button>
        </div>
        <Showdown
          content={`## Common Issues

### Connection Problems

If you're experiencing connection issues:

- [ ] Check your internet connection
- [ ] Verify your API credentials
- [ ] Confirm the endpoint URL is correct
- [ ] Check our [status page](https://status.example.com) for outages

### Authentication Errors

**Error 401: Unauthorized**

This usually means your API key is invalid or expired.

**Solutions:**
1. Generate a new API key from your dashboard
2. Check that you're using the correct header format
3. Ensure your key hasn't been accidentally modified

<details>
<summary>Advanced troubleshooting</summary>

If you're still having issues:

\`\`\`bash
# Test your connection
curl -v https://api.example.com/v1/health

# Validate your API key
curl -H "Authorization: Bearer YOUR_KEY" \
     https://api.example.com/v1/validate
\`\`\`

Check the response headers for additional error information.

</details>`}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple Content Example</h4>
      <div className="p-4 border rounded-lg">
        <Text variant="h3" weight="stronger" className="mb-3">
          Welcome Message
        </Text>
        <Showdown content="Welcome to our platform! We're excited to have you here. Get started by exploring our **dashboard** or check out the [documentation](https://docs.example.com) to learn more." />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Empty State Handling</h4>
      <div className="p-4 border rounded-lg">
        <Text variant="h3" weight="stronger" className="mb-3">
          No Content
        </Text>
        <Showdown content="" />
        <Text variant="subtext" theme="neutral" className="mt-2">
          The Markdown component gracefully handles empty content
        </Text>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use semantic markdown structure with proper heading hierarchy</li>
        <li>
          Include descriptive alt text for images when using HTML img tags
        </li>
        <li>Test external links to ensure they work correctly</li>
        <li>
          Use code blocks with language specification for syntax highlighting
        </li>
        <li>Organize content with lists and tables for better readability</li>
        <li>Consider collapsible sections for long documentation</li>
      </ul>
    </div>
  </div>
)

export const MermaidDiagrams = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Mermaid Diagrams</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The Markdown component supports Mermaid diagram rendering for creating
        flowcharts, sequence diagrams, and other visual representations directly
        from markdown. Diagrams are automatically rendered as interactive SVGs
        with proper styling and responsive behavior.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Flowchart Example</h4>
      <div className="p-4 border rounded-lg">
        <Showdown
          content={`## Application Components

\`\`\`mermaid
graph TD
  cluster[cluster<br/>0-tf-cluster]
  repository[repository<br/>1-tf-repository]
  certificate[certificate<br/>1-tf-certificate]
  img[img_inbox_zero<br/>0-img-ingress-zero]
  builder[builder<br/>2-tf-builder]

  cluster --> builder
  repository --> builder

  style builder fill:#D6B0FC,stroke:#8040BF,color:#000
  style cluster fill:#D6B0FC,stroke:#8040BF,color:#000
  style repository fill:#D6B0FC,stroke:#8040BF,color:#000
  style certificate fill:#D6B0FC,stroke:#8040BF,color:#000
  style img fill:#FCA04A,stroke:#FCA04A,color:#000
\`\`\`

Use **Manage > State** to view the full application state.`}
        />
      </div>
      <Text variant="subtext" theme="neutral">
        Mermaid diagrams support custom styling, multi-line labels with
        &lt;br/&gt; tags, and various diagram types
      </Text>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Sequence Diagram Example</h4>
      <div className="p-4 border rounded-lg">
        <Showdown
          content={`## API Authentication Flow

\`\`\`mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant Auth API
    participant Backend API

    User->>Frontend: Login Request
    Frontend->>Auth API: Authenticate
    Auth API-->>Frontend: JWT Token
    Frontend->>Backend API: API Request + Token
    Backend API-->>Frontend: Protected Data
    Frontend-->>User: Display Content
\`\`\`

This diagram shows the typical authentication flow in our application.`}
        />
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Mermaid Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>
          Supports flowcharts, sequence diagrams, class diagrams, and more
        </li>
        <li>Multi-line node labels using HTML &lt;br/&gt; tags</li>
        <li>Custom styling with fill colors, stroke colors, and CSS classes</li>
        <li>Responsive SVG output that scales with container</li>
        <li>Interactive elements with hover states and clickable nodes</li>
        <li>Automatic error handling with descriptive error messages</li>
      </ul>
    </div>
  </div>
)
