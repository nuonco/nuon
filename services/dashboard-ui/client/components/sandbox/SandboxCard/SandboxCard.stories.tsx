export default {
  title: 'Sandbox/SandboxCard',
}

import { SandboxCard } from './SandboxCard'

export const Default = () => (
  <SandboxCard
    status="active"
    href="/org-123/installs/inst-456/sandbox"
  />
)

export const Provisioning = () => (
  <SandboxCard
    status="provisioning"
    href="/org-123/installs/inst-456/sandbox"
  />
)

export const Error = () => (
  <SandboxCard
    status="error"
    href="/org-123/installs/inst-456/sandbox"
  />
)

export const NoSandbox = () => (
  <SandboxCard error="No sandbox found" />
)

export const NoLink = () => (
  <SandboxCard status="active" />
)

export const Loading = () => <SandboxCard isLoading />
