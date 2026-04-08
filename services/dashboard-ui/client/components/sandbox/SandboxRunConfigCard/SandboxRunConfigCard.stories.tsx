export default {
  title: 'Sandbox/SandboxRunConfigCard',
}

import { SandboxRunConfigCard } from './SandboxRunConfigCard'
import type { TSandboxConfig } from '@/types'

const mockConfig: TSandboxConfig = {
  terraform_version: '1.5.0',
  connected_github_vcs_config: {
    repo: 'nuonco/example-app',
    directory: 'terraform/',
    branch: 'main',
  },
} as TSandboxConfig

export const Default = () => (
  <SandboxRunConfigCard
    config={mockConfig}
    configHref="/org-1/apps/app-1/configs/config-1/sandbox"
  />
)
