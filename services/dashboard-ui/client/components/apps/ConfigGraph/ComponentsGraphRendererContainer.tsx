import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import type { TDependencyViewMode } from '@/components/common/DependencyViewToggle'
import { useOrg } from '@/hooks/use-org'
import { getAppConfigGraph } from '@/lib'
import type { TAPIError } from '@/types'
import { ComponentsGraphInline } from './ComponentsGraphRenderer'
import { ComponentsGraphTable } from './ComponentsGraphTable'
import { parseDotGraph } from './parse-dot'

export const ComponentsGraphInlineContainer = ({
  appId,
  configId,
  view = 'graph',
}: {
  appId: string
  configId: string
  view?: TDependencyViewMode
}) => {
  const { org } = useOrg()

  const { data, error, isLoading } = useQuery({
    queryKey: ['app-config-graph', org?.id, appId, configId],
    queryFn: () => getAppConfigGraph({ orgId: org.id, appId, appConfigId: configId }),
    enabled: !!org?.id,
  })

  const parsed = useMemo(
    () => (view === 'table' && data ? parseDotGraph(data) : undefined),
    [view, data],
  )

  if (view === 'table') {
    const apiError = error as TAPIError | null
    if (apiError?.error) {
      return <Banner theme="error">{apiError.error}</Banner>
    }
    return (
      <ComponentsGraphTable
        nodes={parsed?.nodes ?? []}
        edges={parsed?.edges ?? []}
        isLoading={isLoading}
      />
    )
  }

  return (
    <ComponentsGraphInline
      dotGraph={data}
      error={error}
      isLoading={isLoading}
    />
  )
}
