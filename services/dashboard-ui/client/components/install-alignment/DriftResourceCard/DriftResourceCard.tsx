import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import type { TCollectionView } from '@/components/common/CollectionViewToggle'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ComponentType } from '@/components/components/ComponentType'
import type {
  TAPIError,
  TComponentType,
  TInstallAlignmentResource,
  TInstallAlignmentResourceKind,
} from '@/types'
import { cn } from '@/utils/classnames'
import {
  buildRef,
  driftExplanation,
  explanationEvidence,
  hasOngoingDeployment,
  isDrifted,
} from '../alignment-utils'
import { BuildRef } from '../LiveBuild'

export interface IDriftResourceCard {
  buildHref?: (buildId: string) => string | undefined
  deploymentHref?: string
  explanationHref?: string
  historyHref?: string
  metadata?: ReactNode
  nameHref?: string
  onResolve?: () => void
  onViewPlan?: () => void
  resolveError?: TAPIError | null
  resolving?: boolean
  resource: TInstallAlignmentResource
  view?: TCollectionView
}

const KIND_ICONS: Record<TInstallAlignmentResourceKind, TIconVariant> = {
  component: 'CubeIcon',
  image: 'CubeIcon',
  sandbox: 'ShippingContainerIcon',
  stack: 'StackIcon',
}

const REASON_STATUS_LABELS: Record<string, string> = {
  cancelled: 'Workflow cancelled',
  error: 'Deployment failed',
  'policy-failed': 'Denied by policy',
  'policy-skipped': 'Skipped by policy',
  'auto-skipped': 'No changes',
  'failed-pending-retry': 'Failed — awaiting retry',
  'awaiting-retry': 'Failed — awaiting retry',
  'approval-awaiting': 'Awaiting approval',
  'awaiting-approval': 'Awaiting approval',
  'approval-denied': 'Approval denied',
  'approval-expired': 'Approval expired',
  'user-skipped': 'Deploy skipped',
  'not-attempted': 'Blocked by earlier step',
  discarded: 'Blocked by earlier step',
  unknown: 'Outcome unknown',
}

const earlierAttemptsLabel = (count: number) =>
  `${count} earlier ${count === 1 ? 'attempt' : 'attempts'}`

