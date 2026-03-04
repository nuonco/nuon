import { useState } from 'react'
import { Editor } from './Editor'
import { Text } from './Text'
import { CodeBlock } from './CodeBlock'

export const Basic = () => {
  const [code, setCode] = useState(`function hello() {
  console.log("Hello, world!");
  return true;
}`)

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Basic Editor</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          The Editor component provides a code editing experience with syntax
          highlighting. It captures input as a string for use in forms.
        </p>
      </div>

      <div className="space-y-4">
        <Editor
          value={code}
          onChange={setCode}
          language="javascript"
          placeholder="Enter your code here..."
        />

        <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
          <Text variant="label" weight="strong">
            Current Value:
          </Text>
          <Text variant="base" className="font-mono text-xs mt-2">
            {code.length} characters
          </Text>
        </div>
      </div>
    </div>
  )
}

export const Languages = () => {
  const codeExamples = {
    javascript: `function greet(name) {
  return \`Hello, \${name}!\`;
}`,
    typescript: `interface User {
  id: string;
  name: string;
  email: string;
}

function getUser(id: string): User {
  return { id, name: "John", email: "john@example.com" };
}`,
    python: `def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

print(fibonacci(10))`,
    json: `{
  "name": "nuon",
  "version": "1.0.0",
  "description": "BYOC platform",
  "main": "index.js"
}`,
    yaml: `version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
    environment:
      - NODE_ENV=production`,
    bash: `#!/bin/bash

echo "Starting deployment..."
npm install
npm run build
npm run deploy

echo "Deployment complete!"`,
  }

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Language Support</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          The{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
            language
          </code>{' '}
          prop controls syntax highlighting. Supports JavaScript, TypeScript,
          Python, JSON, YAML, Bash, and more.
        </p>
      </div>

      <div className="space-y-6">
        {Object.entries(codeExamples).map(([lang, code]) => (
          <div key={lang} className="space-y-2">
            <Text variant="label" weight="strong" className="capitalize">
              {lang}
            </Text>
            <Editor
              value={code}
              language={lang as any}
              minHeight={150}
              readOnly
            />
          </div>
        ))}
      </div>
    </div>
  )
}

export const Sizes = () => {
  const sampleCode = `const message = "Hello, world!";
console.log(message);`

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Height Controls</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Use{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
            minHeight
          </code>{' '}
          and{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
            maxHeight
          </code>{' '}
          props to control the editor dimensions.
        </p>
      </div>

      <div className="space-y-6">
        <div className="space-y-2">
          <Text variant="label" weight="strong">
            Small (minHeight: 100px)
          </Text>
          <Editor
            value={sampleCode}
            language="javascript"
            minHeight={100}
            maxHeight={100}
          />
        </div>

        <div className="space-y-2">
          <Text variant="label" weight="strong">
            Medium (minHeight: 200px, default)
          </Text>
          <Editor
            value={sampleCode}
            language="javascript"
            minHeight={200}
            maxHeight={200}
          />
        </div>

        <div className="space-y-2">
          <Text variant="label" weight="strong">
            Large (minHeight: 400px)
          </Text>
          <Editor
            value={sampleCode}
            language="javascript"
            minHeight={400}
            maxHeight={400}
          />
        </div>
      </div>
    </div>
  )
}

export const States = () => {
  const [editableCode, setEditableCode] = useState(`// This editor is editable
const value = "You can edit this";`)

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Editor States</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          The editor supports different states including disabled, read-only,
          and placeholder text.
        </p>
      </div>

      <div className="space-y-6">
        <div className="space-y-2">
          <Text variant="label" weight="strong">
            Editable (default)
          </Text>
          <Editor
            value={editableCode}
            onChange={setEditableCode}
            language="javascript"
            minHeight={120}
          />
        </div>

        <div className="space-y-2">
          <Text variant="label" weight="strong">
            Read-only
          </Text>
          <Editor
            value={`// This editor is read-only
const value = "You cannot edit this";`}
            language="javascript"
            readOnly
            minHeight={120}
          />
        </div>

        <div className="space-y-2">
          <Text variant="label" weight="strong">
            Disabled
          </Text>
          <Editor
            value={`// This editor is disabled
const value = "Disabled state";`}
            language="javascript"
            disabled
            minHeight={120}
          />
        </div>

        <div className="space-y-2">
          <Text variant="label" weight="strong">
            Empty with Placeholder
          </Text>
          <Editor
            value=""
            placeholder="Enter your code here..."
            language="javascript"
            minHeight={120}
          />
        </div>
      </div>
    </div>
  )
}

export const FormIntegration = () => {
  const [formData, setFormData] = useState({
    script: `#!/bin/bash
echo "Hello from the form!"`,
    config: `{
  "enabled": true,
  "timeout": 3000
}`,
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    alert(
      `Form submitted!\n\nScript:\n${formData.script}\n\nConfig:\n${formData.config}`
    )
  }

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Form Integration</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          The Editor component works seamlessly with forms. Use the{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
            name
          </code>{' '}
          prop for form field identification and{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
            onChange
          </code>{' '}
          to capture the value as a string.
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="space-y-2">
          <label htmlFor="script" className="block">
            <Text variant="label" weight="strong">
              Deployment Script
            </Text>
          </label>
          <Editor
            name="script"
            value={formData.script}
            onChange={(value) => setFormData({ ...formData, script: value })}
            language="bash"
            minHeight={150}
            placeholder="Enter your bash script..."
          />
        </div>

        <div className="space-y-2">
          <label htmlFor="config" className="block">
            <Text variant="label" weight="strong">
              Configuration (JSON)
            </Text>
          </label>
          <Editor
            name="config"
            value={formData.config}
            onChange={(value) => setFormData({ ...formData, config: value })}
            language="json"
            minHeight={150}
            placeholder="Enter your JSON config..."
          />
        </div>

        <button
          type="submit"
          className="px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700"
        >
          Submit Form
        </button>
      </form>

      <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-md space-y-3">
        <Text variant="label" weight="strong">
          Form State:
        </Text>
        <div className="space-y-2">
          <Text variant="base" className="text-xs">
            Script: {formData.script.length} characters
          </Text>
          <Text variant="base" className="text-xs">
            Config: {formData.config.length} characters
          </Text>
        </div>
      </div>
    </div>
  )
}

export const WithPreview = () => {
  const [code, setCode] = useState(`function calculateSum(a, b) {
  // Add two numbers together
  const result = a + b;
  return result;
}

console.log(calculateSum(5, 3));`)

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Editor with Preview</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Combine the Editor with the CodeBlock component to show a formatted
          preview.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="space-y-2">
          <Text variant="label" weight="strong">
            Editor (Editable)
          </Text>
          <Editor
            value={code}
            onChange={setCode}
            language="javascript"
            minHeight={300}
            maxHeight={300}
          />
        </div>

        <div className="space-y-2">
          <Text variant="label" weight="strong">
            Preview (Read-only)
          </Text>
          <CodeBlock language="javascript">{code}</CodeBlock>
        </div>
      </div>
    </div>
  )
}
