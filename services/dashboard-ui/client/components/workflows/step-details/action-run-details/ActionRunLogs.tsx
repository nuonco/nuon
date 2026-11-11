import { EmptyState } from '@/components/common/EmptyState'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { ActionRunLogsDeleted } from '@/components/actions/DeleteActionRunLogs'
import { InstallActionRunLogs } from '@/components/actions/InstallActionRunLogs'
import { InstallActionRunOutputsComponent } from '@/components/actions/InstallActionRunOutputs'
import { SSELogs } from '@/components/log-stream/SSELogs'
import { LogsSkeleton } from '@/components/log-stream/SSELogs'
import { TraceView } from '@/components/spans/TraceView'
import { usePrefetchLogStreamSpans } from '@/hooks/use-prefetch-log-stream-spans'
import { useStepTab } from '@/components/workflows/step-details/use-step-tab'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import type { TInstallActionRun } from '@/types'
import type { IActionRunLogs } from './types'

const ActionRunSummary = ({ actionRun }: { actionRun: TInstallActionRun }) => (
  <div className="pt-4 flex flex-col gap-6">
    <div className="flex flex-col gap-2">
      <Text weight="strong">Outputs</Text>
      <InstallActionRunOutputsComponent installActionRun={actionRun} />
    </div>
    {Object.keys(actionRun?.run_env_vars ?? {}).length > 0 && (
      <div className="flex flex-col gap-2">
        <Text weight="strong">Environment variables</Text>
        <KeyValueList
          values={Object.entries(actionRun.run_env_vars!).map(
            ([key, value]) => ({ key, value })
          )}
        />
      </div>
    )}
  </div>
)

export const ActionRunLogs = ({
  actionRun,
  isAdhoc,
  loading,
}: IActionRunLogs) => {
  const logsDeleted = !!actionRun?.log_stream?.logs_deleted_at
  usePrefetchLogStreamSpans(logsDeleted ? undefined : actionRun?.log_stream?.id)
  const [tab, setTab] = useStepTab()

  if (loading) {
    return (
      <div className="flex flex-col gap-4">
        <LogsSkeleton />
      </div>
    )
  }

  if (logsDeleted && isAdhoc) {
    return (
      <div className="flex flex-col gap-2">
        <Text weight="strong">Action logs</Text>
        <ActionRunLogsDeleted />
      </div>
    )
  }

  if (logsDeleted) {
    return (
      <div className="flex flex-col gap-2">
        <Tabs
          naturalHeight
          activeTab={tab}
          onTabChange={setTab}
          tabs={{
            logs: (
              <div className="pt-4">
                <ActionRunLogsDeleted />
              </div>
            ),
            trace: (
              <div className="pt-4">
                <ActionRunLogsDeleted />
              </div>
            ),
            summary: <ActionRunSummary actionRun={actionRun!} />,
          }}
        />
      </div>
    )
  }

  if (!actionRun?.log_stream) {
    return (
      <div className="flex flex-col gap-2">
        {isAdhoc && <Text weight="strong">Action logs</Text>}
        <EmptyState
          variant="history"
          emptyTitle="Waiting for logs"
          emptyMessage="Logs will appear here as soon as the runner starts streaming them."
        />
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-2">
      <LogStreamProvider logStreamId={actionRun.log_stream.id}>
        <LogViewerProvider>
          {isAdhoc ? (
            <>
              <Text weight="strong">Action logs</Text>
              <Tabs
                naturalHeight
                activeTab={tab}
                onTabChange={setTab}
                tabs={{
                  logs: <SSELogs />,
                  trace: (
                    <TraceView
                      logStreamId={actionRun.log_stream.id}
                      shouldPoll={actionRun.log_stream.open}
                    />
                  ),
                }}
              />
            </>
          ) : (
            <Tabs
              naturalHeight
              activeTab={tab}
              onTabChange={setTab}
              tabs={{
                logs: (
                  <div className="pt-4">
                    <InstallActionRunLogs
                      actionConfig={actionRun?.config}
                      layout="horizontal"
                      runSteps={actionRun?.steps}
                    />
                  </div>
                ),
                trace: (
                  <div className="pt-4">
                    <TraceView
                      logStreamId={actionRun.log_stream.id}
                      shouldPoll={actionRun.log_stream.open}
                    />
                  </div>
                ),
                summary: <ActionRunSummary actionRun={actionRun} />,
              }}
            />
          )}
        </LogViewerProvider>
      </LogStreamProvider>
    </div>
  )
}
