/* eslint-disable react/no-unescaped-entities */
import { Markdown } from './Markdown'
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
        automatically opens external links in new tabs, and provides
        proper styling for all standard markdown elements.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple Markdown Example</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
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
        Markdown supports a full range of typography elements including
        headers, text formatting, lists, and more. All elements are
        properly styled and maintain consistency with the design system.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Headers and Text Formatting</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
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
        Markdown provides comprehensive support for both ordered and
        unordered lists, as well as tables with proper alignment and
        styling. Task lists are also supported for interactive content.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Lists and Tables Example</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
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

const codeExampleContent = `## Code Example

Here's how to use the Markdown component:

\`\`\`tsx
import { Markdown } from "./Markdown";

export function MyComponent() {
  const markdownContent = "# Hello\\n\\nThis is **markdown**!";
  
  return <Markdown content={markdownContent} />;
}
\`\`\`

The component automatically:
- Opens external links in new tabs
- Wraps tables in a container for styling
- Supports GitHub-flavored markdown features`

export const CodeBlocks = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Code Blocks and Syntax</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Markdown supports both inline code and code blocks with language
        specification for syntax highlighting. Code blocks are properly
        formatted with monospace fonts and appropriate styling.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Code Examples</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
          content={codeExampleContent}
        />
      </div>
    </div>
  </div>
)

export const InteractiveContent = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Interactive and Advanced Content</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Markdown supports advanced features like collapsible sections,
        HTML elements, and interactive content. This makes it suitable
        for documentation, help content, and rich text displays.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Advanced Features</h4>
      <div className="p-4 border rounded-lg">
        <Markdown
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
  return <Markdown content={content} />
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
        The Markdown component is commonly used for documentation,
        help content, user-generated content, and any interface where
        rich text formatting is needed.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Documentation Panel</h4>
      <div className="p-4 border rounded-lg space-y-4">
        <div className="flex justify-between items-center">
          <Text variant="h3" weight="stronger">API Documentation</Text>
          <Badge theme="info">v2.1</Badge>
        </div>
        <Markdown
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
          <Text variant="h3" weight="stronger">Troubleshooting Guide</Text>
          <Button variant="ghost" size="sm">Contact Support</Button>
        </div>
        <Markdown
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
        <Text variant="h3" weight="stronger" className="mb-3">Welcome Message</Text>
        <Markdown content="Welcome to our platform! We're excited to have you here. Get started by exploring our **dashboard** or check out the [documentation](https://docs.example.com) to learn more." />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Empty State Handling</h4>
      <div className="p-4 border rounded-lg">
        <Text variant="h3" weight="stronger" className="mb-3">No Content</Text>
        <Markdown content="" />
        <Text variant="subtext" theme="neutral" className="mt-2">
          The Markdown component gracefully handles empty content
        </Text>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use semantic markdown structure with proper heading hierarchy</li>
        <li>Include descriptive alt text for images when using HTML img tags</li>
        <li>Test external links to ensure they work correctly</li>
        <li>Use code blocks with language specification for syntax highlighting</li>
        <li>Organize content with lists and tables for better readability</li>
        <li>Consider collapsible sections for long documentation</li>
      </ul>
    </div>
  </div>
)
