export default {
  title: 'Sandbox/SandboxConfigCard',
}

import { SandboxConfigCard, SandboxConfigCardSkeleton } from './SandboxConfigCard'

const mockConfig = {
  id: 'config-1',
  terraform_version: '1.5.0',
  aws_region_type: 'us-east-1',
  drift_schedule: '0 2 * * *',
  env_vars: { API_KEY: 'secret' },
} as any

export const Default = () => (
  <div className="max-w-2xl p-4">
    <SandboxConfigCard
      config={mockConfig}
      onViewEnvVars={() => {}}
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-4">
    <SandboxConfigCardSkeleton />
  </div>
)
