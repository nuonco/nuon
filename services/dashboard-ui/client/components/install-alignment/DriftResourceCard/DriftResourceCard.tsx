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
  TInstallResourceSyncStatusOrigin,
  TInstallResourceSyncStatusResource,
  TInstallResourceSyncStatusResourceKind,
} from '@/types'
import { cn } from '@/utils/classnames'
import { hasOngoingDeployment, isDrifted } from '../alignment-utils'
import { LiveBuild } from '../LiveBuild'

export interface IDriftResourceCard {
  deploymentHref?: string
  driftCheckHref?: string
  historyHref?: string
  metadata?: ReactNode
  nameHref?: string
  onResolve?: () => void
  onViewPlan?: () => void
  originHref?: (origin?: TInstallResourceSyncStatusOrigin) => string | undefined
  resolveError?: TAPIError | null
  resolving?: boolean
  resource: TInstallResourceSyncStatusResource
  view?: TCollectionView
}

const KIND_ICONS: Record<TInstallResourceSyncStatusResourceKind, TIconVariant> = {
  component: 'CubeIcon',
  image: 'CubeIcon',
  sandbox: 'ShippingContainerIcon',
  stack: 'StackIcon',
}

const OPEN_STATUS_LABELS: Record<string, string> = {
  'failed-pending-retry': 'Failed — awaiting retry',
  'approval-awaiting': 'Awaiting approval',
}

export const DriftResourceCard = ({
  deploymentHref,
  driftCheckHref,
  historyHref,
  metadata,
  nameHref,
  onResolve,
  onViewPlan,
  originHref,
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
  const isStateDrift = resource.category === 'state-drift-detected'
  const summary = openDeployment
    ? isImage
      ? 'An image sync is pending or in progress.'
      : 'A deployment is pending or in progress.'
    : isStateDrift
      ? 'Live resources differ from what was last applied.'
      : isImage && resource.expected_build_id !== resource.applied_build_id
        ? 'This image version hasn’t been synced.'
        : resource.reason || 'Saved changes haven’t taken effect.'
  const summaryTime = openDeployment
    ? openDeployment.created_at
    : isStateDrift
      ? resource.plan?.created_at
      : undefined

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
          ) : openDeployment ? (
            <Status status={openDeployment.status} variant="badge">
              {OPEN_STATUS_LABELS[openDeployment.status ?? '']}
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

      {resource.applied_build_id ? (
        <LiveBuild
          appliedAt={resource.applied_deploy_at}
          origin={resource.applied_by}
          originHref={originHref?.(resource.applied_by)}
          variant={isImage ? 'sync' : 'deploy'}
        />
      ) : openDeployment ? (
        <Text variant="subtext" theme="neutral">
          {isImage ? 'Not synced yet' : 'Not deployed yet'}
        </Text>
      ) : null}

      <div
        className="flex flex-wrap items-center gap-x-3 gap-y-1"
        data-testid="drift-summary"
      >
        <Text variant="subtext">{summary}</Text>
        {summaryTime ? (
          <Time
            time={summaryTime}
            format="relative"
            variant="subtext"
            theme="neutral"
          />
        ) : null}
        {isStateDrift && driftCheckHref ? (
          <Link href={driftCheckHref}>View drift scan</Link>
        ) : null}
        {!isStateDrift && historyHref ? (
          <Link href={historyHref}>
            {isImage ? 'View sync history' : 'View deploy history'}
          </Link>
        ) : null}
      </div>
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
