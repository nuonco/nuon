import { Text } from '../atoms/Text'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { installStateJSON } from '../../lib/fixtures/install-state'
import { CodeBlock } from './CodeBlock'

export default {
  title: 'lite/molecules/CodeBlock',
}

const SAMPLES: { language: string; filename: string; code: string }[] = [
  {
    language: 'bash',
    filename: 'sync.sh',
    code: `#!/usr/bin/env bash
set -euo pipefail

APP="\${1:-platform}"
nuon apps sync --dir "./apps/$APP"

for install in $(nuon installs list --output json | jq -r '.[].id'); do
  echo "deploying to $install"
  nuon installs deploy --install "$install" --wait
done`,
  },
  {
    language: 'json',
    filename: 'outputs.json',
    code: `{
  "endpoint": "https://api.example.internal",
  "replicas": 4,
  "enabled": true,
  "regions": ["us-west-2", "eu-west-1"],
  "limits": { "cpu": "2", "memory": "4Gi" }
}`,
  },
  {
    language: 'yaml',
    filename: 'deployment.yaml',
    code: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: apps
spec:
  replicas: 4
  template:
    spec:
      containers:
        - name: api
          image: ghcr.io/example/api:v2.1.0
          env:
            - name: LOG_LEVEL
              value: info`,
  },
  {
    language: 'hcl',
    filename: 'nuon.hcl',
    code: `app "platform" {
  name = "platform"

  sandbox "aws-eks" {
    version = "0.14.0"
  }

  input "region" {
    default     = "us-west-2"
    description = "Where the install runs"
  }
}`,
  },
  {
    language: 'terraform',
    filename: 'cluster.tf',
    code: `resource "aws_eks_cluster" "this" {
  name     = var.cluster_name
  role_arn = aws_iam_role.cluster.arn
  version  = "1.31"

  vpc_config {
    subnet_ids              = var.subnet_ids
    endpoint_private_access = true
  }
}

output "cluster_endpoint" {
  value = aws_eks_cluster.this.endpoint
}`,
  },
  {
    language: 'toml',
    filename: 'runner.toml',
    code: `[runner]
name = "primary"
concurrency = 4

[runner.limits]
cpu = "2"
memory = "4Gi"

[[runner.targets]]
region = "us-west-2"
enabled = true`,
  },
  {
    language: 'markdown',
    filename: 'README.md',
    code: `# Platform

Install the app with \`nuon apps sync\`, then create an install.

## Requirements

- An AWS account
- A VPC with two private subnets

> Provisioning takes about fifteen minutes.`,
  },
  {
    language: 'dockerfile',
    filename: 'Dockerfile',
    code: `FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api

FROM gcr.io/distroless/static
COPY --from=build /out/api /api
ENTRYPOINT ["/api"]`,
  },
  {
    language: 'rego',
    filename: 'authz.rego',
    code: `package nuon.authz

import rego.v1

default allow := false

allow if {
  input.method == "GET"
  input.role in {"admin", "viewer"}
}

deny contains msg if {
  input.method == "DELETE"
  not input.confirmed
  msg := "destructive calls must be confirmed"
}`,
  },
  {
    language: 'mermaid',
    filename: 'flow.mmd',
    code: `graph TD
  sync[Sync config] --> build[Build components]
  build --> plan[Plan changes]
  plan --> approve{Approved?}
  approve -->|yes| deploy[Deploy]
  approve -->|no| stop[Stop]`,
  },
  {
    language: 'text',
    filename: 'notes.txt',
    code: `Plain text gets no grammar and no colour.
It is also what an unrecognised language falls back to.`,
  },
]

export const Overview = () => (
  <ComponentDocs
    name="CodeBlock"
    tier="molecule"
    summary="A block of syntax-highlighted code, at any size."
    use={[
      'Show a manifest, a plan, a command or a config file.',
      'Render install state and other large JSON documents.',
      'Give someone something they can copy and run.',
    ]}
    avoid={[
      'Do not use it for a fragment inside a sentence. That is Code.',
      'Do not use it as an editor. It is read-only.',
      'Do not use it to render a diff. That is Diff.',
    ]}
    rules={[
      'value is a required string. Never infer the content from children.',
      'Pass the language you know it is; an unknown one renders as plain text and warns in development.',
      'Above a few hundred lines it switches to a virtualized scroll region and grows a search field, because browser find stops working once rows leave the DOM.',
      'In the search field, Down and Up step through matches and wrap around. Enter and Shift+Enter do the same.',
      'Colours come from the --syntax-* tokens, so a theme change needs no work here.',
    ]}
    props={[
      { name: 'value', type: 'string', description: 'The code. Required.' },
      {
        name: 'language',
        type: 'string',
        description:
          'Language or alias. Unknown values fall back to plain text.',
      },
      {
        name: 'filename',
        type: 'string',
        description: 'Shows the file header. Omitted means no header.',
      },
      {
        name: 'defaultWrap',
        type: 'boolean',
        default: 'false',
        description:
          'Starting state of the wrap toggle. The block owns it after that.',
      },
      {
        name: 'copy',
        type: 'boolean',
        default: 'false',
        description: 'Show a copy button.',
      },
      {
        name: 'lineNumbers',
        type: 'boolean',
        description: 'Defaults to on for anything longer than a line.',
      },
      {
        name: 'maxHeight',
        type: 'number',
        default: '480',
        description: 'Height of the scroll region, virtualized mode only.',
      },
    ]}
    sections={[
      {
        heading: 'Why it does not fall over',
        body: 'Highlighting means running a TextMate grammar over every line, and the old dashboard did that on the main thread while building a DOM node per token — which is what froze the page on install state. Here tokenizing happens in a worker pool and only the visible rows are ever in the DOM, so a document with thousands of lines costs about the same as one with fifty.',
      },
      {
        heading: 'Searching, and the absence of folding',
        body: 'Virtualized rows are not in the DOM, so browser find cannot see them. The block grows its own search field instead: it matches against the text, scrolls to the hit and marks the line. There is no tree folding for JSON — the renderer has none, and rather than hand-build it we leaned on search. If people ask for folding, it comes back as its own piece of work.',
      },
      {
        heading: 'Languages',
        body: 'Shell, JSON, YAML, HCL, Terraform, TOML, Markdown, Dockerfile, Mermaid, Rego and plain text, with the aliases call sites actually pass. Rego is not in Shiki, so its grammar is vendored from the OPA extension.',
      },
    ]}
  />
)

