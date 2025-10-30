/* eslint-disable react/no-unescaped-entities */
import { CodeBlock } from './CodeBlock'
import { Text } from './Text'
import { ClickToCopy } from './ClickToCopy'

const jsonCode = `{
  "name": "nuon-dashboard",
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
  }
}`

const yamlCode = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: dashboard-ui
  labels:
    app: dashboard-ui
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
        image: nuon/dashboard-ui:latest
        ports:
        - containerPort: 3000`

const diffCode = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: dashboard-ui
- replicas: 2
+ replicas: 3
  selector:
    matchLabels:
      app: dashboard-ui
- image: nuon/dashboard-ui:v2.0.0
+ image: nuon/dashboard-ui:v2.1.0`

const terraformCode = `resource "aws_instance" "web" {
  ami           = "ami-0c02fb55956c7d316"
  instance_type = "t3.micro"
  
  vpc_security_group_ids = [aws_security_group.web.id]
  subnet_id              = aws_subnet.public.id
  
  user_data = <<-EOF
              #!/bin/bash
              yum update -y
              yum install -y httpd
              systemctl start httpd
              systemctl enable httpd
              EOF
              
  tags = {
    Name = "WebServer"
    Environment = "production"
  }
}`

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic CodeBlock Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        CodeBlock components provide syntax-highlighted code display with
        automatic theme switching, language detection, and advanced features
        like line numbers and diff highlighting. Built on react-syntax-highlighter
        with Prism.js for comprehensive language support.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">JSON Configuration</h4>
      <CodeBlock language="json">{jsonCode}</CodeBlock>
      <Text variant="subtext" theme="neutral">
        Automatic syntax highlighting with proper JSON formatting
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Automatic light/dark theme switching based on system preference</li>
        <li>Comprehensive language support (JSON, YAML, HCL, Shell, etc.)</li>
        <li>Scrollable content with maximum height constraints</li>
        <li>Custom styling with CSS variables for consistent theming</li>
      </ul>
    </div>
  </div>
)

export const LanguageSupport = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Language Support</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        CodeBlock supports major languages used in cloud infrastructure,
        configuration management, and development workflows. Each language
        gets appropriate syntax highlighting and formatting.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">YAML (Kubernetes)</h4>
      <CodeBlock language="yaml">{yamlCode}</CodeBlock>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">HCL (Terraform)</h4>
      <CodeBlock language="hcl">{terraformCode}</CodeBlock>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Shell Script</h4>
      <CodeBlock language="bash">
        {`#!/bin/bash

# Deploy application
echo "Starting deployment..."

# Build Docker image
docker build -t nuon/dashboard-ui:latest .

# Push to registry
docker push nuon/dashboard-ui:latest

# Update Kubernetes deployment
kubectl set image deployment/dashboard-ui dashboard-ui=nuon/dashboard-ui:latest

# Wait for rollout to complete
kubectl rollout status deployment/dashboard-ui

echo "Deployment completed successfully!"`}
      </CodeBlock>
    </div>

    <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-sm mt-6">
      <div><strong>json:</strong> Configuration files</div>
      <div><strong>yaml/yml:</strong> Kubernetes, Docker Compose</div>
      <div><strong>hcl:</strong> Terraform infrastructure</div>
      <div><strong>sh/bash:</strong> Shell scripts and commands</div>
      <div><strong>toml:</strong> Configuration files</div>
      <div><strong>markdown/md:</strong> Documentation</div>
    </div>
  </div>
)

export const WithLineNumbers = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Line Numbers</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Use the{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          showLineNumbers
        </code>{' '}
        prop to display line numbers alongside code content. This is useful
        for referencing specific lines in documentation or code reviews.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Configuration with Line Numbers</h4>
      <CodeBlock language="json" showLineNumbers>
        {jsonCode}
      </CodeBlock>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Script with Line Numbers</h4>
      <CodeBlock language="bash" showLineNumbers>
        {`#!/bin/bash

# Configuration
APP_NAME="dashboard-ui"
VERSION="2.1.0"
REGISTRY="nuon"

# Functions
build_image() {
  echo "Building Docker image..."
  docker build -t $REGISTRY/$APP_NAME:$VERSION .
}

deploy_app() {
  echo "Deploying application..."
  kubectl apply -f k8s/
}

# Main execution
build_image
deploy_app
echo "Deployment complete!"`}
      </CodeBlock>
    </div>
  </div>
)

