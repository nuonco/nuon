import type { ReactNode } from 'react'
import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { useAuth } from '@/hooks/use-auth'
import { cn } from '@/utils/classnames'
import type { TInstallWorkflowStep } from '@/types'
import { DetailStatusIcon } from './shared/icons'
import { formatDuration } from './shared/format'
import { STEP_GUTTER as GUTTER } from './shared/StepLayout'

const AdminFooter = ({ workflowId }: { workflowId: string }) => {
  const { isAdmin } = useAuth()
  if (!isAdmin) return null

  return (
    <div className={cn('flex items-center gap-4 py-3 border-t', GUTTER)}>
      <AdminDashboardLink path={`/workflows/${workflowId}`} label="admin panel" />
    </div>
  )
}

export interface IStepCard {
  step: TInstallWorkflowStep
  children?: ReactNode
}

export const StepCard = ({ step, children }: IStepCard) => {
  const isInProgress = step.status?.status === 'in-progress'
  const duration = formatDuration(step.execution_time)
  const description = step.status?.status_human_description
  const stepIndexStr = String(step.group_idx ?? '').padStart(2, '0') || '—'

  return (
    <div
      className={cn(
        'rounded-xl border bg-white dark:bg-dark-grey-900 overflow-hidden transition-all',
        isInProgress && 'border-blue-400/40 dark:border-blue-500/40'
      )}
      style={
        isInProgress
          ? {
              boxShadow:
                '0 0 0 3px rgba(63,116,224,0.08), 0 0 16px rgba(63,116,224,0.10)',
            }
          : undefined
      }
    >
      <div className={cn('flex flex-col gap-1 py-4 border-b', GUTTER)}>
        <div className="flex items-center gap-3">
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

        {description && (
          <Text variant="subtext" theme="neutral" className="pl-[38px]">
            {description}
          </Text>
        )}
      </div>

      {children && <div className="flex flex-col divide-y">{children}</div>}

      {step.install_workflow_id && (
        <AdminFooter workflowId={step.install_workflow_id} />
      )}
    </div>
  )
}
