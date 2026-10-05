import type { ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import {
  ResourceScopeSummary,
  StepContext,
  StepProgress,
  type TDeploymentRun,
} from '@/components/installs/DeploymentDetail/DeploymentProgress'
import {
  isAwaitingDeploymentApproval,
  isDeploymentRunning,
} from '@/components/installs/DeploymentDetail/deployment-progress'

interface IDeploymentRow {
  run: TDeploymentRun
  title: string
  createdAt: string
  onViewDetails: () => void
  children?: ReactNode
  scale?: 'auto' | 'comfortable' | 'compact'
}

export const DeploymentRow = ({
  run,
  title,
  createdAt,
  onViewDetails,
  children,
  scale = 'auto',
}: IDeploymentRow) => {
  const awaiting = run.steps.some(isAwaitingDeploymentApproval)
  return (
    <article
      aria-label={title}
      style={{ containerType: 'inline-size', containerName: 'deployment' }}
      className={`deployment-row relative border-l pl-6 density-${scale}`}
    >
      <Status
        status={run.status}
        variant="timeline"
        iconSize={22}
        className="absolute -left-4 top-4 bg-background"
      />
      <div className="flex flex-col gap-4 border-b py-4">
        <div className="density-header">
          <div className="density-identity flex min-w-0 flex-col gap-3 break-words">
            <div className="density-run-heading">
              <Text variant="h3" weight="strong">
                {title}
              </Text>
              <div className="density-run-meta flex flex-wrap items-center gap-2">
                <Status
                  status={awaiting ? 'approval-awaiting' : run.status}
                  className="!text-sm [&>span:first-child]:size-2"
                >
                  {isDeploymentRunning(run.status) ? (
                    <Text variant="body" theme={awaiting ? 'warn' : 'info'}>
                      {awaiting ? 'Pending approval' : 'Running now'}
                    </Text>
                  ) : undefined}
                </Status>
                <span aria-hidden="true">·</span>
                <Time
                  time={createdAt}
                  format="relative"
                  variant="body"
                  theme="neutral"
                  className="whitespace-nowrap"
                />
              </div>
            </div>
            {run.status === 'error' ? (
              <Text
                variant="base"
                theme="error"
                className="whitespace-pre-wrap break-words"
              >
                {run.activity}
              </Text>
            ) : null}
            <div className="flex min-w-0 flex-col gap-3">
              <StepContext run={run} />
              {run.steps.length > 0 && isDeploymentRunning(run.status) ? (
                <footer>
                  <StepProgress run={run} />
                </footer>
              ) : null}
            </div>
            {children}
          </div>
          <div className="density-resources">
            <ResourceScopeSummary run={run} />
          </div>
          <div className="density-action">
            <Button size="lg" variant="secondary" onClick={onViewDetails}>
              View details <Icon variant="ArrowRightIcon" size={18} />
            </Button>
          </div>
        </div>
      </div>
    </article>
  )
}
