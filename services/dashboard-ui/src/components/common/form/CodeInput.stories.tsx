/* eslint-disable react/no-unescaped-entities */
import { useState } from 'react'
import { CodeInput } from './CodeInput'

export default {
  title: 'Forms/CodeInput',
  component: CodeInput,
}

export const Languages = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Code Input Languages</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          language
        </code>{' '}
        prop enables syntax highlighting and appropriate editor features for
        different programming languages and markup formats. Default language is{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          json
        </code>{' '}
        if no language prop is provided.
      </p>
    </div>

    <div className="space-y-6">
      <div>
        <h4 className="text-sm font-medium mb-3">JSON Configuration</h4>
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
        <h4 className="text-sm font-medium mb-3">YAML Configuration</h4>
        <CodeInput
          language="yaml"
          defaultValue={`name: example-app
version: 1.0.0
ports:
  - 3000
  - 8080
environment:
  NODE_ENV: production`}
          labelProps={{ labelText: 'docker-compose.yml' }}
          helperText="YAML syntax with proper indentation"
        />
      </div>

      <div>
        <h4 className="text-sm font-medium mb-3">Shell Script</h4>
        <CodeInput
          language="shell"
          defaultValue={`#!/bin/bash

# Deploy application
echo 'Starting deployment...'
npm install
npm run build
echo 'Deployment complete!'`}
          labelProps={{ labelText: 'deploy.sh' }}
          helperText="Bash shell script with syntax highlighting"
        />
      </div>

      <div>
        <h4 className="text-sm font-medium mb-3">Terraform (HCL)</h4>
        <CodeInput
          language="hcl"
          defaultValue={`resource "aws_instance" "web" {
  ami           = "ami-0c55b159cbfafe1d0"
  instance_type = "t2.micro"
  
  tags = {
    Name = "HelloWorld"
  }
}`}
          labelProps={{ labelText: 'main.tf' }}
          helperText="HashiCorp Configuration Language"
        />
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 text-sm mt-6">
      <div><strong>json:</strong> JSON configuration files</div>
      <div><strong>yaml:</strong> YAML configuration and CI/CD files</div>
      <div><strong>shell:</strong> Bash and shell scripts</div>
      <div><strong>javascript:</strong> JavaScript source code</div>
      <div><strong>typescript:</strong> TypeScript source code</div>
      <div><strong>hcl:</strong> Terraform and HashiCorp configs</div>
      <div><strong>toml:</strong> TOML configuration files</div>
    </div>
  </div>
)

export const EmptyEditors = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Empty Code Editors</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Empty editors for each supported language. Use these to test typing
        and see how syntax highlighting works in real-time as you type.
      </p>
    </div>

    <div className="space-y-6">
      <div>
        <h4 className="text-sm font-medium mb-3">JSON Editor</h4>
        <CodeInput
          language="json"
          placeholder="Type JSON here..."
          labelProps={{ labelText: 'JSON Configuration' }}
          helperText='Try typing: { "name": "test" }'
        />
      </div>

      <div>
        <h4 className="text-sm font-medium mb-3">YAML Editor</h4>
        <CodeInput
          language="yaml"
          placeholder="Type YAML here..."
          labelProps={{ labelText: 'YAML Configuration' }}
          helperText="Try typing: name: test"
        />
      </div>

      <div>
        <h4 className="text-sm font-medium mb-3">JavaScript Editor</h4>
        <CodeInput
          language="javascript"
          placeholder="Type JavaScript here..."
          labelProps={{ labelText: 'JavaScript Code' }}
          helperText="Try typing: const name = 'test';"
        />
      </div>

      <div>
        <h4 className="text-sm font-medium mb-3">TypeScript Editor</h4>
        <CodeInput
          language="typescript"
          placeholder="Type TypeScript here..."
          labelProps={{ labelText: 'TypeScript Code' }}
          helperText="Try typing: interface User { name: string }"
        />
      </div>

      <div>
        <h4 className="text-sm font-medium mb-3">Shell Script Editor</h4>
        <CodeInput
          language="shell"
          placeholder="Type shell commands here..."
          labelProps={{ labelText: 'Shell Script' }}
          helperText="Try typing: #!/bin/bash"
        />
      </div>

      <div>
        <h4 className="text-sm font-medium mb-3">Terraform (HCL) Editor</h4>
        <CodeInput
          language="hcl"
          placeholder="Type Terraform configuration here..."
          labelProps={{ labelText: 'Terraform Configuration' }}
          helperText='Try typing: resource "aws_instance" "example" {}'
        />
      </div>

      <div>
        <h4 className="text-sm font-medium mb-3">TOML Editor</h4>
        <CodeInput
          language="toml"
          placeholder="Type TOML configuration here..."
          labelProps={{ labelText: 'TOML Configuration' }}
          helperText="Try typing: [package]"
        />
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Interactive Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Real-time syntax highlighting as you type</li>
        <li>Language-specific autocomplete and formatting</li>
        <li>Proper indentation and code folding</li>
        <li>Error detection for supported languages</li>
        <li>Keyboard shortcuts for common operations</li>
      </ul>
    </div>
  </div>
)

