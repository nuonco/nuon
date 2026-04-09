export default {
  title: 'Install Components/ComponentCard',
}

import { ComponentCard } from './ComponentCard'

export const Default = () => (
  <ComponentCard
    name="networking"
    type="terraform_module"
    status="active"
    href="/org-123/installs/inst-456/components/comp-789"
  />
)

export const HelmChart = () => (
  <ComponentCard
    name="ingress-nginx"
    type="helm_chart"
    status="provisioning"
    href="/org-123/installs/inst-456/components/comp-abc"
  />
)

export const DockerBuild = () => (
  <ComponentCard
    name="api-server"
    type="docker_build"
    status="error"
    href="/org-123/installs/inst-456/components/comp-def"
  />
)

export const NoLink = () => (
  <ComponentCard name="database" type="terraform_module" status="active" />
)

export const Loading = () => <ComponentCard isLoading />

export const Error = () => <ComponentCard error="Failed to load component" />

export const NotFound = () => (
  <ComponentCard error='Component "missing-module" not found' />
)
