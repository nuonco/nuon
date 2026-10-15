import type { ReactNode } from 'react'
import { RunTiming } from '@/components/actions/RunTiming'
import { getTriggerLabel } from '@/components/actions/TriggeredBy'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { RunAdhocActionButton } from '@/components/installs/management/RunAdhocAction'
import type { IActionRunMetadata } from '../types'
import { useInstallLink } from '@/hooks/use-install-path'
import { adhocDisplayName } from '@/utils/action-utils'

interface IActionRunMetadataPresentation extends IActionRunMetadata {
  orgId: string
  rerunButton?: ReactNode
}

export const ActionRunMetadata = ({
  actionRun,
  createdBy,
  loading,
  step,
  orgId,
  rerunButton,
}: IActionRunMetadataPresentation) => {
  const installLink = useInstallLink()
  if (loading) {
    return (
      <div className="flex items-start gap-6">
        <LabeledStatus label="Status" loading />
        <LabeledValue label="Triggered by" loading />
      </div>
    )
  }

  const isAdhocActionRun = actionRun?.trigger_type === 'adhoc'
  const firstStep = actionRun?.steps?.at(0)
  const adhocConfig = firstStep?.adhoc_config
  const trigger = actionRun?.trigger
  const runbookName =
    trigger?.type === 'runbook' ? trigger.parent_name : undefined
  const triggeredBy =
    trigger?.actor?.email ||
    trigger?.actor?.name ||
    (isAdhocActionRun ? createdBy?.email : undefined) ||
    getTriggerLabel(trigger) ||
    actionRun?.triggered_by_type

  return (
    <div className="flex items-start justify-between gap-6">
      <div className="flex flex-wrap items-start gap-x-6 gap-y-3">
        <LabeledStatus
          label="Status"
          statusProps={{
            status: actionRun?.status_v2?.status,
          }}
          tooltipProps={{
            position: 'top',
            tipContent: actionRun?.status_v2?.status_human_description,
          }}
        />

        {runbookName ? (
          <LabeledValue label="Runbook">
            <Link
              href={installLink({
                orgId,
                installId: step?.owner_id,
                suffix: `/runbooks/${encodeURIComponent(runbookName)}`,
              })}
            >
              {runbookName}
            </Link>
          </LabeledValue>
        ) : null}

        <LabeledValue label="Triggered by">
          <Text variant="subtext">{triggeredBy ?? '—'}</Text>
        </LabeledValue>

        {actionRun?.run_env_vars?.COMPONENT_ID ? (
          <LabeledValue label="Component">
            <Link
              href={installLink({
                orgId,
                installId: step?.owner_id,
                suffix: `/components/${actionRun?.run_env_vars?.COMPONENT_ID}`,
              })}
            >
              {actionRun?.run_env_vars?.COMPONENT_NAME}
            </Link>
          </LabeledValue>
        ) : null}
        {actionRun?.created_at ? (
          <LabeledValue label="Started">
            <Time
              variant="subtext"
              time={actionRun.created_at}
              format="relative"
            />
          </LabeledValue>
        ) : null}

        {actionRun?.created_at ? (
          <LabeledValue label="Duration">
            <RunTiming variant="subtext" run={actionRun} />
          </LabeledValue>
        ) : null}

        {actionRun?.effective_timeout ? (
          <LabeledValue label="Timeout">
            <Duration
              variant="subtext"
              nanoseconds={actionRun.effective_timeout}
            />
          </LabeledValue>
        ) : null}

        {actionRun?.runner_job?.install_role_usage?.role_name ? (
          <LabeledValue label="Role">
            <Text variant="subtext">
              {actionRun.runner_job.install_role_usage.role_name}
            </Text>
          </LabeledValue>
        ) : null}
      </div>

      {isAdhocActionRun ? (
        <div className="self-end">
          <RunAdhocActionButton
            initialValues={{
              name: adhocDisplayName(adhocConfig?.name) ?? '',
              command: adhocConfig?.command,
              inline_contents: adhocConfig?.inline_contents,
              env_vars: adhocConfig?.env_vars,
              timeout: (adhocConfig as any)?.timeout,
              role: actionRun?.role,
            }}
          >
            Edit and rerun
            <Icon variant="TerminalWindowIcon" />
          </RunAdhocActionButton>
        </div>
      ) : rerunButton ? (
        <div className="self-end">{rerunButton}</div>
      ) : null}
    </div>
  )
}