export const Languages = () => (
  <div className="flex max-w-3xl flex-col gap-6 p-8">
    <Text variant="caption" color="tertiary">
      Every language the block registers. Rego is not in Shiki, so its grammar
      is vendored from the OPA extension; Mermaid only lands here as a fallback
      when a real diagram cannot render; text is the fallback for anything
      unrecognised.
    </Text>
    {SAMPLES.map(({ language, filename, code }) => (
      <div key={language} className="flex flex-col gap-1.5">
        <Text variant="label" color="tertiary">
          {language}
        </Text>
        <CodeBlock language={language} filename={filename} value={code} copy />
      </div>
    ))}
  </div>
)

export const WithAndWithoutHeader = () => {
  const yaml = SAMPLES.find((sample) => sample.language === 'yaml')!.code

  return (
    <div className="flex max-w-2xl flex-col gap-6 p-8">
      <div className="flex flex-col gap-1.5">
        <Text variant="label" color="tertiary">
          With a filename
        </Text>
        <CodeBlock
          language="yaml"
          filename="deployment.yaml"
          value={yaml}
          copy
        />
      </div>
      <div className="flex flex-col gap-1.5">
        <Text variant="label" color="tertiary">
          Without one
        </Text>
        <CodeBlock language="yaml" value={yaml} copy />
      </div>
    </div>
  )
}

export const UnknownLanguage = () => (
  <div className="flex max-w-2xl flex-col gap-3 p-8">
    <Text variant="caption" color="tertiary">
      An unrecognised language renders as plain text and warns in development
      rather than throwing.
    </Text>
    <CodeBlock language="brainfuck" value={'++++[>++++<-]>.'} />
  </div>
)

export const LongLines = () => (
  <div className="flex max-w-xl flex-col gap-3 p-8">
    <Text variant="caption" color="tertiary">
      A short block scrolls sideways and takes its wrapping from defaultWrap.
      The toggle lives in the toolbar, which only appears once a block is big
      enough to be virtualized.
    </Text>
    <CodeBlock
      language="bash"
      value={
        'export NUON_API_TOKEN="a-very-long-token-value-that-keeps-going-well-past-the-edge-of-the-block-and-then-some-more-for-good-measure-because-real-tokens-are-long"'
      }
    />
    <CodeBlock
      defaultWrap
      language="bash"
      value={
        'export NUON_API_TOKEN="a-very-long-token-value-that-keeps-going-well-past-the-edge-of-the-block-and-then-some-more-for-good-measure-because-real-tokens-are-long"'
      }
    />
  </div>
)

export const SingleLine = () => (
  <div className="max-w-2xl p-8">
    <CodeBlock language="bash" value="nuon apps sync" copy />
  </div>
)

export const InstallState = () => {
  const value = installStateJSON()

  return (
    <div className="flex max-w-4xl flex-col gap-3 p-8">
      <Text variant="caption" color="tertiary">
        Shaped from a real production install state — {value.split('\n').length}{' '}
        lines, 30 components, 130 action workflows, and an embedded stack
        template that lands as a single line of about{' '}
        {Math.round(
          Math.max(...value.split('\n').map((line) => line.length)) / 1000
        )}
        k characters.
      </Text>
      <CodeBlock
        language="json"
        filename="install-state.json"
        value={value}
        copy
      />
    </div>
  )
}

export const InstallStateWrapped = () => (
  <div className="flex max-w-4xl flex-col gap-3 p-8">
    <Text variant="caption" color="tertiary">
      The same document starting with wrapping on, which turns that one enormous
      line into a very tall row. This is the case that breaks height estimation
      if anything is going to. The toolbar toggle switches it back.
    </Text>
    <CodeBlock
      defaultWrap
      language="json"
      filename="install-state.json"
      value={installStateJSON()}
    />
  </div>
)
