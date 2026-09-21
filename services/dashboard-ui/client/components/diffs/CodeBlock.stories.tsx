export default {
  title: 'Diffs/CodeBlock',
}

import { CodeBlock } from './CodeBlock'

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

const shellCode = `#!/bin/bash
set -euo pipefail

nuon app list --output json | jq -r '.[].id' | while read -r app; do
  nuon app get "$app"
done`

const regoCode = `package nuon.authz

default allow := false

allow if {
  input.method == "GET"
  startswith(input.path, "/v1/installs")
}`

const largeConfig = [
  'apiVersion: v1',
  'kind: ConfigMap',
  'metadata:',
  '  name: acme-settings',
  'data:',
  ...Array.from({ length: 400 }, (_, i) => `  key_${i}: "value-${i}"`),
].join('\n')

export const BasicUsage = () => (
  <div className="flex flex-col gap-4 p-4">
    <CodeBlock value={jsonCode} language="json" />
  </div>
)

export const LanguageSupport = () => (
  <div className="flex flex-col gap-4 p-4">
    <CodeBlock value={jsonCode} language="json" filename="package.json" />
    <CodeBlock value={yamlCode} language="yaml" filename="deployment.yaml" />
    <CodeBlock value={terraformCode} language="terraform" filename="main.tf" />
    <CodeBlock value={shellCode} language="shell" filename="sync.sh" />
    <CodeBlock value={regoCode} language="rego" filename="authz.rego" />
  </div>
)

export const WithLineNumbers = () => (
  <div className="flex flex-col gap-4 p-4">
    <CodeBlock value={yamlCode} language="yaml" lineNumbers />
  </div>
)

export const SingleLine = () => (
  <div className="flex flex-col gap-4 p-4">
    <CodeBlock value="nuon install create --app acme-api" language="shell" />
  </div>
)

export const WithCopy = () => (
  <div className="flex flex-col gap-4 p-4">
    <CodeBlock value={terraformCode} language="terraform" copy />
  </div>
)

export const Wrapped = () => (
  <div className="flex flex-col gap-4 p-4">
    <CodeBlock value={terraformCode} language="terraform" defaultWrap />
  </div>
)

export const VirtualizedWithSearch = () => (
  <div className="flex flex-col gap-4 p-4">
    <CodeBlock
      value={largeConfig}
      language="yaml"
      filename="settings.yaml"
      copy
    />
  </div>
)
