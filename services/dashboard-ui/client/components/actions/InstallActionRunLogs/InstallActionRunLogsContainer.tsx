import { useContext, useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { useUnifiedLogData, useLogViewer } from '@/hooks/use-logs'
import { InstallActionRunContext } from '@/providers/install-action-run-provider'
import type { TActionConfig, TInstallActionRunStep } from '@/types'
import { InstallActionRunLogs } from './InstallActionRunLogs'

export const InstallActionRunLogsContainer = ({
  actionConfig,
  layout = 'vertical',
  runSteps: runStepsProp,
}: {
  actionConfig: TActionConfig
  layout?: 'vertical' | 'horizontal'
  runSteps?: TInstallActionRunStep[]
}) => {
  const [searchParams] = useSearchParams()
  const { logs: allLogs, loadMore, hasMore, isLoading, isStreamOpen } = useUnifiedLogData()
  const { filteredLogs, activeLog, handleActiveLog, filters } = useLogViewer()
  const actionRunCtx = useContext(InstallActionRunContext)
  const runSteps = actionRunCtx?.installActionRun?.steps ?? runStepsProp

  const stepStatuses = useMemo(() => {
    if (!runSteps?.length || !actionConfig?.steps?.length) return undefined
    const idToName: Record<string, string> = {}
    for (const s of actionConfig.steps) {
      if (s?.id && s?.name) idToName[s.id] = s.name
    }
    const out: Record<string, string> = {}
    for (const rs of runSteps) {
      const name = rs?.step_id && idToName[rs.step_id]
      if (name && rs?.status) out[name] = rs.status
    }
    return Object.keys(out).length > 0 ? out : undefined
  }, [runSteps, actionConfig?.steps])

  return (
    <InstallActionRunLogs
      actionConfig={actionConfig}
      layout={layout}
      allLogs={allLogs}
      filteredLogs={filteredLogs}
      loadMore={loadMore}
      hasMore={hasMore}
      isLoading={isLoading}
      isStreamOpen={isStreamOpen}
      activeLog={activeLog}
      handleActiveLog={handleActiveLog}
      filters={filters}
      searchParamPanel={searchParams.get('panel')}
      stepStatuses={stepStatuses}
    />
  )
}
