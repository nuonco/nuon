import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getDeploy } from '@/lib'
import type { TDeploy, TInstallComponent } from '@/types'
import { InstallComponentHeader } from './InstallComponentHeader'

interface IInstallComponentHeaderContainer {
  initDeploy: TDeploy
  installComponent: TInstallComponent
  pollInterval?: number
  shouldPoll?: boolean
}

export const InstallComponentHeaderContainer = ({
  initDeploy,
  installComponent,
  pollInterval = 20000,
  shouldPoll = false,
}: IInstallComponentHeaderContainer) => {
  const { install } = useInstall()
  const { org } = useOrg()

  const { data: deploy } = useQuery<TDeploy>({
    queryKey: ['deploy', org?.id, install?.id, initDeploy?.id],
    queryFn: () =>
      getDeploy({
        orgId: org.id,
        installId: install.id,
        deployId: initDeploy?.id,
      }),
    initialData: initDeploy,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id && !!initDeploy?.id,
  })

  return (
    <InstallComponentHeader
      deploy={deploy ?? initDeploy}
      installComponent={installComponent}
      componentId={installComponent.component_id}
      deployId={initDeploy?.id}
    />
  )
}
