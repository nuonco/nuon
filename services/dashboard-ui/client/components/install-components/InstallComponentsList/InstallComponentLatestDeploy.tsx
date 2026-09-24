import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { HealthTimeline } from '@/components/install-health/HealthTimeline'
import {
  LatestDeployCard,
  type TLatestDeployCardVariant,
} from '@/components/install-components/LatestDeployCard'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallComponent } from '@/lib'

export const InstallComponentLatestDeploy = ({
  componentId,
  heading,
  showBuild = true,
  showHealth = false,
  variant = 'deploy',
}: {
  componentId: string
  heading?: string
  showBuild?: boolean
  showHealth?: boolean
  variant?: TLatestDeployCardVariant
}) => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: installComponent, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-component', org?.id, install?.id, componentId],
    queryFn: () =>
      getInstallComponent({
        orgId: org.id,
        installId: install.id,
        componentId,
      }),
    refetchInterval: 20000,
    enabled: !!org?.id && !!install?.id && !!componentId,
  })

  const latestDeploy = installComponent?.install_deploys?.[0]
  const build = showBuild ? latestDeploy?.component_build : undefined
  const title = heading ?? (variant === 'sync' ? 'Latest sync' : 'Latest deploy')

  return (
    <>
      <div className="flex flex-col gap-2">
        <Text variant="subtext" weight="strong" theme="neutral">
          {title}
        </Text>
        <LatestDeployCard
          flush
          deploy={latestDeploy}
          isLoading={isLoading}
          variant={variant}
          build={build}
          buildHref={
            build?.id && install?.app_id
              ? `/${org?.id}/apps/${install.app_id}/components/${componentId}/builds/${build.id}`
              : undefined
          }
          href={
            latestDeploy?.id
              ? `/${org?.id}/installs/${install?.id}/components/${componentId}/deploys/${latestDeploy.id}`
              : undefined
          }
        />
      </div>

      {showHealth && latestDeploy && installComponent?.id ? (
        <div className="flex flex-col gap-2 border-t pt-4">
          <Text variant="subtext" weight="strong" theme="neutral">
            Health
          </Text>
          <HealthTimeline installComponentId={installComponent.id} />
        </div>
      ) : null}
    </>
  )
}
