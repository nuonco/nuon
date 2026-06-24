import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import type { TInstallWorkflowStep } from '@/types'
import { DetailStatusIcon } from './shared/icons'
import { statusTheme, formatDuration } from './shared/format'
import { CommitStep } from './steps/CommitStep'
import { ConfigStep } from './steps/ConfigStep'
import { BuildStep } from './steps/BuildStep'
import { PlanGroupStep } from './steps/PlanGroupStep'
import { DeployGroupStep } from './steps/DeployGroupStep'

interface IWorkflowStepDetail {
  step: TInstallWorkflowStep
  onClose: () => void
}

export const WorkflowStepDetail = ({ step, onClose: _onClose }: IWorkflowStepDetail) => {
  const metadata = step.status?.metadata || {}

  const isCommitStep = step.name?.toLowerCase().includes('commit')
  const isBuildStep = step.name?.toLowerCase().includes('build')
  const isConfigStep = step.name?.toLowerCase().includes('config') && !step.name?.toLowerCase().includes('diff')
  const isPlanGroupStep = step.name?.toLowerCase().includes('plan install group')
  const isDeployGroupStep = step.name?.toLowerCase().includes('deploy install group')

  const isInProgress = step.status?.status === 'in-progress'
  const duration = formatDuration(step.execution_time)

  const cardBorderClass = isInProgress
    ? 'border-blue-400/40 dark:border-blue-500/40'
    : 'border-cool-grey-200 dark:border-dark-grey-700'
  const cardShadow = isInProgress
    ? '0 0 0 3px rgba(63,116,224,0.08), 0 0 16px rgba(63,116,224,0.10)'
    : undefined

  const stepIndexStr = String(step.group_idx ?? '').padStart(2, '0') || '—'

  return (
    <div
      className={`rounded-xl border bg-white dark:bg-dark-grey-900 overflow-hidden transition-all ${cardBorderClass}`}
      style={cardShadow ? { boxShadow: cardShadow } : undefined}
    >
      {/* ── Header row ── */}
      <div className="flex items-center gap-3 px-5 py-4 border-b border-cool-grey-100 dark:border-dark-grey-800">
        <DetailStatusIcon status={step.status?.status} />
        <span className="font-mono text-[12px] text-cool-grey-400 dark:text-cool-grey-500 shrink-0">
          {stepIndexStr}
        </span>
        <h2 className="text-[18px] font-semibold text-cool-grey-900 dark:text-white leading-tight flex-none">
          {step.name || 'Step details'}
        </h2>
        {step.group_idx !== undefined && (
          <span className="text-[10.5px] uppercase tracking-[0.07em] font-semibold px-2 py-0.5 rounded-full border border-cool-grey-300 dark:border-dark-grey-600 text-cool-grey-500 dark:text-cool-grey-400 bg-cool-grey-50 dark:bg-dark-grey-800 shrink-0">
            Group {step.group_idx}
          </span>
        )}
        <Badge theme={statusTheme(step.status?.status)} size="sm" className="shrink-0">
          {isInProgress && (
            <svg className="animate-spin w-3 h-3 shrink-0" viewBox="0 0 12 12" fill="none">
              <circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeOpacity="0.3" strokeWidth="1.5" />
              <path d="M6 1.5 A4.5 4.5 0 0 1 10.5 6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
            </svg>
          )}
          {step.status?.status || 'pending'}
        </Badge>
        <div className="flex-1" />
        {duration && (
          <div className="flex items-center gap-1.5 text-cool-grey-400 dark:text-cool-grey-500 shrink-0">
            <Icon variant="ClockIcon" size={13} />
            <span className="font-mono text-[12px]">{duration}</span>
          </div>
        )}
      </div>

      {/* ── Sub-bar: metadata row ── */}
      <div className="flex items-start gap-6 px-5 py-3 bg-cool-grey-50 dark:bg-dark-grey-800 border-b border-cool-grey-100 dark:border-dark-grey-800 flex-wrap">
        <div className="flex flex-col gap-0.5">
          <span className="text-[10.5px] uppercase tracking-[0.06em] font-semibold text-cool-grey-400 dark:text-cool-grey-500">Step ID</span>
          <ID className="text-[12px]">{step.id}</ID>
        </div>
        {step.started_at && (
          <div className="flex flex-col gap-0.5">
            <span className="text-[10.5px] uppercase tracking-[0.06em] font-semibold text-cool-grey-400 dark:text-cool-grey-500">Started</span>
            <Time time={step.started_at} format="relative" variant="subtext" />
          </div>
        )}
        <div className="flex flex-col gap-0.5">
          <span className="text-[10.5px] uppercase tracking-[0.06em] font-semibold text-cool-grey-400 dark:text-cool-grey-500">Execution</span>
          <span className="text-[12px] text-cool-grey-700 dark:text-cool-grey-200">{step.execution_type || 'system'}</span>
        </div>
        {step.retryable !== undefined && (
          <div className="flex flex-col gap-0.5">
            <span className="text-[10.5px] uppercase tracking-[0.06em] font-semibold text-cool-grey-400 dark:text-cool-grey-500">Retryable</span>
            <Badge theme={step.retryable ? 'success' : 'neutral'} size="sm">
              {step.retryable ? 'Yes' : 'No'}
            </Badge>
          </div>
        )}
      </div>

      {/* ── Content area ── */}
      <div className="p-5 space-y-4">
        {isCommitStep && <CommitStep metadata={metadata} />}
        {isConfigStep && <ConfigStep metadata={metadata} status={step.status?.status} />}
        {isBuildStep && <BuildStep metadata={metadata} status={step.status?.status} />}
        {isPlanGroupStep && <PlanGroupStep step={step} metadata={metadata} />}
        {isDeployGroupStep && <DeployGroupStep step={step} metadata={metadata} />}

        {/* Generic fallback */}
        {!isCommitStep && !isBuildStep && !isConfigStep && !isPlanGroupStep && !isDeployGroupStep && step.status?.status_human_description && (
          <div className="p-3 bg-cool-grey-100 dark:bg-dark-grey-800 rounded-md">
            <Text variant="base">{step.status.status_human_description}</Text>
          </div>
        )}

        {/* Footer */}
        {step.install_workflow_id && (
          <div className="flex items-center gap-4 pt-3 border-t border-cool-grey-200 dark:border-dark-grey-700">
            <AdminDashboardLink path={`/workflows/${step.install_workflow_id}`} label="admin panel" />
          </div>
        )}
      </div>
    </div>
  )
}
