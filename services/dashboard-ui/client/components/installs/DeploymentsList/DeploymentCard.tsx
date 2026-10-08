import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import {
  resourceCategories,
  ResourceCategoryIcon,
  StepProgress,
  type TDeploymentRun,
} from '@/components/installs/DeploymentDetail/DeploymentProgress'
import {
  deploymentStepContext,
  isAwaitingDeploymentApproval,
  isAwaitingDeploymentRetry,
  isDeploymentRunning,
} from '@/components/installs/DeploymentDetail/deployment-progress'

export const DeploymentCard = ({
  run,
  title,
  typeLabel,
  createdAt,
  href,
  onViewDetails,
  children,
}: {
  run: TDeploymentRun
  title: string
  typeLabel: string
  createdAt: string
  href: string
  onViewDetails: () => void
  children?: ReactNode
}) => {
  const awaiting =
    run.status === 'approval-awaiting' ||
    run.steps.some(isAwaitingDeploymentApproval)
  const retry = isAwaitingDeploymentRetry(run.status)
  const { current, next } = deploymentStepContext(run.status, run.steps)
  const status = awaiting ? 'approval-awaiting' : run.status
  const statusLabel = awaiting
    ? 'Pending approval'
    : retry
      ? 'Failed pending retry'
      : isDeploymentRunning(run.status)
        ? 'Running now'
        : 'Queued'

  return (
    <Card role="article" aria-label={title} className="@container gap-4 p-4">
      <div className="flex items-start gap-3">
        <Status
          status={status}
          variant="timeline"
          isWithoutText
          iconSize={20}
        />
        <div className="flex min-w-0 flex-1 flex-col gap-2">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <Link href={href} className="min-w-0 break-words">
              <Text weight="strong" className="break-words">
                {title}
              </Text>
            </Link>
            <Badge size="sm" theme="neutral">
              {typeLabel}
            </Badge>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Status status={status}>{statusLabel}</Status>
            <span aria-hidden="true">·</span>
            <Time
              time={createdAt}
              format="relative"
              variant="subtext"
              theme="neutral"
            />
          </div>
        </div>
      </div>
      <div className="grid min-w-0 grid-cols-1 gap-2 @sm:grid-cols-2">
        <div role="group" aria-label="Current step">
          <Text
            variant="subtext"
            theme={retry ? 'error' : 'default'}
            className="break-words"
          >
            {awaiting
              ? 'Waiting for approval: '
              : retry
                ? 'Stopped: '
                : 'Now: '}
            {current?.name ||
              (!run.steps.length ? 'Preparing steps' : run.activity)}
            {retry && current ? ' failed — retry or skip' : null}
          </Text>
        </div>
        <div
          role="group"
          aria-label="Next step"
          className="@sm:border-l @sm:pl-3"
        >
          <Text variant="subtext" theme="neutral" className="break-words">
            Next:{' '}
            {next?.name ||
              (run.steps.length ? 'No remaining steps' : 'Waiting for steps')}
          </Text>
        </div>
      </div>
      {children}
      <StepProgress run={run} />
      <div className="flex flex-wrap items-center justify-between gap-3 border-t pt-3">
        <div
          role="group"
          aria-label="Resource outcomes"
          className="flex flex-wrap gap-x-4 gap-y-2"
        >
          {resourceCategories(run).map(({ category, status, label }) => (
            <span
              key={category}
              role="group"
              aria-label={`${category}: ${label}`}
              className="flex items-center gap-2"
            >
              <ResourceCategoryIcon status={status} size={18} />
              <span className="flex flex-col">
                <Text variant="subtext">{category}</Text>
                <Text
                  variant="subtext"
                  theme={status === 'error' ? 'error' : 'neutral'}
                >
                  {label}
                </Text>
              </span>
            </span>
          ))}
        </div>
        <Button variant="secondary" onClick={onViewDetails}>
          View details
        </Button>
      </div>
    </Card>
  )
}
