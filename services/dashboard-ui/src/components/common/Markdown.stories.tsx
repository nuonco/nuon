import { Markdown } from './Markdown'

export const Basic = () => (
  <Markdown
    content={`# Hello World

This is a basic markdown example with **bold** and *italic* text.`}
  />
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

export const ComplexMarkdown = () => (
  <Markdown content={complexMarkdownContent} />
)

export const SimpleContent = () => (
  <Markdown content="A simple paragraph with some **emphasis** and a [link](https://example.com)." />
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

export const CodeExample = () => <Markdown content={codeExampleContent} />

export const EmptyContent = () => <Markdown content="" />
