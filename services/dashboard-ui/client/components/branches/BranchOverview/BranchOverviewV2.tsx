import { useEffect, type ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CompositeError } from '@/components/common/CompositeError'
import { Duration } from '@/components/common/Duration'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { Status } from '@/components/common/Status'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { DeploymentStageRail } from '@/components/branches/DeploymentStageRail'
import {
  GroupApproveAllButton,
  InstallApprovalButton,
} from '@/components/branches/InstallGroupCards/InstallApprovalActions'
import { InstallStatusCounts } from '@/components/branches/InstallGroupCards/InstallStatusCounts'
import { InstallStatusTiles } from '@/components/branches/InstallGroupCards/InstallStatusTiles'
import {
  installStatusBucket,
  statusSlices,
  withQueuedInstalls,
} from '@/components/branches/InstallGroupCards/install-status'
import { stepStatusCategory } from '@/components/branches/shared/step-status'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { cn } from '@/utils/classnames'
import type { IBranchOverview } from './BranchOverview'
import { GroupApprovedVersion } from './GroupApprovedVersion'
import { OverviewLoadingTrack } from './OverviewLoadingTrack'
import type { TGroupApprovedVersion } from './group-approved-version'
import type { IGroupPlanApproval } from './RolloutGroupsCard'
import type { TTrackGroup, TTrackInstall } from './RolloutTrack'
import { RunCommitSummary, RunSourceMark } from './RunSourceCard'

const groupAnchorId = (groupId: string) => `deployment-group-${groupId}`

const installCountLabel: Record<string, string> = {
  success: 'completed',
  error: 'errored',
  'in-progress': 'in progress',
  cancelled: 'cancelled',
  pending: 'pending',
  'approval-awaiting': 'awaiting approval',
}

const groupColumns =
  'items-center gap-3 px-4 md:grid-cols-[minmax(8rem,1.1fr)_minmax(8rem,0.9fr)_minmax(6rem,0.7fr)_minmax(12rem,1.6fr)_7rem]'

const approvalFor = (
  approvals: IGroupPlanApproval[] | undefined,
  name: string
) =>
  approvals?.find((item) => item.groupName.toLowerCase() === name.toLowerCase())

const completedInstalls = (group: TTrackGroup) =>
  group.installs.filter(
    (install) => installStatusBucket(install.status) === 'success'
  ).length

const installTotal = (group: TTrackGroup) =>
  Math.max(group.plannedCount ?? 0, group.installs.length)

const scrollToGroup = (groupId: string) => {
  const reduceMotion = window.matchMedia(
    '(prefers-reduced-motion: reduce)'
  ).matches
  document.getElementById(groupAnchorId(groupId))?.scrollIntoView({
    behavior: reduceMotion ? 'auto' : 'smooth',
    block: 'center',
  })
}

const groupMatchLabel = (group: TTrackGroup) => {
  if (group.match?.kind === 'default') return 'Default group'
  const labels = Object.entries(group.match?.labels ?? {})
  return labels.length
    ? labels.map(([key, value]) => `${key}=${value}`).join(', ')
    : undefined
}

const GroupDrawerHeading = ({ group }: { group: TTrackGroup }) => {
  const label = groupMatchLabel(group)
  return (
    <div className="flex flex-col gap-0.5">
      <Text variant="subtext" theme="neutral">
        Group information
      </Text>
      <span className="flex items-baseline gap-2">
        <Text variant="base" weight="strong" family="mono">
          {group.name}
        </Text>
        {label ? (
          <Text variant="subtext" theme="neutral" family="mono">
            {label}
          </Text>
        ) : null}
      </span>
    </div>
  )
}

const GroupInstallList = ({
  group,
  onSelectInstall,
}: {
  group: TTrackGroup
  onSelectInstall?: (install: TTrackInstall) => void
}) => {
  const failedInstallId = (
    group.error?.data as unknown as { install_id?: string } | undefined
  )?.install_id
  const failedInList = group.installs.some(
    (install) => install.id === failedInstallId
  )

  if (!group.installs.length) {
    return (
      <Text variant="subtext" theme="neutral">
        This group has not started deploying.
      </Text>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      {group.error && !failedInList && group.error.message ? (
        <Text variant="subtext" theme="error">
          {group.error.message}
        </Text>
      ) : null}
      <ul className="flex flex-col divide-y overflow-hidden rounded-md border">
        {group.installs.map((install) => {
          const error =
            install.id === failedInstallId ? group.error?.message : undefined
          return (
            <li key={install.id} className="flex flex-col gap-2 px-4 py-3">
              <div className="flex flex-wrap items-center justify-between gap-3">
                {install.overviewHref ? (
                  <Link
                    href={install.overviewHref}
                    className="min-w-0 truncate font-mono"
                  >
                    {install.name}
                  </Link>
                ) : (
                  <Text family="mono" className="min-w-0 truncate">
                    {install.name}
                  </Text>
                )}
                <span className="flex items-center gap-3">
                  <Status status={install.status} />
                  {installStatusBucket(install.status) ===
                  'approval-awaiting' ? (
                    <InstallApprovalButton install={install} label="Approve" />
                  ) : null}
                  {onSelectInstall && install.workflowId ? (
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => onSelectInstall(install)}
                    >
                      View details
                    </Button>
                  ) : null}
                </span>
              </div>
              {error ? (
                <Text variant="subtext" theme="error">
                  {error}
                </Text>
              ) : install.detail ? (
                <Text variant="subtext" theme="neutral">
                  {install.detail}
                </Text>
              ) : null}
            </li>
          )
        })}
      </ul>
    </div>
  )
}

