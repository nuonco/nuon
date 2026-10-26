import { useId, type ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import type { TCollectionView } from '@/components/common/CollectionViewToggle'
import { Expand } from '@/components/common/Expand'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link, LINK_COLOR_CLASSES } from '@/components/common/Link'
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
  driftExplanation,
  explanationEvidence,
  hasOngoingDeployment,
  isDrifted,
  shortSha,
} from '../alignment-utils'

export interface IDriftResourceCard {
  branchRunHref?: string
  deploymentHref?: string
  explanationHref?: string
  lastAttemptHref?: string
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

const newerRunsLabel = (count?: number | null) =>
  count ? `${count} newer branch ${count === 1 ? 'run' : 'runs'}` : null

const attemptLabel = (type?: string) =>
  type === 'input_update'
    ? 'Last input update'
    : type === 'component_enabled'
      ? 'Last enable attempt'
      : type === 'component_disabled'
        ? 'Last disable attempt'
        : 'Last attempt'

const DETAIL_ROW = '!flex-row flex-wrap items-center !gap-1.5'

const REASON_STATUS_LABELS: Record<string, string> = {
  cancelled: 'Workflow cancelled',
  error: 'Deployment failed',
  'failed-pending-retry': 'Failed — awaiting retry',
  'awaiting-retry': 'Failed — awaiting retry',
  'approval-awaiting': 'Awaiting approval',
  'awaiting-approval': 'Awaiting approval',
  'user-skipped': 'Deploy skipped',
  'not-attempted': 'Blocked by earlier step',
  discarded: 'Blocked by earlier step',
}

export const DriftResourceCard = ({
  branchRunHref,
  deploymentHref,
  explanationHref,
  lastAttemptHref,
  metadata,
  nameHref,
  onResolve,
  onViewPlan,
  resolveError,
  resolving = false,
  resource,
  view = 'list',
}: IDriftResourceCard) => {
  const detailsId = useId()
  const isGrid = view === 'grid'
  const openDeployment = hasOngoingDeployment(resource)
    ? resource.open_deployment
    : undefined
  const lastAttempt = resource.last_attempt
  const branchRun = lastAttempt?.branch_run
  const newerRuns = newerRunsLabel(resource.newer_branch_runs)
  const isImage = resource.kind === 'image'
  const ongoingOnly = !!openDeployment && !isDrifted(resource)
  const explanation = ongoingOnly
    ? isImage
      ? 'Open the sync to view its progress.'
      : 'Open the deployment to view its progress.'
    : driftExplanation(resource)
  const evidenceWorkflowId = explanationEvidence(resource)?.workflowId
  const isStateDrift = resource.category === 'state-drift-detected'
  const consequence =
    (ongoingOnly
      ? isImage
        ? 'An image sync is pending or in progress.'
        : 'A deployment is pending or in progress.'
      : isImage
        ? 'This image version hasn’t been synced.'
        : resource.reason) ||
    (isStateDrift
      ? 'Live resources differ from what was last applied.'
      : 'Saved changes haven’t taken effect.')
  const reasonStatus = openDeployment
    ? (evidenceWorkflowId === openDeployment.id &&
        resource.explanation?.status) ||
      openDeployment.status
    : resource.explanation?.status
  const evidenceAttempt =
    evidenceWorkflowId === openDeployment?.id
      ? openDeployment
      : evidenceWorkflowId && evidenceWorkflowId === lastAttempt?.workflow_id
        ? lastAttempt
        : undefined

  const showLastAttempt =
    !!lastAttempt?.status &&
    !!lastAttemptHref &&
    lastAttempt.workflow_id !== evidenceWorkflowId
  const showBranchRun = !!(branchRun?.sha || branchRun?.workflow_id)

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
    <div className="flex flex-col gap-1">
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

      <Expand
        id={detailsId}
        headerClassName={cn(
          LINK_COLOR_CLASSES,
          '!p-1 !justify-start text-left rounded-md [&>svg]:shrink-0 focus-visible:[--tw-outline-style:solid] focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-primary-400/80'
        )}
        heading={
          <Text
            variant="subtext"
            className="underline underline-offset-2"
            data-testid="drift-consequence"
          >
            {consequence}
          </Text>
        }
      >
        <div className="flex flex-col gap-3 pt-3" data-testid="drift-details">
          <LabeledValue
            label={
              ongoingOnly ? 'Progress' : isStateDrift ? 'Detected by' : 'Why'
            }
          >
            <Text variant="subtext" data-testid="drift-explanation">
              {explanation}
            </Text>
          </LabeledValue>
          {explanationHref && evidenceWorkflowId ? (
            <LabeledValue
              label={isStateDrift ? 'Drift scan' : 'Related workflow'}
            >
              <span className="flex flex-wrap items-center gap-3">
                <Link href={explanationHref}>
                  {isStateDrift ? 'View drift scan' : 'View workflow'}
                </Link>
                {evidenceAttempt?.created_at ? (
                  <Text variant="subtext" theme="neutral">
                    Started{' '}
                    <Time
                      time={evidenceAttempt.created_at}
                      format="relative"
                      variant="subtext"
                      theme="neutral"
                    />
                  </Text>
                ) : null}
              </span>
            </LabeledValue>
          ) : null}
          {showLastAttempt ? (
            <LabeledValue
              label={attemptLabel(lastAttempt?.type)}
              className={DETAIL_ROW}
            >
              <span className="flex flex-wrap items-center gap-2">
                <Link href={lastAttemptHref}>
                  <Status status={lastAttempt!.status} />
                </Link>
                {lastAttempt?.created_at ? (
                  <Text variant="subtext" theme="neutral">
                    Started{' '}
                    <Time
                      time={lastAttempt.created_at}
                      format="relative"
                      variant="subtext"
                      theme="neutral"
                    />
                  </Text>
                ) : null}
              </span>
            </LabeledValue>
          ) : null}
          {showBranchRun ? (
            <LabeledValue
              label="Last attempted branch run"
              className={DETAIL_ROW}
            >
              <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
                {branchRunHref ? (
                  <Link
                    href={branchRunHref}
                    className={cn(
                      'inline-flex items-center gap-2',
                      branchRun?.sha && '!font-mono'
                    )}
                  >
                    <Icon variant="GitCommitIcon" size={14} />
                    {shortSha(branchRun?.sha) ?? 'View run'}
                  </Link>
                ) : branchRun?.sha ? (
                  <Text as="span" family="mono" variant="subtext">
                    {shortSha(branchRun.sha)}
                  </Text>
                ) : null}
                {newerRuns ? (
                  <Text variant="subtext" theme="neutral">
                    {newerRuns}
                  </Text>
                ) : null}
              </span>
            </LabeledValue>
          ) : null}
        </div>
      </Expand>
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
