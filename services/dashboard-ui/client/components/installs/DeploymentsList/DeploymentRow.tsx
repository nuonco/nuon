import type { ReactNode } from 'react'
import { Link } from '@/components/common/Link'
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
import type { TWorkflow } from '@/types'
import { DeploymentViewDetails } from './DeploymentActions'

interface IDeploymentRow {
  run: TDeploymentRun
  title: string
  createdAt: string
  href: string
  onViewDetails: () => void
  workflow?: TWorkflow
  children?: ReactNode
  scale?: 'auto' | 'comfortable' | 'compact'
  history?: boolean
}

export const DeploymentRow = ({
  run,
  title,
  createdAt,
  href,
  onViewDetails,
  workflow,
  children,
  scale = 'auto',
  history = false,
}: IDeploymentRow) => {
  const awaiting = run.steps.some(isAwaitingDeploymentApproval)
  if (history) {
    return (
      <article aria-label={title} className="relative flex gap-3 border-l">
        <Status
          status={run.status}
          variant="timeline"
          isWithoutText
          iconSize={18}
          className="-ml-3.5 mt-3 self-start bg-background"
        />
        <div className="grid min-w-0 flex-1 grid-cols-1 items-center gap-3 border-b py-3 @4xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
          <div className="flex min-w-0 flex-col gap-1">
            <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
              <Link href={href} className="min-w-0 break-words">
                <Text weight="strong" className="break-words">
                  {title}
                </Text>
              </Link>
              <Status status={run.status} />
              <Time
                time={createdAt}
                format="relative"
                variant="subtext"
                theme="neutral"
              />
            </div>
            {run.status === 'error' && run.activity ? (
              <Text
                variant="subtext"
                theme="error"
                className="whitespace-pre-wrap break-words"
              >
                {run.activity}
              </Text>
            ) : null}
            {children}
          </div>
          <ResourceScopeSummary run={run} />
          <DeploymentViewDetails
            className="justify-self-end"
            workflow={workflow}
            onViewDetails={onViewDetails}
          />
        </div>
      </article>
    )
  }
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
              <Link href={href} className="min-w-0 break-words">
                <Text variant="h3" weight="strong">
                  {title}
                </Text>
              </Link>
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
            <DeploymentViewDetails
              workflow={workflow}
              onViewDetails={onViewDetails}
              size="lg"
              showArrow
            />
          </div>
        </div>
      </div>
    </article>
  )
}
