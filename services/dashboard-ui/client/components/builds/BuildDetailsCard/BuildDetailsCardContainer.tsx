import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { BuildDetailsCard, BuildDetailsCardSkeleton } from './BuildDetailsCard'
import type { ICard } from '@/components/common/Card'
import type { TBuild } from '@/types'

interface IBuildDetailsCardContainer extends Omit<ICard, 'children'> {
  build: TBuild
}

export const BuildDetailsCardContainer = ({ build, ...props }: IBuildDetailsCardContainer) => {
  const { install } = useInstall()
  const { org } = useOrg()

  return (
    <BuildDetailsCard
      build={build}
      orgId={org.id}
      installAppId={install.app_id}
      installAppConfigId={install.app_config_id}
      {...props}
    />
  )
}

export { BuildDetailsCardSkeleton }