const GroupDrawerBody = ({
  group,
  configChanges,
  onSelectInstall,
  onApproveAll,
}: {
  group: TTrackGroup
  configChanges?: ReactNode
  onSelectInstall?: (install: TTrackInstall) => void
  onApproveAll?: (groupId: string) => void
}) => (
  <Tabs
    className="gap-4"
    naturalHeight
    initActiveTab="installs"
    actions={
      <GroupApproveAllButton
        groupName={group.name}
        installs={group.installs}
        onApproved={() => onApproveAll?.(group.id)}
      />
    }
    tabs={{
      changes: configChanges ?? (
        <Text variant="subtext" theme="neutral">
          No config changes for this run.
        </Text>
      ),
      installs: (
        <GroupInstallList group={group} onSelectInstall={onSelectInstall} />
      ),
    }}
  />
)

const GroupDetails = ({
  group,
  configChanges,
  configChangesKey,
  onSelectInstall,
  onApproveAll,
}: {
  group: TTrackGroup
  configChanges?: ReactNode
  configChangesKey?: string
  onSelectInstall?: (install: TTrackInstall) => void
  onApproveAll?: (groupId: string) => void
}) => {
  const { panels, updatePanel } = useSurfaces()
  const panelKey = `install-group-${group.id}`
  const openPanelId =
    panels.find((panel) => panel.key === panelKey && panel.isVisible)?.id ??
    null
  const errorMessage = group.error?.message
  const installSignature = group.installs
    .map((install) => `${install.id}:${install.status}:${install.detail ?? ''}`)
    .join(',')
  const heading = <GroupDrawerHeading group={group} />
  const body = (
    <GroupDrawerBody
      group={group}
      configChanges={configChanges}
      onSelectInstall={onSelectInstall}
      onApproveAll={onApproveAll}
    />
  )

  useEffect(() => {
    if (!openPanelId) return
    updatePanel(
      openPanelId,
      <Panel panelKey={panelKey} size="half" heading={heading}>
        {body}
      </Panel>
    )
  }, [
    openPanelId,
    errorMessage,
    installSignature,
    configChangesKey,
    updatePanel,
  ])

  return (
    <Panel
      panelKey={panelKey}
      size="half"
      heading={heading}
      triggerButton={{
        children: 'View details',
        size: 'sm',
        variant: 'secondary',
        id: `group-details-${group.id}`,
      }}
    >
      {body}
    </Panel>
  )
}

