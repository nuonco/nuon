import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import type { TInstallWorkflowStep } from '@/types'

function statusTheme(status?: string) {
  if (status === 'success') return 'success'
  if (status === 'error') return 'error'
  if (status === 'in-progress') return 'info'
  return 'neutral'
}

interface IWorkflowStepDetail {
  step: TInstallWorkflowStep
  onClose: () => void
}

export const WorkflowStepDetail = ({ step, onClose }: IWorkflowStepDetail) => {
  return (
    <Card>
      <div className="p-6">
        <div className="flex items-center justify-between mb-4">
          <Text variant="h3" weight="strong">
            Step details
          </Text>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <Icon variant="XIcon" size={20} />
          </Button>
        </div>

        <div className="space-y-4">
          <div className="flex items-start justify-between">
            <div>
              <Text variant="base" weight="strong" className="mb-2">
                {step.name || 'Unknown step'}
              </Text>
              <div className="flex items-center gap-3">
                <Badge theme={statusTheme(step.status?.status)}>
                  {step.status?.status || 'pending'}
                </Badge>
                {step.group_idx !== undefined && (
                  <Badge theme="neutral">
                    Group {step.group_idx}
                  </Badge>
                )}
              </div>
            </div>
            <div className="flex flex-col items-end gap-1">
              {step.started_at && (
                <Text variant="subtext" theme="neutral">
                  Started <Time time={step.started_at} format="relative" />
                </Text>
              )}
              {step.finished_at && (
                <Text variant="subtext" theme="neutral">
                  Finished <Time time={step.finished_at} format="relative" />
                </Text>
              )}
              {step.execution_time && (
                <Text variant="subtext" theme="neutral">
                  Duration: {(step.execution_time / 1000000000).toFixed(2)}s
                </Text>
              )}
            </div>
          </div>

          {step.status?.status_human_description && (
            <div className="p-4 bg-cool-grey-100 dark:bg-dark-grey-800 rounded-md">
              <Text variant="label" theme="neutral" className="mb-1">
                Status
              </Text>
              <Text variant="base">
                {step.status.status_human_description}
              </Text>
            </div>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Text variant="label" theme="neutral" className="mb-1">
                Step ID
              </Text>
              <ID>{step.id}</ID>
            </div>
            {step.idx !== undefined && (
              <div>
                <Text variant="label" theme="neutral" className="mb-1">
                  Index
                </Text>
                <Text variant="base">{step.idx}</Text>
              </div>
            )}
            {step.execution_type && (
              <div>
                <Text variant="label" theme="neutral" className="mb-1">
                  Execution type
                </Text>
                <Text variant="base">{step.execution_type}</Text>
              </div>
            )}
            {step.retryable !== undefined && (
              <div>
                <Text variant="label" theme="neutral" className="mb-1">
                  Retryable
                </Text>
                <Badge theme={step.retryable ? 'success' : 'neutral'}>
                  {step.retryable ? 'Yes' : 'No'}
                </Badge>
              </div>
            )}
          </div>

          {step.install_workflow_id && (
            <div>
              <Text variant="label" theme="neutral" className="mb-2">
                Quick links
              </Text>
              <div className="flex flex-wrap gap-2">
                <AdminDashboardLink
                  path={`/workflows/${step.install_workflow_id}`}
                  label="Admin panel"
                />
              </div>
            </div>
          )}
        </div>
      </div>
    </Card>
  )
}
