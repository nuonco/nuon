import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { LatestDeployCard } from '@/components/install-components/LatestDeployCard'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallComponent } from '@/lib'

export const InstallComponentLatestDeploy = ({
  componentId,
}: {
  componentId: string
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

  return (
    <LatestDeployCard
      deploy={latestDeploy}
      isLoading={isLoading}
      href={
        latestDeploy?.id
          ? `/${org?.id}/installs/${install?.id}/components/${componentId}/deploys/${latestDeploy.id}`
          : undefined
      }
    />
  )
}
