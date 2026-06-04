export default {
  title: 'Sandbox/SandboxConfigCard',
}

import { SandboxConfigCard, SandboxConfigCardSkeleton } from './SandboxConfigCard'

const mockTerraformConfig = {
  id: 'config-1',
  terraform_version: '1.5.0',
  aws_region_type: 'us-east-1',
  drift_schedule: '0 2 * * *',
  env_vars: { API_KEY: 'secret' },
} as any

const mockPulumiConfig = {
  id: 'config-2',
  type: 'pulumi',
  runtime: 'go',
  pulumi_version: '3.100.0',
  pulumi_config: { 'gcp:project': 'my-project', 'gcp:region': 'us-central1' },
  drift_schedule: '0 2 * * *',
  env_vars: { API_KEY: 'secret' },
  public_git_vcs_config: {
    repo: 'https://github.com/my-org/pulumi-sandbox',
    branch: 'main',
    directory: '/',
  },
} as any

export const Default = () => (
  <div className="max-w-2xl p-4">
    <SandboxConfigCard
      config={mockTerraformConfig}
      onViewEnvVars={() => {}}
    />
  </div>
)

export const Pulumi = () => (
  <div className="max-w-2xl p-4">
    <SandboxConfigCard
      config={mockPulumiConfig}
      onViewEnvVars={() => {}}
      onViewPulumiConfig={() => {}}
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-4">
    <SandboxConfigCardSkeleton />
  </div>
)
