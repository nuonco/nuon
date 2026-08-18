import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { AppConfigDiff } from '@/components/branches/AppConfigDiff'
import { getBranchWorkflowRuns } from '@/lib'
import { useParams } from 'react-router'
import { Text } from '@/components/common/Text'
import { Status } from '@/components/common/Status'

interface IConfigStepContainer {
  metadata: Record<string, any>
  status?: string
  statusDescription?: string
}

export const ConfigStepContainer = ({ metadata, status, statusDescription }: IConfigStepContainer) => {
  const { org } = useOrg()
  const { app } = useApp()
  const params = useParams()
  const branchId = params.branchId as string
  const appConfigId = metadata.app_config_id as string | undefined

  const { data: branchRunsResult } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-runs', org?.id, app?.id, branchId],
    queryFn: () => getBranchWorkflowRuns({ orgId: org!.id, appId: app!.id, branchId, limit: 10 }),
    enabled: !!org?.id && !!app?.id && !!branchId,
  })

  const oldConfigId = (() => {
    const runs = branchRunsResult?.data
    if (!runs || !appConfigId) return undefined
    const sorted = [...runs].sort(
      (a, b) => new Date(b.created_at || 0).getTime() - new Date(a.created_at || 0).getTime()
    )
    const currentIdx = sorted.findIndex((r) => r.app_branch_runs?.at(0)?.app_config_id === appConfigId)
    if (currentIdx < 0) return undefined
    for (let i = currentIdx + 1; i < sorted.length; i++) {
      const prevRun = sorted[i].app_branch_runs?.at(0)
      if (prevRun?.app_config_id && prevRun.app_config_id !== appConfigId) {
        return prevRun.app_config_id
      }
    }
    return undefined
  })()

  return (
    <div className="flex flex-col gap-3">
      {status && (
        <div className="flex items-center gap-2">
          <Status status={status} variant="badge" />
          {statusDescription && (
            <Text variant="subtext" theme="neutral">{statusDescription}</Text>
          )}
        </div>
      )}
      {appConfigId && (
        <AppConfigDiff appConfigId={appConfigId} oldConfigId={oldConfigId} />
      )}
    </div>
  )
}
