import { Badge } from '@/components/common/Badge'
import { Checkbox } from '@/components/common/form/CheckboxInput'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import type { TInstall, TWorkflow } from '@/types'
import { cn } from '@/utils/classnames'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import { getPendingApprovalCount } from '@/utils/workflow-utils'
import { CancelWorkflowButton } from './CancelWorkflow'

function getWorkflowTitle(workflow: TWorkflow) {
  if (
    workflow?.type === 'action_workflow_run' &&
    workflow?.metadata?.adhoc_action
  ) {
    return `Adhoc action run (${workflow?.metadata?.install_action_workflow_name})`
  }
  return workflow.name || toSentenceCase(snakeToWords(workflow.type))
}

export const ActiveWorkflowCard = ({
  workflow,
  install,
  cancelMode,
  selected,
  onToggle,
}: {
  workflow: TWorkflow
  install?: TInstall
  cancelMode?: boolean
  selected?: boolean
  onToggle?: (workflowId: string) => void
}) => {
  const { org } = useOrg()
  const installId = workflow.owner_id
  const installName = workflow.metadata?.owner_name
  const pendingApprovals = getPendingApprovalCount(workflow)

  return (
    <div
      className={cn(
        'flex gap-4 rounded-lg border p-4',
        cancelMode && selected
          ? 'border-orange-400 dark:border-orange-500/40 bg-[#FFF5EB] dark:bg-[#2E1E10]'
          : 'border-cool-grey-200 dark:border-white/10 bg-cool-grey-50 dark:bg-white/[0.03]',
        cancelMode && 'cursor-pointer'
      )}
      onClick={cancelMode ? () => onToggle?.(workflow.id) : undefined}
    >
      {cancelMode && (
        <div className="flex items-start shrink-0 pt-1">
          <Checkbox
            checked={selected}
            onChange={() => onToggle?.(workflow.id)}
            onClick={(e) => e.stopPropagation()}
          />
        </div>
      )}
      <div className="flex flex-col gap-4 flex-1 min-w-0">
      <div className="flex flex-col gap-0.5 min-w-0">
        <div className="flex items-center gap-2">
          <Icon variant="Loading" size={14} className="shrink-0" />
          <Link
            href={`/${org.id}/installs/${installId}/workflows/${workflow.id}`}
            className="min-w-0"
          >
            <Text variant="base" weight="strong" className="truncate">
              {installName && !install && (
                <span className="text-cool-grey-500 dark:text-white/50 mr-1.5">
                  {installName} /
                </span>
              )}
              {getWorkflowTitle(workflow)}
            </Text>
          </Link>
          {pendingApprovals > 0 && (
            <Badge size="sm" theme="warn" className="shrink-0">
              Pending approval
            </Badge>
          )}
        </div>
        <ID>{workflow?.id}</ID>
      </div>

      <div className="flex items-center gap-6">
        <LabeledValue label="Initiated by" className="flex-1">

            {workflow?.created_by?.email?.split('@')[0] ?? '—'}

        </LabeledValue>
        <LabeledValue label="Elapsed time" className="flex-1">
          <Duration
            variant="subtext"
            beginTime={workflow.created_at}
            durationUnits={['hours', 'minutes', 'seconds']}
          />
        </LabeledValue>
        <LabeledValue label="Type" className="flex-1">
          <Text variant="subtext" className="truncate">
            {toSentenceCase(snakeToWords(workflow.type))}
          </Text>
        </LabeledValue>


      </div>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {workflow.plan_only && (
            <Badge variant="code" size="sm">
              drift scan
            </Badge>
          )}
          {workflow.plan_only &&
            install?.drifted_objects?.find(
              (d) => d?.install_workflow_id === workflow?.id
            ) && (
              <Badge size="sm" variant="code" theme="warn">
                drift detected
              </Badge>
            )}
          {(workflow?.type === 'drift_run_reprovision_sandbox' ||
            workflow?.type === 'drift_run') && (
            <Badge variant="code" size="sm">
              cron scheduled
            </Badge>
          )}
        </div>
        {!cancelMode && (
          <CancelWorkflowButton workflow={workflow} size="sm" />
        )}
      </div>
      </div>
    </div>
  )
}
