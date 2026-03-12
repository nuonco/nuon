'use client'

import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { RunAdhocActionButton } from '@/components/installs/management/RunAdhocAction'
import { useOrg } from '@/hooks/use-org'
import type { IActionRunMetadata } from './types'

export const ActionRunMetadata = ({
  actionRun,
  createdBy,
  step,
}: IActionRunMetadata) => {
  const { org } = useOrg()
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
                href={`/${org.id}/installs/${step.owner_id}/components/${actionRun?.run_env_vars?.COMPONENT_ID}`}
              >
                {actionRun?.run_env_vars?.COMPONENT_NAME}
              </Link>
            ) : null}
          </Badge>
        </LabeledValue>

        {actionRun?.role ? (
          <LabeledValue label="Role used">
            <Link
              className="text-xs"
              href={`/${org.id}/installs/${step.owner_id}/roles#${actionRun?.role}`}
            >
              {actionRun?.role}
            </Link>
          </LabeledValue>
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
