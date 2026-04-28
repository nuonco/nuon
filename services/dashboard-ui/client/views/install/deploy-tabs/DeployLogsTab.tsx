import { useOutletContext } from 'react-router'
import { Plan } from '@/components/approvals/Plan'
import { SSELogs, LogsSkeleton } from '@/components/log-stream/SSELogs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider'
import { useDeploy } from '@/hooks/use-deploy'
import { useRespondedApprovals } from '@/hooks/use-responded-approvals'
import type { TDeployOutletContext } from './types'

export const DeployLogsTab = () => {
  const { deploy } = useDeploy()
  const { step } = useOutletContext<TDeployOutletContext>()
  const { hasResponded } = useRespondedApprovals()

  const responded = step ? hasResponded(step.id) : false
  const stepStatus = step?.status?.status
  const isTerminal = stepStatus === 'error' || stepStatus === 'cancelled' || stepStatus === 'discarded'
  const isAutoApprove =
    step?.approval?.type === 'approve-all' ||
    step?.approval?.response?.type === 'auto-approve'
  const completedApproval =
    step?.approval && (!!step?.approval?.response || responded) && !isTerminal && stepStatus !== 'auto-skipped'
  const showPlanBelow = completedApproval || isAutoApprove

  const logStream = deploy?.log_stream

  return (
    <div className="flex flex-col gap-6">
      {logStream ? (
        <LogStreamProvider logStreamId={logStream.id} shouldPoll={logStream.open}>
          <UnifiedLogsProvider>
            <LogViewerProvider>
              <SSELogs />
            </LogViewerProvider>
          </UnifiedLogsProvider>
        </LogStreamProvider>
      ) : (
        <LogsSkeleton />
      )}

      {showPlanBelow && step ? (
        <Plan step={step} />
      ) : null}
    </div>
  )
}
