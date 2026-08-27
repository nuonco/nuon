import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import type { TInstallWorkflowStep } from '@/types'
import { DetailStatusIcon } from './shared/icons'
import { formatDuration } from './shared/format'
import { CommitStep } from './steps/CommitStep'
import { ConfigStep } from './steps/ConfigStep'
import { BuildStep } from './steps/BuildStep'
import { ComparisonStep } from './steps/ComparisonStep'
import { PlanGroupStep } from './steps/PlanGroupStep'
import { DeployGroupStep } from './steps/DeployGroupStep'
import { PostDeployRunbooksStep } from './steps/PostDeployRunbooksStep'

interface IWorkflowStepDetail {
  step: TInstallWorkflowStep
  appBranchId?: string
  appBranchRunId?: string
  onClose: () => void
}

export const WorkflowStepDetail = ({
  step,
  appBranchId,
  appBranchRunId,
  onClose: _onClose,
}: IWorkflowStepDetail) => {
  const metadata = step.status?.metadata || {}

  const isCommitStep = step.name?.toLowerCase().includes('commit')
  const isBuildStep = step.name?.toLowerCase().includes('build')
  const isComparisonStep =
    step.name?.toLowerCase().includes('compute differences') ||
    step.name?.toLowerCase().includes('compute run comparison')
  const isConfigStep =
    step.name?.toLowerCase().includes('config') &&
    !step.name?.toLowerCase().includes('diff') &&
    !step.name?.toLowerCase().includes('differences')
  const isPlanGroupStep = step.name
    ?.toLowerCase()
    .includes('plan install group')
  const isDeployGroupStep = step.name
    ?.toLowerCase()
    .includes('deploy install group')
  const isPostDeployRunbooksStep = step.name
    ?.toLowerCase()
    .includes('post-deploy runbooks')

  const isInProgress = step.status?.status === 'in-progress'
  const duration = formatDuration(step.execution_time)

  const cardBorderClass = isInProgress
    ? 'border-blue-400/40 dark:border-blue-500/40'
    : ''
  const cardShadow = isInProgress
    ? '0 0 0 3px rgba(63,116,224,0.08), 0 0 16px rgba(63,116,224,0.10)'
    : undefined

  const stepIndexStr = String(step.group_idx ?? '').padStart(2, '0') || '—'

  return (
    <div
      className={`rounded-xl border bg-white dark:bg-dark-grey-900 overflow-hidden transition-all ${cardBorderClass}`}
      style={cardShadow ? { boxShadow: cardShadow } : undefined}
    >
      <div className="flex items-center gap-3 px-5 py-4 border-b">
        <DetailStatusIcon status={step.status?.status} />
        <Text
          variant="subtext"
          family="mono"
          theme="neutral"
          className="shrink-0"
        >
          {stepIndexStr}
        </Text>
        <Text
          as="h2"
          variant="h3"
          weight="strong"
          className="leading-tight flex-none"
        >
          {step.name || 'Step details'}
        </Text>

        <Status
          status={step.status?.status || 'pending'}
          variant="badge"
          className="shrink-0"
        />
        <div className="flex-1" />
        <div className="flex items-center gap-3 shrink-0">
          {step.started_at && (
            <div className="flex items-center gap-1.5">
              <Text variant="subtext" theme="neutral">
                Started
              </Text>
              <Time
                time={step.started_at}
                format="relative"
                variant="subtext"
                theme="neutral"
              />
            </div>
          )}
          {duration && (
            <div className="flex items-center gap-1.5 text-cool-grey-400 dark:text-cool-grey-500">
              <Icon variant="ClockIcon" size={13} />
              <Text variant="subtext" family="mono" theme="neutral">
                {duration}
              </Text>
            </div>
          )}
        </div>
      </div>

      <div className="p-5 flex flex-col gap-4">
        {isCommitStep && <CommitStep metadata={metadata} />}
        {isConfigStep && (
          <ConfigStep
            metadata={metadata}
            status={step.status?.status}
            statusDescription={step.status?.status_human_description}
          />
        )}
        {isComparisonStep && (
          <ComparisonStep
            metadata={metadata}
            status={step.status?.status}
            appBranchId={appBranchId}
            appBranchRunId={appBranchRunId}
          />
        )}
        {isBuildStep && (
          <BuildStep
            metadata={metadata}
            status={step.status?.status}
            appBranchId={appBranchId}
            appBranchRunId={appBranchRunId}
          />
        )}
        {isPlanGroupStep && <PlanGroupStep step={step} metadata={metadata} />}
        {isDeployGroupStep && (
          <DeployGroupStep step={step} metadata={metadata} />
        )}
        {isPostDeployRunbooksStep && (
          <PostDeployRunbooksStep step={step} metadata={metadata} />
        )}

        {!isCommitStep &&
          !isBuildStep &&
          !isComparisonStep &&
          !isConfigStep &&
          !isPlanGroupStep &&
          !isDeployGroupStep &&
          !isPostDeployRunbooksStep &&
          step.status?.status_human_description && (
            <div className="p-3 bg-cool-grey-100 dark:bg-dark-grey-800 rounded-md">
              <Text variant="base">{step.status.status_human_description}</Text>
            </div>
          )}

        {step.install_workflow_id && (
          <div className="flex items-center gap-4 pt-3 border-t">
            <AdminDashboardLink
              path={`/workflows/${step.install_workflow_id}`}
              label="admin panel"
            />
          </div>
        )}
      </div>
    </div>
  )
}
