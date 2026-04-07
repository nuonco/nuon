import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getAppConfigGraph } from '@/lib'
import { ComponentsGraphInline } from './ComponentsGraphRenderer'

export const ComponentsGraphInlineContainer = ({
  appId,
  configId,
}: {
  appId: string
  configId: string
}) => {
  const { org } = useOrg()

  const { data, error, isLoading } = useQuery({
    queryKey: ['app-config-graph', org?.id, appId, configId],
    queryFn: () => getAppConfigGraph({ orgId: org.id, appId, appConfigId: configId }),
    enabled: !!org?.id,
  })

  return (
    <ComponentsGraphInline
      dotGraph={data}
      error={error}
      isLoading={isLoading}
    />
  )
}
