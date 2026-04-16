import { Badge } from '@/components/common/Badge'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import type { TInstall, TWorkflow } from '@/types'
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
}: {
  workflow: TWorkflow
  install?: TInstall
}) => {
  const { org } = useOrg()
  const installId = workflow.owner_id
  const installName = workflow.metadata?.owner_name
  const pendingApprovals = getPendingApprovalCount(workflow)

  return (
    <div className="flex flex-col gap-4 rounded-lg border border-cool-grey-200 dark:border-white/10 bg-cool-grey-50 dark:bg-white/[0.03] p-4">
      <div className="flex items-start justify-between gap-4">
        <div className="flex flex-col gap-0.5 min-w-0">
          <Link
            href={`/${org.id}/installs/${installId}/workflows/${workflow.id}`}
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
          <ID>{workflow?.id}</ID>
        </div>
        <Icon variant="Loading" size={16} className="shrink-0 mt-1" />
      </div>

      <div className="flex items-center gap-6 flex-wrap">
        <LabeledValue label="Initiated by" className="w-24">

            {workflow?.created_by?.email?.split('@')[0] ?? '—'}

        </LabeledValue>
        <LabeledValue label="Elapsed time" className="w-24">
          <Duration
            variant="subtext"
            beginTime={workflow.created_at}
            durationUnits={['hours', 'minutes', 'seconds']}
          />
        </LabeledValue>
        <LabeledValue label="Type" className="w-24">
          <Text variant="subtext" className="truncate">
            {toSentenceCase(snakeToWords(workflow.type))}
          </Text>
        </LabeledValue>

        {(workflow.plan_only ||
          workflow?.type === 'drift_run_reprovision_sandbox' ||
          workflow?.type === 'drift_run' ||
          pendingApprovals > 0) && (
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
            {pendingApprovals > 0 && (
              <Badge size="sm" theme="warn">
                Pending approval
              </Badge>
            )}
          </div>
        )}

        <div className="ml-auto shrink-0">
          <CancelWorkflowButton workflow={workflow} size="sm" />
        </div>
      </div>
    </div>
  )
}
