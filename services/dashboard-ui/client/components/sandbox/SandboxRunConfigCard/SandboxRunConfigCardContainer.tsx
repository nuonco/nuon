import { useInstall } from '@/hooks/use-install'
import type { TSandboxConfig } from '@/types'
import { SandboxRunConfigCard } from './SandboxRunConfigCard'

export const SandboxRunConfigCardContainer = ({
  config,
}: {
  config: TSandboxConfig
}) => {
  const { install } = useInstall()

  return (
    <SandboxRunConfigCard
      config={config}
      configHref={`/${install?.org_id}/apps/${install?.app_id}/configs/${install?.app_config_id}/sandbox`}
    />
  )
}
