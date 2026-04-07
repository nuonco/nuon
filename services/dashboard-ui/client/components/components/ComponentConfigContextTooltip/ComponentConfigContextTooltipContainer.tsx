import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getComponentConfig } from '@/lib'
import { ComponentConfigContextTooltip } from './ComponentConfigContextTooltip'

interface ComponentConfigContextTooltipContainerProps {
  componentId: string
  configId: string
  appId: string
  children?: React.ReactNode
}

export const ComponentConfigContextTooltipContainer = ({
  componentId,
  configId,
  appId,
  children,
}: ComponentConfigContextTooltipContainerProps) => {
  const { org } = useOrg()
  const { addModal } = useSurfaces()

  const {
    data: result,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['component-config', org?.id, appId, componentId, configId],
    queryFn: () => getComponentConfig({ orgId: org.id, appId, componentId, configId }),
    enabled: !!org?.id && !!appId && !!componentId && !!configId,
  })

  return (
    <ComponentConfigContextTooltip
      config={result ?? null}
      isLoading={isLoading}
      hasError={!!error}
      orgId={org.id}
      appId={appId}
      addModal={addModal}
    >
      {children}
    </ComponentConfigContextTooltip>
  )
}
