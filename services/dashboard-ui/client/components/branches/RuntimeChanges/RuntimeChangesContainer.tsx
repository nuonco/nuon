import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getBranchRunBuilds, getComponents } from '@/lib'
import type { TComponentType } from '@/types'
import { RuntimeChanges, type IRuntimeChangeRow } from './RuntimeChanges'

const IMAGE_TYPES: TComponentType[] = ['docker_build', 'external_image']

interface IRuntimeChangesContainer {
  branchId: string
  appBranchRunId: string
}

export const RuntimeChangesContainer = ({ branchId, appBranchRunId }: IRuntimeChangesContainer) => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: builds } = useQuery({
    queryKey: ['branch-run-builds', org?.id, app?.id, branchId, appBranchRunId],
    queryFn: () =>
      getBranchRunBuilds({ orgId: org!.id, appId: app!.id, branchId, runId: appBranchRunId }),
    enabled: !!org?.id && !!app?.id && !!branchId && !!appBranchRunId,
    refetchInterval: 10000,
  })

  const { data: componentsResult } = useQuery({
    queryKey: ['components', org?.id, app?.id],
    queryFn: () => getComponents({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled: !!org?.id && !!app?.id,
  })

  const rows = useMemo<IRuntimeChangeRow[]>(() => {
    const typeMap: Record<string, TComponentType> = {}
    for (const c of componentsResult?.data || []) {
      if (c.id && c.type) typeMap[c.id] = c.type
    }

    return (builds || [])
      .filter((b) => {
        const type = b.component_id ? typeMap[b.component_id] : undefined
        return !!type && IMAGE_TYPES.includes(type)
      })
      .map((b) => ({
        buildId: b.id || '',
        componentName: b.component_name,
        componentHref:
          org?.id && app?.id && b.component_id
            ? `/${org.id}/apps/${app.id}/components/${b.component_id}`
            : undefined,
        image: b.source_image,
        resolvedTag: b.resolved_tag,
        digest: b.source_digest,
        noOp: b.no_op,
        status: b.status_v2?.status || b.status,
        buildHref:
          org?.id && app?.id && b.component_id && b.id
            ? `/${org.id}/apps/${app.id}/components/${b.component_id}/builds/${b.id}`
            : undefined,
      }))
  }, [builds, componentsResult, org?.id, app?.id])

  return <RuntimeChanges rows={rows} />
}
