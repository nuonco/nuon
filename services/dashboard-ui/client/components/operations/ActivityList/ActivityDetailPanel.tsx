import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { actorName } from '@/components/actions/TriggeredBy'
import { Badge } from '@/components/common/Badge'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PolicyReportDetails } from '@/components/policies/PolicyReportPanel'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import {
  InstallWorkflowPanelContent,
  InstallWorkflowPanelLoading,
} from '@/components/workflows/InstallWorkflowPanel'
import { useInstallLink } from '@/hooks/use-install-path'
import {
  getAppPoliciesConfigs,
  getInstallActionRun,
  getPolicyReport,
} from '@/lib'
import { WorkflowProvider } from '@/providers/workflow-provider'
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
  const actionHref =
    activity.action?.name && !activity.action.adhoc
      ? installLink({
          orgId,
          installId,
          suffix: activity.action.action_workflow_id
            ? `/actions/${activity.action.action_workflow_id}`
            : `/operations/actions?q=${encodeURIComponent(activity.action.name)}`,
        })
      : undefined
  const runbookHref = activity.runbook?.name
    ? installLink({
        orgId,
        installId,
        suffix: activity.runbook.runbook_id
          ? `/runbooks/${activity.runbook.runbook_id}`
          : `/runbooks/${encodeURIComponent(activity.runbook.name)}`,
      })
    : undefined
  const actionRunId = activity.action?.run_id
  const { data: actionRun, isLoading: actionRunLoading } = useQuery({
    queryKey: ['install-action-run', orgId, installId, actionRunId],
    queryFn: () =>
      getInstallActionRun({ orgId, installId, runId: actionRunId! }),
    enabled: !activity.workflow?.id && !!orgId && !!installId && !!actionRunId,
  })
  const workflowId =
    activity.workflow?.id ||
    actionRun?.workflow_id ||
    actionRun?.install_workflow_id
  const showsWorkflow = !!workflowId || !!actionRunId
  const loading = (
    <InstallWorkflowPanelLoading
      status={activity.status}
      createdAt={activity.created_at}
    />
  )

  return (
    <Panel
      heading={activity.title}
      size={showsWorkflow ? '3/4' : 'half'}
      aria-label="Activity details"
      {...props}
    >
      {workflowId ? (
        <WorkflowProvider
          workflowId={workflowId}
          shouldPoll
          loadingElement={loading}
        >
          <InstallWorkflowPanelContent />
        </WorkflowProvider>
      ) : actionRunLoading ? (
        loading
      ) : activity.policy?.report_id ? (
        <PolicyActivityDetails
          activity={activity}
          orgId={orgId}
          actionHref={actionHref}
          runbookHref={runbookHref}
        />
      ) : (
        <ActivitySummary
          activity={activity}
          actionHref={actionHref}
          runbookHref={runbookHref}
        />
      )}
    </Panel>
  )
}

const PolicyActivityDetails = ({
  activity,
  orgId,
  actionHref,
  runbookHref,
}: {
  activity: TInstallActivity
  orgId: string
  actionHref?: string
  runbookHref?: string
}) => {
  const reportId = activity.policy?.report_id
  const { data: report, isLoading } = useQuery({
    queryKey: ['policy-report', orgId, reportId],
    queryFn: () => getPolicyReport({ orgId, reportId: reportId! }),
    enabled: !!orgId && !!reportId,
  })
  const appId = report?.app_id
  const { data: policiesConfigs } = useQuery({
    queryKey: ['app-policies-configs', orgId, appId],
    queryFn: () => getAppPoliciesConfigs({ orgId, appId: appId! }),
    enabled: !!orgId && !!appId,
  })
  const policyNameMap = useMemo(() => {
    const map = new Map<string, string>()
    for (const config of policiesConfigs ?? []) {
      for (const policy of config.policies ?? []) {
        if (policy.id && policy.name) map.set(policy.id, policy.name)
      }
    }
    return map
  }, [policiesConfigs])

  if (isLoading) {
    return <Text variant="subtext" loading loadingWidth={24} />
  }
  if (!report) {
    return (
      <ActivitySummary
        activity={activity}
        actionHref={actionHref}
        runbookHref={runbookHref}
      />
    )
  }

  return (
    <PolicyReportDetails
      report={report}
      orgId={orgId}
      policyNameMap={policyNameMap}
    />
  )
}

const ActivitySummary = ({
  activity,
  actionHref,
  runbookHref,
}: {
  activity: TInstallActivity
  actionHref?: string
  runbookHref?: string
}) => (
  <>
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
        <Time time={activity.created_at} format="relative" variant="subtext" />
      </LabeledValue>
      <LabeledValue label="Activity ID">
        <ID>{activity.id}</ID>
      </LabeledValue>
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
      {activity.actor && (
        <LabeledValue label="Run by">
          <Text variant="subtext">{actorName(activity.actor)}</Text>
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
              {activity.policy.deny_count} denied, {activity.policy.warn_count}{' '}
              warnings, {activity.policy.pass_count} passed
            </Text>
          </LabeledValue>
        </>
      )}
    </div>
  </>
)
