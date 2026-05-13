import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import {
  ProcessCard,
  ProcessCardSkeleton,
} from '@/components/runners/ProcessCard'
import type { TRunnerProcess, TRunnerSettings, TWorkflowStep } from '@/types'

export interface IRunnerStepDetails {
  step?: TWorkflowStep
  orgId: string
  processes: TRunnerProcess[]
  processesLoading: boolean
  settings?: TRunnerSettings
}

export const RunnerStepDetails = ({
  step,
  orgId,
  processes,
  processesLoading,
  settings,
}: IRunnerStepDetails) => {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-4">
        <Text variant="base" weight="strong">
          Install runner
        </Text>

        <Text variant="subtext">
          <Link href={`/${orgId}/installs/${step?.owner_id}/runner`}>
            View runner <Icon variant="CaretRightIcon" />
          </Link>
        </Text>
      </div>

      {processesLoading ? (
        <div className="@container">
          <div className="grid grid-cols-1 @4xl:grid-cols-2 gap-6">
            <ProcessCardSkeleton />
            <ProcessCardSkeleton />
          </div>
        </div>
      ) : processes.length === 0 ? (
        <Card>
          <EmptyState
            emptyTitle="No active processes"
            emptyMessage="No runner processes are currently active or offline."
            variant="table"
          />
        </Card>
      ) : processes.length === 1 ? (
        <ProcessCard
          process={processes[0]}
          settings={settings}
          shouldPoll
        />
      ) : (
        <div className="@container">
          <div className="grid grid-cols-1 @4xl:grid-cols-2 gap-6">
            {processes.map((process) => (
              <ProcessCard
                key={process.id}
                process={process}
                settings={settings}
                shouldPoll
              />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
