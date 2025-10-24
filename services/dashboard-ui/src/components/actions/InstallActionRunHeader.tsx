'use client'

import { BackLink } from '@/components/common/BackLink'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useInstall } from '@/hooks/use-install'
import { useInstallActionRun } from '@/hooks/use-install-action-run'
import { useOrg } from '@/hooks/use-org'
import type { TWorkflow } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'
import { getWorkflowStep } from '@/utils/workflow-utils'

// NOTE: old components
import { RunnerJobPlanModal } from '@/components/old/OldRunners'
import { InstallWorkflowCancelModal } from '@/components/old/InstallWorkflows'

interface IInstallActionRunHeader {
  actionId: string
  actionName: string
  workflow: TWorkflow
}

export const InstallActionRunHeader = ({
  actionId,
  actionName,
  workflow,
}: IInstallActionRunHeader) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { installActionRun } = useInstallActionRun()
  const step = getWorkflowStep({
    workflow,
    stepTargetId: installActionRun?.id,
  })

  return (
    <header className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center gap-4 justify-between w-full">
        <HeadingGroup className="">
          <BackLink className="mb-4" />
          <Text
            className="inline-flex items-center gap-4"
            variant="h3"
            weight="strong"
          >
            {actionName}
          </Text>
          <ID>{actionId}</ID>
          <Time
            time={installActionRun?.updated_at}
            format="relative"
            variant="subtext"
            theme="info"
          />
        </HeadingGroup>

        <div className="flex items-center gap-4">
          <InstallWorkflowCancelModal installWorkflow={workflow} />
          <RunnerJobPlanModal runnerJobId={installActionRun?.runner_job?.id} />
        </div>
      </div>

      <Card>
        <div className="grid grid-cols-4">
          <LabeledValue
            label={`Triggered via ${installActionRun?.triggered_by_type}`}
          >
            {installActionRun?.created_by?.email || (
              <ID theme="default">{installActionRun?.created_by_id}</ID>
            )}
          </LabeledValue>

          <LabeledStatus
            label="Status"
            statusProps={{
              status: installActionRun?.status_v2?.status,
            }}
            tooltipProps={{
              tipContent: toSentenceCase(
                installActionRun?.status_v2?.status_human_description
              ),
            }}
          />

          <LabeledValue label={`Total duration`}>
            <Duration nanoseconds={installActionRun?.execution_time} />
          </LabeledValue>
        </div>
      </Card>

      {workflow ? (
        <Button
          href={`/${org.id}/installs/${install.id}/workflows/${workflow.id}?target=${step?.id}`}
        >
          View workflow
          <Icon variant="CaretRightIcon" />
        </Button>
      ) : null}
    </header>
  )
}