export const DriftResourceCard = ({
  buildHref,
  deploymentHref,
  explanationHref,
  historyHref,
  metadata,
  nameHref,
  onResolve,
  onViewPlan,
  resolveError,
  resolving = false,
  resource,
  view = 'list',
}: IDriftResourceCard) => {
  const isGrid = view === 'grid'
  const openDeployment = hasOngoingDeployment(resource)
    ? resource.open_deployment
    : undefined
  const isImage = resource.kind === 'image'
  const ongoingOnly = !!openDeployment && !isDrifted(resource)
  const isStateDrift = resource.category === 'state-drift-detected'
  const explanation = ongoingOnly
    ? isImage
      ? 'An image sync is pending or in progress.'
      : 'A deployment is pending or in progress.'
    : driftExplanation(resource)
  const evidenceWorkflowId = explanationEvidence(resource)?.workflowId
  const reasonStatus = openDeployment
    ? (evidenceWorkflowId === openDeployment.id &&
        resource.explanation?.status) ||
      openDeployment.status
    : resource.explanation?.status
  const evidenceTime = openDeployment
    ? openDeployment.created_at
    : isStateDrift
      ? resource.plan?.created_at
      : evidenceWorkflowId &&
          evidenceWorkflowId === resource.blocking_attempt?.workflow_id
        ? resource.blocking_attempt?.created_at
        : undefined

  const expected = resource.expected_build
  const applied = resource.applied_build
  const showBuilds =
    !isStateDrift && !!expected?.id && expected.id !== applied?.id
  const consequence = showBuilds
    ? undefined
    : (isImage && !isStateDrift
        ? 'This image version hasn’t been synced.'
        : resource.reason) ||
      (isStateDrift
        ? 'Live resources differ from what was last applied.'
        : 'Saved changes haven’t taken effect.')
  const earlier = openDeployment
    ? (resource.earlier_attempts ?? 0) + (resource.blocking_attempt ? 1 : 0)
    : (resource.earlier_attempts ?? 0)

  const name = (
    <Text
      family="mono"
      weight="stronger"
      role="heading"
      level={3}
      className="break-all"
    >
      {nameHref ? (
        <Link href={nameHref} variant="inline" className="!font-mono">
          {resource.name}
        </Link>
      ) : (
        resource.name
      )}
    </Text>
  )

  const content = (
    <div className="flex flex-col gap-2">
      <div
        className={cn(
          'flex flex-col gap-2',
          !isGrid && 'sm:flex-row sm:items-start sm:justify-between sm:gap-4'
        )}
      >
        <div className="flex items-center gap-2 min-w-0">
          {resource.component_type ? (
            <ComponentType
              type={resource.component_type as TComponentType}
              displayVariant="icon-only"
              colorVariant="color"
              iconSize="16"
            />
          ) : (
            <Icon
              variant={KIND_ICONS[resource.kind ?? 'component']}
              size={16}
            />
          )}
          {name}
        </div>
        <div
          className={cn(
            'flex flex-wrap items-center gap-2 shrink-0',
            !isGrid && 'sm:justify-end'
          )}
        >
          {isStateDrift && !openDeployment ? (
            <Badge size="sm" theme="warn">
              Drift detected
            </Badge>
          ) : reasonStatus ? (
            <Status status={reasonStatus} variant="badge">
              {isImage && reasonStatus === 'error'
                ? 'Sync failed'
                : isImage && reasonStatus === 'cancelled'
                  ? 'Sync cancelled'
                  : REASON_STATUS_LABELS[reasonStatus]}
            </Status>
          ) : null}
          {onViewPlan ? (
            <Button variant="secondary" onClick={onViewPlan}>
              <Icon variant="FileTextIcon" size={16} />
              View plan
            </Button>
          ) : null}
          {openDeployment && deploymentHref ? (
            <Button variant="secondary" href={deploymentHref}>
              {isImage ? 'View sync' : 'View deployment'}
            </Button>
          ) : !openDeployment && isDrifted(resource) && onResolve ? (
            <Button
              variant="secondary"
              onClick={onResolve}
              disabled={resolving}
              aria-busy={resolving}
            >
              <Icon variant={resolving ? 'Loading' : 'PlayIcon'} size={16} />
              {resolving
                ? isImage
                  ? 'Starting sync'
                  : 'Starting deployment'
                : 'Resolve drift'}
            </Button>
          ) : null}
        </div>
      </div>

      {resolveError ? (
        <Banner theme="error" role="alert">
          <Text variant="subtext">
            Drift resolution failed.{' '}
            {resolveError.description || resolveError.error}
          </Text>
        </Banner>
      ) : null}

      {showBuilds ? (
        <div
          className="flex flex-wrap items-center gap-x-2 gap-y-1"
          data-testid="drift-builds"
        >
          <Text as="span" variant="subtext" theme="neutral">
            Live
          </Text>
          <BuildRef
            build={applied}
            href={applied?.id ? buildHref?.(applied.id) : undefined}
          />
          <Icon
            variant="ArrowRightIcon"
            size={14}
            className="text-cool-grey-600 dark:text-cool-grey-500"
          />
          <Text as="span" variant="subtext" theme="neutral">
            Wanted
          </Text>
          <BuildRef build={expected} href={buildHref?.(expected!.id!)} />
        </div>
      ) : (
        <Text variant="subtext" data-testid="drift-consequence">
          {consequence}
        </Text>
      )}

      <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
        <Text variant="subtext" data-testid="drift-explanation">
          {explanation}
        </Text>
        {explanationHref && evidenceWorkflowId && !openDeployment ? (
          <Link href={explanationHref}>
            {isStateDrift ? 'View drift scan' : 'View workflow'}
          </Link>
        ) : null}
        {evidenceTime ? (
          <Time
            time={evidenceTime}
            format="relative"
            variant="subtext"
            theme="neutral"
          />
        ) : null}
      </div>

      {earlier > 0 && !isStateDrift ? (
        <div
          className="flex flex-wrap items-center gap-x-3 gap-y-1"
          data-testid="drift-history"
        >
          <Text variant="subtext" theme="neutral">
            {earlierAttemptsLabel(earlier)}
            {applied?.id ? (
              <>
                {' since '}
                <Text as="span" family="mono" variant="subtext" theme="neutral">
                  {buildRef(applied)}
                </Text>
              </>
            ) : null}
          </Text>
          {historyHref ? (
            <Link href={historyHref}>View deploy history</Link>
          ) : null}
        </div>
      ) : null}
    </div>
  )

  const dataProps = {
    'data-testid': 'drift-resource-card',
    'data-resource-id': resource.id,
    'data-category': resource.category,
  }

  return isGrid ? (
    <Card className="!p-3 !gap-3 min-w-0 md:h-full" {...dataProps}>
      {content}
      {metadata}
    </Card>
  ) : (
    <div className="flex flex-col gap-3 py-3 min-w-0" {...dataProps}>
      {content}
      {metadata}
    </div>
  )
}
