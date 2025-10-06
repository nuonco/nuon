import { CodeBlock } from './CodeBlock'

const code = `
{
  "key": "value",
  "number": 42,
  "nested": {
    "boolean": true
  }
}
`

const diffCode = `
- "boolean": true
+ "boolean": false
`

export const Default = () => <CodeBlock language="json">{code}</CodeBlock>

export const WithLineNumbers = () => (
  <CodeBlock language="json" showLineNumbers>
    {code}
  </CodeBlock>
)

export const Diff = () => (
  <CodeBlock language="yaml" isDiff>
    {diffCode}
  </CodeBlock>
)
