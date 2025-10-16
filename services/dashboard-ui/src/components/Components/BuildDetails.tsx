'use client'

import { ClickToCopy } from '@/components/ClickToCopy'
import { CancelRunnerJobButton } from '@/components/CancelRunnerJobButton'
import { RunnerJobPlanModal } from '@/components/OldRunners/RunnerJobPlanModal'
import { StatusBadge } from '@/components/Status'
import { ToolTip } from '@/components/ToolTip'
import { Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TBuild } from '@/types'

export interface IBuildDetails extends IPollingProps {
  initBuild: TBuild
}

export const BuildDetails = ({
  initBuild,
  pollInterval = 5000,
  shouldPoll = false,
}: IBuildDetails) => {
  const { org } = useOrg()
  const { data: build } = usePolling<TBuild>({
    initData: initBuild,
    path: `/api/orgs/${org.id}/components/${initBuild?.component_id}/builds/${initBuild?.id}`,
    pollInterval,
    shouldPoll,
  })

  const status = build?.status_v2 || {
    status: build?.status || 'Unknown',
    status_human_description: build?.status_description || undefined,
  }

  return (
    <div className="flex gap-6 items-start justify-start">
      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Status
        </Text>
        <StatusBadge
          description={status?.status_human_description}
          status={status?.status}
          descriptionAlignment="right"
        />
      </span>

      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Component
        </Text>
        <Text variant="med-12">{initBuild?.component_name}</Text>
        <Text variant="mono-12">
          <ToolTip alignment="right" tipContent={build.component_id}>
            <ClickToCopy>{build.component_id}</ClickToCopy>
          </ToolTip>
        </Text>
      </span>

      {build?.runner_job ? (
        <RunnerJobPlanModal runnerJobId={build?.runner_job?.id} />
      ) : null}

      {build?.runner_job &&
      build?.status_v2?.status !== 'active' &&
      build?.status_v2?.status !== 'error' ? (
        <CancelRunnerJobButton
          jobType="build"
          runnerJobId={build?.runner_job?.id}
          orgId={org.id}
        />
      ) : null}
    </div>
  )
}
