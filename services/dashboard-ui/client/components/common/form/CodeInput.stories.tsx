import { useState } from 'react'
import { CodeInput } from './CodeInput'

export default {
  title: 'Common/Forms/CodeInput',
}

export const Languages = () => (
  <div className="space-y-6">
    <div className="space-y-6">
      <div>
        <h4 className="text-sm font-medium mb-3">JSON configuration</h4>
        <CodeInput
          language="json"
          defaultValue={`{
  "name": "example-app",
  "version": "1.0.0",
  "dependencies": {
    "react": "^18.0.0"
  }
}`}
          labelProps={{ labelText: 'package.json' }}
          helperText="Enter valid JSON configuration"
        />
      </div>

      <div>
        <h4 className="text-sm font-medium mb-3">Shell script</h4>
        <CodeInput
          language="shell"
          defaultValue={`#!/bin/bash

echo 'Starting deployment...'
npm install
npm run build
echo 'Deployment complete!'`}
          labelProps={{ labelText: 'deploy.sh' }}
          helperText="Bash shell script with syntax highlighting"
        />
      </div>
    </div>
  </div>
)

export const Sizes = () => (
  <div className="space-y-6">
    <CodeInput
      size="sm"
      language="json"
      defaultValue='{"size": "small", "compact": true}'
      labelProps={{ labelText: 'Small size' }}
      helperText="Compact editor for shorter code snippets"
      minHeight={80}
    />
    <CodeInput
      size="md"
      language="json"
      defaultValue={`{
  "size": "medium",
  "default": true,
  "balanced": "readability and space"
}`}
      labelProps={{ labelText: 'Medium size (default)' }}
      helperText="Standard size for most use cases"
    />
    <CodeInput
      size="lg"
      language="json"
      defaultValue={`{
  "size": "large",
  "spacious": true,
  "readability": "enhanced",
  "use_case": "complex configurations"
}`}
      labelProps={{ labelText: 'Large size' }}
      helperText="Larger editor for complex code and better readability"
      minHeight={160}
    />
  </div>
)

export const WithLabelsAndText = () => (
  <div className="space-y-6">
    <CodeInput
      language="json"
      defaultValue={`{"no": "label"}`}
    />

    <CodeInput
      language="json"
      defaultValue={`{
  "with": "label"
}`}
      labelProps={{ labelText: 'Configuration file' }}
    />

    <CodeInput
      language="json"
      defaultValue={`{
  "with": "helper text"
}`}
      labelProps={{ labelText: 'API configuration' }}
      helperText="Enter your API endpoints and authentication settings"
    />
  </div>
)

export const States = () => (
  <div className="space-y-6">
    <div>
      <h4 className="text-sm font-medium mb-3">Normal</h4>
      <CodeInput
        language="json"
        defaultValue={`{
  "status": "normal",
  "editable": true
}`}
        labelProps={{ labelText: 'Editable configuration' }}
        helperText="This configuration can be modified"
      />
    </div>

    <div>
      <h4 className="text-sm font-medium mb-3">Error</h4>
      <CodeInput
        language="json"
        defaultValue={`{
  "invalid": json syntax,
  "missing": quotes
}`}
        labelProps={{ labelText: 'Invalid configuration' }}
        error
        errorMessage="Invalid JSON syntax. Check for missing quotes and commas."
      />
    </div>

    <div>
      <h4 className="text-sm font-medium mb-3">Disabled</h4>
      <CodeInput
        language="json"
        defaultValue={`{
  "readonly": true,
  "system": "generated",
  "editable": false
}`}
        labelProps={{ labelText: 'Read-only configuration' }}
        disabled
        helperText="This configuration is managed by the system"
      />
    </div>
  </div>
)

export const FormIntegration = () => {
  const [formData, setFormData] = useState({
    configContent: `{
  "port": 3000,
  "host": "localhost",
  "database": {
    "url": "postgresql://localhost:5432/myapp"
  }
}`,
    deployScript: `#!/bin/bash
echo "Deploying application..."
npm run build
npm run deploy`,
  })

  return (
    <div className="space-y-6">
      <form
        onSubmit={(e) => {
          e.preventDefault()
          alert(`Saved!\n\nConfig: ${formData.configContent.slice(0, 50)}...`)
        }}
        className="space-y-6 max-w-3xl"
      >
        <CodeInput
          id="config-content"
          name="configContent"
          language="json"
          value={formData.configContent}
          onChange={(e) =>
            setFormData((prev) => ({ ...prev, configContent: e.target.value }))
          }
          labelProps={{ labelText: 'Application configuration' }}
          helperText="Enter the JSON configuration for your application"
        />

        <CodeInput
          id="deploy-script"
          name="deployScript"
          language="bash"
          value={formData.deployScript}
          onChange={(e) =>
            setFormData((prev) => ({ ...prev, deployScript: e.target.value }))
          }
          labelProps={{ labelText: 'Deployment script' }}
          helperText="Optional shell script to run during deployment"
          minHeight={100}
        />

        <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
          <pre className="text-xs overflow-x-auto">
            {JSON.stringify(formData, null, 2)}
          </pre>
        </div>
      </form>
    </div>
  )
}
