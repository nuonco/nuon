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

export const ActiveWorkflowCard = ({
  workflow,
  install,
  selectable,
  selected,
  onToggle,
  compact,
}: {
  workflow: TWorkflow
  install?: TInstall
  selectable?: boolean
  selected?: boolean
  onToggle?: (workflowId: string) => void
  compact?: boolean
}) => {
  const { org } = useOrg()
  const installId = workflow.owner_id
  const installName = workflow.metadata?.owner_name
  const pendingApprovals = getPendingApprovalCount(workflow)
  const pendingApprovalStep = workflow?.steps?.find(
    (s) =>
      s?.execution_type === 'approval' &&
      s?.status?.status === 'approval-awaiting' &&
      !!s?.approval &&
      !s?.approval?.response
  )
  const showDriftDetected =
    workflow.plan_only &&
    !!install?.drifted_objects?.find(
      (d) => d?.install_workflow_id === workflow?.id
    )
  const showCronScheduled =
    workflow?.type === 'drift_run_reprovision_sandbox' ||
    workflow?.type === 'drift_run'

  const titleAndBadges = (
    <>
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
          {workflow.name}
        </Text>
      </Link>
      {pendingApprovals > 0 &&
        (pendingApprovalStep ? (
          <Link
            href={`/${org.id}/installs/${installId}/workflows/${workflow.id}?panel=${pendingApprovalStep.id}`}
            className="shrink-0 hover:opacity-80 transition-opacity !text-inherit"
          >
            <Badge size="sm" theme="warn">
              Pending approval
            </Badge>
          </Link>
        ) : (
          <Badge size="sm" theme="warn" className="shrink-0">
            Pending approval
          </Badge>
        ))}
      {workflow.plan_only && (
        <Badge variant="code" size="sm" className="shrink-0">
          drift scan
        </Badge>
      )}
      {showDriftDetected && (
        <Badge size="sm" variant="code" theme="warn" className="shrink-0">
          drift detected
        </Badge>
      )}
      {showCronScheduled && (
        <Badge variant="code" size="sm" className="shrink-0">
          cron scheduled
        </Badge>
      )}
    </>
  )

  return (
    <div
      className={cn(
        'flex gap-4 rounded-lg border p-4 transition-colors',
        selectable && selected
          ? 'border-blue-500 dark:border-blue-400 bg-blue-50 dark:bg-blue-900/15'
          : 'border-cool-grey-200 dark:border-white/10 bg-cool-grey-50 dark:bg-white/[0.03]'
      )}
    >
      {selectable && (
        <div className="flex items-start shrink-0 pt-1">
          <Checkbox
            checked={selected}
            onChange={() => onToggle?.(workflow.id)}
          />
        </div>
      )}
      <div className="flex flex-col gap-4 flex-1 min-w-0">
        <div className="flex flex-col gap-0.5 min-w-0">
          {compact ? (
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center flex-wrap gap-2 min-w-0">
                {titleAndBadges}
              </div>
              <div className="flex items-center gap-3 shrink-0">
                <div className="flex items-center gap-1.5">
                  <Icon variant="Loading" size={14} className="shrink-0" />
                  <Duration
                    variant="subtext"
                    theme="neutral"
                    beginTime={workflow.created_at}
                    durationUnits={['hours', 'minutes', 'seconds']}
                  />
                </div>
                <CancelWorkflowButton workflow={workflow} size="sm">
                  Cancel
                </CancelWorkflowButton>
              </div>
            </div>
          ) : (
            <div className="flex items-center flex-wrap gap-2">
              {titleAndBadges}
            </div>
          )}
          <ID>{workflow?.id}</ID>
        </div>

        {!compact && (
          <div className="flex items-end gap-6">
            <LabeledValue label="Initiated by" className="flex-1">
              {workflow?.created_by?.email?.split('@')[0] ?? '—'}
            </LabeledValue>
            <LabeledValue label="Elapsed time" className="flex-1">
              <div className="flex items-center gap-1.5">
                <Icon variant="Loading" size={14} className="shrink-0" />
                <Duration
                  variant="subtext"
                  beginTime={workflow.created_at}
                  durationUnits={['hours', 'minutes', 'seconds']}
                />
              </div>
            </LabeledValue>
            <LabeledValue label="Type" className="flex-1">
              <Text variant="subtext" className="truncate">
                {toSentenceCase(snakeToWords(workflow.type))}
              </Text>
            </LabeledValue>
            <CancelWorkflowButton workflow={workflow} size="sm">
              Cancel
            </CancelWorkflowButton>
          </div>
        )}
      </div>
    </div>
  )
}
