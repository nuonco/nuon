import { useInstall } from '@/hooks/use-install'
import { InstallComponentConfigCard, InstallComponentConfigCardSkeleton } from './InstallComponentConfigCard'
import type { TComponentConfig } from '@/types'

export const InstallComponentConfigCardContainer = ({ config }: { config: TComponentConfig }) => {
  const { install } = useInstall()

  return (
    <InstallComponentConfigCard
      config={config}
      orgId={install?.org_id}
      installAppId={install?.app_id}
      installAppConfigId={install?.app_config_id}
    />
  )
}

export { InstallComponentConfigCardSkeleton }
