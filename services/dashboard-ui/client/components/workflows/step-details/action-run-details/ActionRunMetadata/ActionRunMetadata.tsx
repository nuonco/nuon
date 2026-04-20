import { Badge } from '@/components/common/Badge'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { RunAdhocActionButton } from '@/components/installs/management/RunAdhocAction'
import type { IActionRunMetadata } from '../types'

interface IActionRunMetadataPresentation extends IActionRunMetadata {
  orgId: string
}

export const ActionRunMetadata = ({
  actionRun,
  createdBy,
  step,
  orgId,
}: IActionRunMetadataPresentation) => {
  const isAdhocActionRun = actionRun?.trigger_type === 'adhoc'
  const firstStep = actionRun?.steps?.at(0)
  const adhocConfig = firstStep?.adhoc_config

  return (
    <div className="flex items-start justify-between gap-6">
      <div className="flex items-start gap-6">
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

        <LabeledValue label="Triggered by">
          <Badge size="md" variant="code">
            {isAdhocActionRun && createdBy ? ' ' + createdBy?.email : null}

            {!isAdhocActionRun ? actionRun?.triggered_by_type : null}
            {actionRun?.run_env_vars?.COMPONENT_ID ? (
              <Link
                href={`/${orgId}/installs/${step?.owner_id}/components/${actionRun?.run_env_vars?.COMPONENT_ID}`}
              >
                {actionRun?.run_env_vars?.COMPONENT_NAME}
              </Link>
            ) : null}
          </Badge>
        </LabeledValue>
      </div>

      <div className="flex flex-wrap gap-x-4 gap-y-1 items-center">
        {actionRun?.created_at ? (
          <Text theme="info" flex className="gap-1">
            <Icon variant="CalendarBlankIcon" />
            <Time variant="subtext" time={actionRun.created_at} />
          </Text>
        ) : null}
        {actionRun?.created_at ? (
          <Text theme="info" flex className="gap-1">
            <Icon variant="TimerIcon" />
            <Duration variant="subtext" beginTime={actionRun.created_at} endTime={actionRun.updated_at} />
          </Text>
        ) : null}
        {actionRun?.runner_job?.install_role_usage?.role_name ? (
          <Text theme="info" flex className="gap-1">
            <Icon variant="FileLockIcon" />
            <Text variant="subtext">{actionRun.runner_job.install_role_usage.role_name}</Text>
          </Text>
        ) : null}
      </div>

      {isAdhocActionRun ? (
        <div className="self-end">
          <RunAdhocActionButton
            initialValues={{
              name: adhocConfig?.name,
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
      ) : null}
    </div>
  )
}

export const ActionRunMetadataSkeleton = () => {
  return (
    <div className="flex items-start gap-6">
      <LabeledValue label={<Skeleton height="17px" width="34px" />}>
        <Skeleton height="23px" width="75px" />
      </LabeledValue>

      <LabeledValue label={<Skeleton height="17px" width="34px" />}>
        <Skeleton height="23px" width="162px" />
      </LabeledValue>
    </div>
  )
}
