import { useInstall } from '@/hooks/use-install'
import { InstallComponentBuildCard, InstallComponentBuildCardSkeleton } from './InstallComponentBuildCard'
import type { TBuild } from '@/types'

export const InstallComponentBuildCardContainer = ({ build }: { build: TBuild }) => {
  const { install } = useInstall()

  return (
    <InstallComponentBuildCard
      build={build}
      orgId={install?.org_id}
      installAppId={install?.app_id}
      installAppConfigId={install?.app_config_id}
    />
  )
}

export { InstallComponentBuildCardSkeleton }
