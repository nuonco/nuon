import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getAppConfig } from '@/lib'
import type { TAppConfig } from '@/types'
import { SandboxConfigContextTooltip } from './SandboxConfigContextTooltip'

interface SandboxConfigContextTooltipContainerProps {
  appConfigId: string
  appId: string
  children?: React.ReactNode
}

export const SandboxConfigContextTooltipContainer = ({
  appConfigId,
  appId,
  children,
}: SandboxConfigContextTooltipContainerProps) => {
  const { org } = useOrg()
  const { addModal } = useSurfaces()

  const {
    data: appConfig,
    isLoading,
    error,
  } = useQuery<TAppConfig>({
    queryKey: ['app-config', org?.id, appId, appConfigId],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId,
        appConfigId,
        recurse: true,
      }),
    enabled: !!org?.id && !!appId && !!appConfigId,
  })

  return (
    <SandboxConfigContextTooltip
      appConfigId={appConfigId}
      orgId={org.id}
      appId={appId}
      config={appConfig?.sandbox}
      isLoading={isLoading}
      error={error}
      addModal={addModal}
    >
      {children}
    </SandboxConfigContextTooltip>
  )
}