export const DiffVisualization = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Diff Visualization</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          isDiff
        </code>{' '}
        prop enables diff highlighting with color-coded additions and deletions.
        Lines starting with + are highlighted in green, while lines starting
        with - are highlighted in red. Line numbers are automatically enabled.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Configuration Changes</h4>
      <CodeBlock language="yaml" isDiff>
        {diffCode}
      </CodeBlock>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Terraform Plan Output</h4>
      <CodeBlock language="hcl" isDiff>
        {`# aws_instance.web will be updated in-place
~ resource "aws_instance" "web" {
      id                    = "i-1234567890abcdef0"
-     instance_type         = "t3.micro"
+     instance_type         = "t3.small"
      # (Known after apply)
    }

# aws_security_group.web will be created
+ resource "aws_security_group" "web" {
+     arn                    = (known after apply)
+     description            = "Web server security group"
+     id                     = (known after apply)
+   }

Plan: 1 to add, 1 to change, 0 to destroy.`}
      </CodeBlock>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>Green lines (+):</strong> Additions, new resources, or
        increased values
      </div>
      <div>
        <strong>Red lines (-):</strong> Deletions, removed resources, or
        previous values
      </div>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        CodeBlock components are essential for technical documentation,
        configuration examples, and deployment workflows. Here are common
        patterns and recommended approaches.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Documentation Examples</h4>
      <div className="space-y-4 p-4 border rounded-lg">
        <div>
          <Text weight="strong">Environment Configuration</Text>
          <Text variant="subtext" theme="neutral" className="mt-1 mb-3">
            Create a .env.local file with the following configuration:
          </Text>
          <ClickToCopy>
            <CodeBlock language="bash">
              {`NUON_API_URL=https://api.nuon.co
NUON_API_TOKEN=your_api_token_here
NEXT_PUBLIC_ANALYTICS_ID=analytics_key
DATABASE_URL=postgresql://user:password@localhost:5432/nuon`}
            </CodeBlock>
          </ClickToCopy>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Deployment Configuration</h4>
      <div className="p-4 border rounded-lg bg-blue-50 dark:bg-blue-950/20">
        <Text weight="strong">Kubernetes Deployment</Text>
        <Text variant="subtext" theme="neutral" className="mt-1 mb-3">
          Apply this configuration to deploy the dashboard:
        </Text>
        <CodeBlock language="yaml" showLineNumbers>
          {`apiVersion: apps/v1
kind: Deployment
metadata:
  name: dashboard-ui
  namespace: nuon-system
spec:
  replicas: 3
  selector:
    matchLabels:
      app: dashboard-ui
  template:
    spec:
      containers:
      - name: dashboard-ui
        image: nuon/dashboard-ui:2.1.0
        env:
        - name: NODE_ENV
          value: "production"`}
        </CodeBlock>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Infrastructure Changes</h4>
      <div className="p-4 border rounded-lg bg-yellow-50 dark:bg-yellow-950/20">
        <Text weight="strong">Terraform Plan Review</Text>
        <Text variant="subtext" theme="neutral" className="mt-1 mb-3">
          Review the following changes before applying:
        </Text>
        <CodeBlock language="hcl" isDiff>
          {`# aws_instance.web will be replaced
-/+ resource "aws_instance" "web" {
~     ami                    = "ami-12345" -> "ami-67890"
      instance_type          = "t3.micro"
+     monitoring             = true
      # (Known after apply)
    }

Plan: 1 to add, 0 to change, 1 to destroy.`}
        </CodeBlock>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use appropriate language identifiers for proper syntax highlighting</li>
        <li>Enable line numbers for code that will be referenced or discussed</li>
        <li>Use diff mode for showing configuration changes and updates</li>
        <li>Combine with ClickToCopy for easy code snippet copying</li>
        <li>Keep code examples focused and relevant to the documentation context</li>
      </ul>
    </div>
  </div>
)
