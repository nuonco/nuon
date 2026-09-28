export default {
  title: 'Features / Workflows / Sandbox run details / Sandbox run plan',
}

import { SandboxRunPlan } from './SandboxRunPlan'

export const Default = () => (
  <SandboxRunPlan plan={{ resource_changes: [] }} isLoading={false} />
)

export const Loading = () => <SandboxRunPlan plan={null} isLoading={true} />
