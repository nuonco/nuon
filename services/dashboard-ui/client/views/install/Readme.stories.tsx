export default {
  title: 'Views/Install/Readme',
}

import { Expand } from '@/components/common/Expand'
import { Markdown } from '@/components/common/Markdown'
import { ReadmeWarnings } from '@/components/installs/ReadmeWarnings'

const warnings = [
  'unable to execute template: {{ .nuon.sandbox.outputs.nuon_dns.public_domain.name }}',
  'unable to execute template: {{ .nuon.install_stack.quick_link_url }}',
  'unable to execute template: {{ .nuon.install_stack.template_url }}',
  'unable to execute template: {{ .nuon.install_stack.template_url}}',
  'unable to execute template: {{ .nuon.install_stack.outputs.region }}',
]

const readmeMarkdown = `# My Application

This application deploys a Kubernetes cluster with the following components:

- **API Gateway** — Routes traffic to backend services
- **Worker Pool** — Processes background jobs
- **Database** — PostgreSQL with automated backups

## Access

Open your app at {{ .nuon.install_stack.quick_link_url }}

Region: {{ .nuon.install_stack.outputs.region }}

Public domain: {{ .nuon.sandbox.outputs.nuon_dns.public_domain.name }}

## Architecture

The platform is split across several layers, each provisioned independently so
they can scale on their own schedule.

### Networking

- VPC with public and private subnets across three availability zones
- NAT gateways for outbound traffic from private subnets
- Internal load balancer fronting the API gateway

### Compute

- Managed node groups for stateless services
- A dedicated node group for background workers
- Cluster autoscaler tuned for burst workloads

### Data

- Primary PostgreSQL instance with a hot standby
- Automated daily backups retained for 30 days
- Redis for caching and job queues

## Configuration

| Setting | Default | Description |
| --- | --- | --- |
| Region | {{ .nuon.install_stack.outputs.region }} | Deployment region |
| Domain | {{ .nuon.sandbox.outputs.nuon_dns.public_domain.name }} | Public domain |
| Instance size | m5.large | Default node size |
| Replicas | 3 | API gateway replicas |

## Runbook

1. Confirm the install has reached an active state.
2. Verify the quick link resolves: {{ .nuon.install_stack.quick_link_url }}
3. Check worker health in the dashboard.
4. Review recent deploy logs for template rendering warnings.

## Support

Reach out to your account team if any of the templated values above fail to
resolve after the install finishes provisioning.
`

export const WithWarnings = () => (
  <div className="flex flex-col gap-4 max-w-4xl">
    <ReadmeWarnings warnings={warnings} />
    <Expand
      id="incomplete-readme"
      heading="View incomplete README"
      className="border rounded-lg"
    >
      <div className="p-4 border-t max-h-[32rem] overflow-y-auto">
        <Markdown content={readmeMarkdown} mode="install" />
      </div>
    </Expand>
  </div>
)

export const WithWarningsExpanded = () => (
  <div className="flex flex-col gap-4 max-w-4xl">
    <ReadmeWarnings warnings={[warnings[0]]} />
    <Expand
      id="incomplete-readme-open"
      heading="View incomplete README"
      className="border rounded-lg"
      isOpen
    >
      <div className="p-4 border-t max-h-[32rem] overflow-y-auto">
        <Markdown content={readmeMarkdown} mode="install" />
      </div>
    </Expand>
  </div>
)

export const NoWarnings = () => (
  <div className="flex flex-col gap-4 max-w-4xl">
    <Markdown content={readmeMarkdown} mode="install" />
  </div>
)
