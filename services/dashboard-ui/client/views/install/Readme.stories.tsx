export default {
  title: 'Views/Install/Readme',
}

import { ReadmeWarnings } from '@/components/installs/ReadmeWarnings'

const warnings = [
  'unable to execute template: {{ .nuon.sandbox.outputs.nuon_dns.public_domain.name }}',
  'unable to execute template: {{ .nuon.install_stack.quick_link_url }}',
  'unable to execute template: {{ .nuon.install_stack.template_url }}',
  'unable to execute template: {{ .nuon.install_stack.template_url}}',
  'unable to execute template: {{ .nuon.install_stack.outputs.region }}',
]

const PlaceholderReadme = () => (
  <div className="prose dark:prose-invert max-w-none">
    <h1>My Application</h1>
    <p>This application deploys a Kubernetes cluster with the following components:</p>
    <ul>
      <li><strong>API Gateway</strong> — Routes traffic to backend services</li>
      <li><strong>Worker Pool</strong> — Processes background jobs</li>
      <li><strong>Database</strong> — PostgreSQL with automated backups</li>
    </ul>
  </div>
)

export const MultipleWarnings = () => (
  <div className="flex flex-col gap-4 max-w-4xl">
    <ReadmeWarnings warnings={warnings} />
    <PlaceholderReadme />
  </div>
)

export const SingleWarning = () => (
  <div className="flex flex-col gap-4 max-w-4xl">
    <ReadmeWarnings warnings={[warnings[0]]} />
    <PlaceholderReadme />
  </div>
)

export const NoWarnings = () => (
  <div className="flex flex-col gap-4 max-w-4xl">
    <PlaceholderReadme />
  </div>
)