export const Sizes = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Code Input Sizes</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          size
        </code>{' '}
        prop controls the padding, font size, and overall dimensions of the code editor.
        Default size is{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          md
        </code>{' '}
        if no size prop is provided.
      </p>
    </div>

    <div className="space-y-6">
      <CodeInput
        size="sm"
        language="json"
        defaultValue='{"size": "small", "compact": true}'
        labelProps={{ labelText: 'Small Size' }}
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
        labelProps={{ labelText: 'Medium Size (Default)' }}
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
        labelProps={{ labelText: 'Large Size' }}
        helperText="Larger editor for complex code and better readability"
        minHeight={160}
      />
    </div>

    <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm mt-6">
      <div><strong>sm:</strong> 12px font, 12px padding, compact layout</div>
      <div><strong>md:</strong> 14px font, 16px padding, balanced (default)</div>
      <div><strong>lg:</strong> 16px font, 20px padding, spacious layout</div>
    </div>
  </div>
)

export const WithLabelsAndText = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Labels and Helper Text</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        CodeInput supports the same label and helper text patterns as other form
        components. Use{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          labelProps
        </code>{' '}
        for labels and{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          helperText
        </code>{' '}
        for additional context.
      </p>
    </div>

    <div className="space-y-6">
      <CodeInput
        language="yaml"
        defaultValue={`# Basic configuration
name: my-app
version: 1.0.0`}
        placeholder="Enter YAML configuration..."
      />

      <CodeInput
        language="json"
        defaultValue={`{
  "with": "label"
}`}
        labelProps={{
          labelText: 'Configuration File'
        }}
      />

      <CodeInput
        language="json"
        defaultValue={`{
  "with": "helper text"
}`}
        labelProps={{
          labelText: 'API Configuration'
        }}
        helperText="Enter your API endpoints and authentication settings"
      />

      <CodeInput
        language="json"
        defaultValue={`{
  "custom": "styling"
}`}
        labelProps={{
          labelText: 'Advanced Configuration',
          labelTextProps: { weight: 'strong', theme: 'default' }
        }}
        helperText="This configuration requires elevated permissions"
        helperTextProps={{ theme: 'warn' }}
      />
    </div>
  </div>
)

export const States = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Code Input States</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        CodeInput supports error states with custom error messages and disabled
        states for read-only content. Error states change the border color and
        display error messages below the editor.
      </p>
    </div>

    <div className="space-y-6">
      <div>
        <h4 className="text-sm font-medium mb-3">Normal State</h4>
        <CodeInput
          language="json"
          defaultValue={`{
  "status": "normal",
  "editable": true
}`}
          labelProps={{ labelText: 'Editable Configuration' }}
          helperText="This configuration can be modified"
        />
      </div>

      <div>
        <h4 className="text-sm font-medium mb-3">Error State</h4>
        <CodeInput
          language="json"
          defaultValue={`{
  "invalid": json syntax,
  "missing": quotes
}`}
          labelProps={{ labelText: 'Invalid Configuration' }}
          error
          errorMessage="Invalid JSON syntax. Check for missing quotes and commas."
        />
      </div>

      <div>
        <h4 className="text-sm font-medium mb-3">Disabled State</h4>
        <CodeInput
          language="json"
          defaultValue={`{
  "readonly": true,
  "system": "generated",
  "editable": false
}`}
          labelProps={{ labelText: 'Read-only Configuration' }}
          disabled
          helperText="This configuration is managed by the system"
        />
      </div>
    </div>
  </div>
)