const ActiveGroup = ({
  group,
  index,
  total,
  approval,
  onSelectInstall,
  onApproveAll,
}: {
  group: TTrackGroup
  index: number
  total: number
  approval?: ReactNode
  onSelectInstall?: (install: TTrackInstall) => void
  onApproveAll?: (groupId: string) => void
}) => {
  const installs = withQueuedInstalls(group.installs, group.plannedCount)
  const completed = completedInstalls(group)
  const planned = installTotal(group)
  const category = stepStatusCategory(group.status)
  const deploying = group.installs.find(
    (install) => installStatusBucket(install.status) === 'in-progress'
  )

  return (
    <Card elevation="1" className="!gap-4 !p-4" aria-current="step">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex min-w-0 flex-col gap-1">
          <Text
            variant="subtext"
            weight="strong"
            theme="info"
            aria-live="polite"
          >
            {category === 'awaiting'
              ? 'Waiting for approval'
              : category === 'error'
                ? 'Rollout failed'
                : 'Deploying'}
            <Text as="span" variant="subtext" theme="neutral">
              {' '}
              · Group {index + 1} of {total}
            </Text>
          </Text>
          <Text variant="body" weight="strong">
            {group.name}
          </Text>
        </div>
        <Status status={group.status} />
      </div>
      {approval}
      <div className="flex flex-col gap-2">
        <span className="flex flex-wrap items-baseline gap-x-3 gap-y-1">
          <Text variant="subtext">
            {completed} of {planned} installs completed
          </Text>
          {group.startedAt ? (
            <Text variant="subtext" theme="neutral">
              Elapsed{' '}
              <Duration beginTime={group.startedAt} unitDisplay="short" />
            </Text>
          ) : null}
        </span>
        <InstallStatusTiles
          installs={installs}
          variant="rail"
          onSelect={(install) => {
            const match = group.installs.find((item) => item.id === install.id)
            if (match?.workflowId) onSelectInstall?.(match)
          }}
        />
        <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
          <InstallStatusCounts installs={installs} showTotal={false} />
        </span>
      </div>
      <div className="flex flex-wrap items-center justify-between gap-3">
        {deploying ? (
          <Text variant="subtext" theme="neutral">
            Currently deploying{' '}
            <Text as="span" family="mono" variant="subtext">
              {deploying.name}
            </Text>
          </Text>
        ) : (
          <span />
        )}
        <span className="flex flex-wrap items-center gap-2">
          <GroupApproveAllButton
            groupName={group.name}
            installs={group.installs}
            onApproved={() => onApproveAll?.(group.id)}
          />
          <Button
            size="sm"
            variant="secondary"
            onClick={() =>
              document.getElementById(`group-details-${group.id}`)?.click()
            }
          >
            View group details
          </Button>
        </span>
      </div>
    </Card>
  )
}

const GroupColumnHeader = () => (
  <li className={cn('hidden py-2 md:grid', groupColumns)} aria-hidden="true">
    <Text variant="subtext" theme="neutral">
      Group
    </Text>
    <Text variant="subtext" theme="neutral">
      Current config
    </Text>
    <Text variant="subtext" theme="neutral">
      Status
    </Text>
    <Text variant="subtext" theme="neutral">
      Installs
    </Text>
    <span />
  </li>
)

const GroupRow = ({
  group,
  current,
  pending,
  approvedVersion,
  approvedVersionsLoading,
  configChanges,
  configChangesKey,
  onSelectInstall,
  onApproveAll,
}: {
  group: TTrackGroup
  current?: boolean
  pending?: boolean
  approvedVersion?: TGroupApprovedVersion
  approvedVersionsLoading?: boolean
  configChanges?: ReactNode
  configChangesKey?: string
  onSelectInstall?: (install: TTrackInstall) => void
  onApproveAll?: (groupId: string) => void
}) => {
  const installs = withQueuedInstalls(group.installs, group.plannedCount)

  return (
    <li
      id={groupAnchorId(group.id)}
      className={cn('grid grid-cols-1 py-3', groupColumns)}
      aria-current={current ? 'step' : undefined}
      data-current={current ? 'true' : undefined}
    >
      <span className="flex min-w-0 items-baseline gap-2">
        <Text variant="body" weight="strong" className="min-w-0 truncate">
          {group.name}
        </Text>
        {current ? (
          <Text
            variant="subtext"
            weight="strong"
            theme="info"
            className="shrink-0"
          >
            Current
          </Text>
        ) : null}
      </span>
      <GroupApprovedVersion
        version={approvedVersion}
        isLoading={approvedVersionsLoading}
      />
      <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
        <Status status={pending ? 'pending' : group.status} />
      </span>
      <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
        {installs.length ? (
          statusSlices(installs).map((slice) => (
            <Status key={slice.status} status={slice.status}>
              <span className="font-normal">
                {slice.count} {installCountLabel[slice.status] ?? slice.label}
              </span>
            </Status>
          ))
        ) : (
          <Text variant="subtext" theme="neutral">
            None
          </Text>
        )}
      </span>
      <GroupDetails
        group={group}
        configChanges={configChanges}
        configChangesKey={configChangesKey}
        onSelectInstall={onSelectInstall}
        onApproveAll={onApproveAll}
      />
    </li>
  )
}

