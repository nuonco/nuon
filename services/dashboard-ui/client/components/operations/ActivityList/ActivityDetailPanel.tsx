import { Badge } from '@/components/common/Badge'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useInstallLink } from '@/hooks/use-install-path'
import type { TInstallActivity } from '@/types'
import { humanize } from '@/utils/string-utils'
import { ACTIVITY_TYPE_LABELS } from './ActivityListPresenter'

const POLICY_OWNER_LABELS: Record<string, string> = {
  install_deploys: 'Component deploy',
  install_sandbox_runs: 'Sandbox',
  component_builds: 'Component build',
}

export const ActivityDetailPanel = ({
  activity,
  orgId,
  installId,
  ...props
}: {
  activity: TInstallActivity
  orgId: string
  installId: string
} & Partial<IPanel>) => {
  const installLink = useInstallLink()
  const workflowHref = activity.workflow
    ? installLink({
        orgId,
        installId,
        suffix: `/deployments/${activity.workflow.id}`,
      })
    : undefined
  const actionHref =
    activity.action?.action_workflow_id && activity.action.run_id
      ? installLink({
          orgId,
          installId,
          suffix: `/actions/${activity.action.action_workflow_id}/runs/${activity.action.run_id}`,
        })
      : undefined
  const runbookHref = activity.runbook?.runbook_id
    ? installLink({
        orgId,
        installId,
        suffix: `/runbooks/${activity.runbook.runbook_id}`,
      })
    : undefined

  return (
    <Panel heading={activity.title} size="half" {...props}>
      <div className="flex items-center gap-2 flex-wrap">
        <Status status={activity.status} variant="badge" />
        <Badge size="sm" theme="neutral">
          {ACTIVITY_TYPE_LABELS[activity.type]}
        </Badge>
      </div>

      {activity.summary && (
        <Text variant="subtext" theme="neutral">
          {activity.summary}
        </Text>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <LabeledValue label="Ran">
          <Time
            time={activity.created_at}
            format="relative"
            variant="subtext"
          />
        </LabeledValue>
        <LabeledValue label="Activity ID">
          <ID>{activity.id}</ID>
        </LabeledValue>
        {activity.workflow && workflowHref && (
          <LabeledValue label="Workflow">
            <Link href={workflowHref}>{activity.workflow.name}</Link>
          </LabeledValue>
        )}
        {activity.action && (
          <LabeledValue label="Action">
            {actionHref ? (
              <Link href={actionHref}>
                {activity.action.name || activity.title}
              </Link>
            ) : (
              <Text variant="subtext">{activity.title}</Text>
            )}
          </LabeledValue>
        )}
        {activity.action?.trigger_type && (
          <LabeledValue label="Trigger">
            <Text variant="subtext">
              {humanize(activity.action.trigger_type)}
            </Text>
          </LabeledValue>
        )}
        {activity.runbook && (
          <LabeledValue label="Runbook">
            {runbookHref ? (
              <Link href={runbookHref}>
                {activity.runbook.name || activity.title}
              </Link>
            ) : (
              <Text variant="subtext">{activity.title}</Text>
            )}
          </LabeledValue>
        )}
        {activity.policy && (
          <>
            <LabeledValue label="Checked">
              <Text variant="subtext">
                {POLICY_OWNER_LABELS[activity.policy.owner_type] ??
                  humanize(activity.policy.owner_type)}
              </Text>
            </LabeledValue>
            {activity.policy.component_name && (
              <LabeledValue label="Component">
                <Text variant="subtext">{activity.policy.component_name}</Text>
              </LabeledValue>
            )}
            <LabeledValue label="Results">
              <Text variant="subtext">
                {activity.policy.deny_count} denied,{' '}
                {activity.policy.warn_count} warnings,{' '}
                {activity.policy.pass_count} passed
              </Text>
            </LabeledValue>
          </>
        )}
      </div>
    </Panel>
  )
}