export const RealWorldExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Real-World Examples</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Common use cases for CodeInput in application forms, showing practical
        examples of different configuration types you might encounter in a
        cloud platform or development environment.
      </p>
    </div>

    <div className="space-y-6">
      <CodeInput
        language="yaml"
        defaultValue={`version: '3.8'
services:
  web:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./html:/usr/share/nginx/html
  db:
    image: postgres:13
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: \${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:`}
        labelProps={{
          labelText: 'Docker Compose Configuration',
          labelTextProps: { weight: 'strong' }
        }}
        helperText="Define your application services and dependencies"
        minHeight={200}
      />
      
      <CodeInput
        language="json"
        defaultValue={`{
  "name": "@company/api-service",
  "version": "1.2.3",
  "description": "RESTful API service",
  "main": "dist/index.js",
  "scripts": {
    "start": "node dist/index.js",
    "dev": "ts-node-dev --respawn src/index.ts",
    "build": "tsc",
    "test": "jest"
  },
  "dependencies": {
    "express": "^4.18.0",
    "cors": "^2.8.5",
    "helmet": "^6.0.0"
  },
  "engines": {
    "node": ">=18.0.0"
  }
}`}
        labelProps={{
          labelText: 'Package Configuration',
          labelTextProps: { weight: 'strong' }
        }}
        helperText="Node.js package.json for your application"
        minHeight={180}
      />

      <CodeInput
        language="hcl"
        defaultValue={`variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "\${var.environment}-vpc"
    Environment = var.environment
  }
}

resource "aws_subnet" "public" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.\${count.index + 1}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  map_public_ip_on_launch = true

  tags = {
    Name = "\${var.environment}-public-subnet-\${count.index + 1}"
    Type = "public"
  }
}`}
        labelProps={{
          labelText: 'Infrastructure Configuration',
          labelTextProps: { weight: 'strong' }
        }}
        helperText="Terraform configuration for AWS infrastructure"
        minHeight={220}
      />
    </div>
  </div>
)

export const FormIntegration = () => {
  const [formData, setFormData] = useState({
    configName: 'my-application',
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
npm run deploy`
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    alert(`Configuration saved!\n\nName: ${formData.configName}\nConfig: ${formData.configContent.slice(0, 50)}...\nScript: ${formData.deployScript.slice(0, 50)}...`)
  }

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Form Integration</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          CodeInput integrates seamlessly with forms and controlled components.
          It supports standard form attributes like{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
            name
          </code>,{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
            required
          </code>, and{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
            onChange
          </code>{' '}
          handlers.
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6 max-w-3xl">
        <div>
          <label htmlFor="config-name" className="block text-sm font-medium mb-1">
            Configuration Name *
          </label>
          <input
            id="config-name"
            type="text"
            value={formData.configName}
            onChange={(e) => setFormData(prev => ({ ...prev, configName: e.target.value }))}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 bg-white dark:bg-gray-900"
            placeholder="Enter configuration name"
            required
          />
        </div>
        
        <CodeInput
          id="config-content"
          name="configContent"
          language="json"
          value={formData.configContent}
          onChange={(e) => setFormData(prev => ({ ...prev, configContent: e.target.value }))}
          labelProps={{
            labelText: 'Application Configuration *',
          }}
          helperText="Enter the JSON configuration for your application"
          required
        />

        <CodeInput
          id="deploy-script"
          name="deployScript"
          language="shell"
          value={formData.deployScript}
          onChange={(e) => setFormData(prev => ({ ...prev, deployScript: e.target.value }))}
          labelProps={{
            labelText: 'Deployment Script',
          }}
          helperText="Optional shell script to run during deployment"
          minHeight={100}
        />
        
        <div className="flex gap-3">
          <button
            type="submit"
            className="px-6 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            Save Configuration
          </button>
          <button
            type="button"
            onClick={() => setFormData({
              configName: '',
              configContent: `{
  
}`,
              deployScript: '#!/bin/bash\n'
            })}
            className="px-6 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-md hover:bg-gray-50 dark:hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-gray-500"
          >
            Reset
          </button>
        </div>

        <div className="text-sm text-gray-600 dark:text-gray-400 mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
          <strong>Form Data:</strong>
          <pre className="mt-2 text-xs overflow-x-auto">
            {JSON.stringify(formData, null, 2)}
          </pre>
        </div>
      </form>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Form Integration Features:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>Controlled and uncontrolled component support</li>
          <li>Standard form attributes (name, required, etc.)</li>
          <li>onChange event handling for real-time validation</li>
          <li>Form submission with proper value serialization</li>
          <li>Integration with form libraries like React Hook Form</li>
        </ul>
      </div>
    </div>
  )
}