export const BranchOverviewV2 = ({
  hasPlan,
  showInstalls,
  previewMode,
  isLoading,
  rollout,
  changes,
  configChanges,
  configChangesKey,
  groups,
  loadingStages,
  compositeError,
  failedBuilds,
  approvals,
  runHeaderAction,
  versionLabel,
  approvedVersions,
  approvedVersionsLoading,
  onSelectInstall,
  onApproveAll,
}: IBranchOverview) => {
  const currentIndex = groups.findIndex(
    (group) => stepStatusCategory(group.status) !== 'success'
  )
  const current = currentIndex >= 0 ? groups[currentIndex] : undefined
  const doneGroups = current ? currentIndex : groups.length
  const completed = groups.reduce(
    (sum, group) => sum + completedInstalls(group),
    0
  )
  const total = groups.reduce((sum, group) => sum + installTotal(group), 0)
  const showGroups =
    (showInstalls ?? hasPlan) && !isLoading && groups.length > 0

  return (
    <div className="flex flex-col gap-8 p-4 md:p-6">
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-3">
          {rollout ? (
            <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
              {versionLabel ? (
                <Text
                  variant="subtext"
                  theme="neutral"
                  family="mono"
                  className="shrink-0"
                >
                  {versionLabel}
                </Text>
              ) : null}
              {loadingStages?.length ? (
                <OverviewLoadingTrack stages={loadingStages} />
              ) : null}
              {rollout.source ? (
                <RunSourceMark source={rollout.source} />
              ) : null}
              {runHeaderAction}
            </div>
          ) : loadingStages?.length ? (
            <div className="flex items-center justify-between gap-4">
              <OverviewLoadingTrack stages={loadingStages} />
              {runHeaderAction}
            </div>
          ) : isLoading ? (
            <Loading />
          ) : null}

          {compositeError ? <CompositeError error={compositeError} /> : null}
          {failedBuilds?.length ? (
            <div className="flex flex-col gap-2">
              <Text variant="subtext" weight="strong">
                Failed builds
              </Text>
              <ul className="flex flex-col gap-1">
                {failedBuilds.map((build) => (
                  <li key={build.id}>
                    <Link href={build.href} textVariant="subtext">
                      <Text as="span" variant="subtext" family="mono">
                        {build.name}
                      </Text>
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ) : null}
          {previewMode === 'build-only' ? (
            <Banner theme="info">
              This preview will not update any install.
            </Banner>
          ) : null}
          {previewMode === 'plan-only' ? (
            <Banner theme="info">
              This preview will plan the selected install. It will not roll out
              to the branch.
            </Banner>
          ) : null}

          {rollout ? (
            <Card
              elevation="1"
              className="!flex-row !flex-wrap !items-center !justify-between !gap-4 !p-4"
            >
              <RunCommitSummary
                source={rollout.source}
                title={rollout.title}
                sha={rollout.sha}
                shaUrl={rollout.shaUrl}
                author={rollout.author}
                status={rollout.status}
                commit={rollout.commit}
                previewMode={rollout.previewMode}
                baseline={rollout.baseline}
              />
              {changes}
            </Card>
          ) : !isLoading ? (
            <Text variant="subtext" theme="neutral">
              {hasPlan
                ? 'No runs yet. Push a commit or start a run.'
                : 'This branch has no install groups yet. Every install updates at once.'}
            </Text>
          ) : null}
        </div>
        {showGroups && current ? (
          <ActiveGroup
            group={current}
            index={currentIndex}
            total={groups.length}
            approval={approvalFor(approvals, current.name)?.banner}
            onSelectInstall={onSelectInstall}
            onApproveAll={onApproveAll}
          />
        ) : null}
      </div>

      {showGroups ? (
        <section className="flex flex-col gap-3">
          <div className="flex flex-col gap-2 md:flex-row md:items-center md:gap-4">
            <Text variant="base" weight="strong" className="shrink-0">
              Deployment plan
            </Text>
            <div className="flex min-w-0 flex-1 flex-col gap-2 md:flex-row md:items-center md:gap-4">
              <DeploymentStageRail
                stages={groups.map((group) => ({
                  id: group.id,
                  name: group.name,
                  status: group.status,
                }))}
                currentId={current?.id}
                onSelect={scrollToGroup}
              />
              <Text variant="subtext" theme="neutral" className="shrink-0">
                {doneGroups} of {groups.length} groups · {completed} of {total}{' '}
                installs
              </Text>
            </div>
          </div>
          <ul className="flex flex-col divide-y overflow-hidden rounded-md border">
            <GroupColumnHeader />
            {groups.map((group, index) => (
              <GroupRow
                key={group.id}
                group={group}
                current={group.id === current?.id}
                pending={currentIndex >= 0 && index > currentIndex}
                approvedVersion={
                  approvedVersions?.[group.name.trim().toLowerCase()]
                }
                approvedVersionsLoading={approvedVersionsLoading}
                configChanges={configChanges}
                configChangesKey={configChangesKey}
                onSelectInstall={onSelectInstall}
                onApproveAll={onApproveAll}
              />
            ))}
          </ul>
        </section>
      ) : (showInstalls ?? hasPlan) && !isLoading ? (
        <Text variant="subtext" theme="neutral">
          No install groups in this run yet.
        </Text>
      ) : null}
    </div>
  )
}
