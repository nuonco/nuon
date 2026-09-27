import { Fragment, useState, type ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabelBadge } from '@/components/common/LabelBadge'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { HistoryPanelButton } from '@/components/layout/HistoryPanelButton'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import { BranchDetailActionsComponent } from '@/components/branches/BranchDetailActions'
import { BranchHeaderMeta } from '@/components/branches/BranchHeaderMeta'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { ComponentType } from '@/components/components/ComponentType'
import { cn } from '@/utils/classnames'
import type {
  TBranchOverview,
  TPlanGroup,
  TChange,
  TTemplateEntry,
  TTemplateItem,
} from './fixtures'
import {
  RolloutTrack,
  type TTrackGroup,
} from '@/components/branches/BranchOverview/RolloutTrack'
import { RolloutTiles } from '@/components/branches/BranchOverview/RolloutTiles'
import { RunSourceCard } from '@/components/branches/BranchOverview/RunSourceCard'

const pluralize = (count: number, noun: string) =>
  `${count} ${noun}${count === 1 ? '' : 's'}`

const GroupRules = ({ group }: { group: TPlanGroup }) => (
  <span className="flex flex-wrap items-center gap-1.5">
    {group.selector ? (
      Object.entries(group.selector).map(([key, value]) => (
        <LabelBadge key={key} labelKey={key} labelValue={value} size="sm" />
      ))
    ) : (
      <Text variant="subtext" theme="neutral">
        Every other install
      </Text>
    )}
  </span>
)

const playgroundInstallStatus = (state: string) => {
  if (state === 'error') {
    return {
      resources: { status: 'error', detail: '2 resources have issues' },
      deployment: { status: 'error', detail: 'Deploy failed' },
      health: { status: 'unhealthy', detail: '1 health check failed' },
    }
  }
  if (state === 'in-progress') {
    return {
      resources: { status: 'active', detail: 'All resources are deployed' },
      deployment: { status: 'in-progress', detail: 'Updating components' },
      health: { status: 'pending', detail: 'Health checks pending' },
    }
  }
  if (state === 'success') {
    return {
      resources: { status: 'active', detail: 'All resources are deployed' },
      deployment: { status: 'success', detail: 'Components are current' },
      health: { status: 'healthy', detail: 'All health checks pass' },
    }
  }
  return {
    resources: { status: 'pending', detail: 'Not started' },
    deployment: { status: 'pending', detail: 'Waiting for this group' },
    health: { status: 'unknown' },
  }
}

const groupRules = (group: TPlanGroup) =>
  [
    group.selector
      ? Object.entries(group.selector)
          .map(([key, value]) => `${key}=${value}`)
          .join(', ')
      : 'Every other install',
    `up to ${group.maxParallel} at a time`,
    group.approval === 'manual' ? 'manual approval' : 'auto-approve',
  ].join(' · ')

const trackGroups = (branch: TBranchOverview): TTrackGroup[] => {
  const stages = branch.rollout?.stages.filter(
    (stage) => stage.kind === 'group'
  )
  return branch.groups.map((group) => {
    const stage = stages?.find((item) => item.groupId === group.id)
    return {
      id: group.id,
      name: group.name,
      status: stage?.state ?? 'pending',
      plannedCount: group.installCount,
      rules: groupRules(group),
      installs: (stage?.installs ?? []).map((install) => ({
        id: install.id,
        name: install.name,
        status: install.state,
        detail: install.region,
        durationNs: install.durationNs,
        ...playgroundInstallStatus(install.state),
        overviewHref: `#installs/${install.id}`,
        workflowHref:
          install.state === 'pending' ? undefined : `#workflows/${install.id}`,
      })),
    }
  })
}

const RecentRuns = ({ branch }: { branch: TBranchOverview }) => (
  <section className="flex flex-col gap-3">
    <SectionHeader
      title="Recent runs"
      actions={<Link href="#activity">View activity</Link>}
    />
    {branch.recentRuns.length === 0 ? (
      <Text variant="subtext" theme="neutral">
        Earlier runs will appear here.
      </Text>
    ) : (
      <ul className="flex flex-col divide-y border-y">
        {branch.recentRuns.map((run) => (
          <li
            key={run.id}
            className="grid grid-cols-[auto_minmax(0,1fr)_5rem_13rem_7rem] items-center gap-4 py-2.5"
          >
            <Status status={run.state} isWithoutText />
            <Link href={`#${run.id}`} className="truncate">
              {run.title}
            </Link>
            <Text variant="subtext" family="mono" theme="neutral">
              {run.sha.slice(0, 7)}
            </Text>
            <Text
              variant="subtext"
              theme={run.state === 'error' ? 'error' : 'neutral'}
              className="truncate"
            >
              {run.outcome}
            </Text>
            <Time
              time={run.createdAt}
              format="relative"
              variant="subtext"
              theme="neutral"
              className="text-right"
            />
          </li>
        ))}
      </ul>
    )}
  </section>
)

const CHANGE_MARK: Record<
  TChange['op'],
  { mark: string; theme: 'success' | 'warn' | 'error' }
> = {
  add: { mark: '+', theme: 'success' },
  change: { mark: '~', theme: 'warn' },
  remove: { mark: '-', theme: 'error' },
}

const Changes = ({ changes }: { changes: TChange[] }) => {
  const added = changes.filter((change) => change.op === 'add').length
  const updated = changes.filter((change) => change.op === 'change').length
  const removed = changes.filter((change) => change.op === 'remove').length

  return (
    <div className="flex min-w-0 flex-col gap-3">
      <SectionHeader
        title="What's changed"
        actions={
          <span className="flex items-center gap-4">
            <Link href="#builds">View builds</Link>
            <ChangeCountSummary
              added={added}
              updated={updated}
              removed={removed}
            />
          </span>
        }
      />
      {changes.length === 0 ? (
        <Text variant="subtext" theme="neutral">
          No config changes in this rollout.
        </Text>
      ) : (
        <ul className="flex flex-col divide-y border-y">
          {changes.map((change) => (
            <li
              key={`${change.op}-${change.name}`}
              className="grid grid-cols-[1.5rem_minmax(0,10rem)_minmax(0,1fr)] items-center gap-3 py-2.5"
            >
              <Text
                variant="subtext"
                family="mono"
                weight="strong"
                theme={CHANGE_MARK[change.op].theme}
              >
                {CHANGE_MARK[change.op].mark}
              </Text>
              <Text variant="subtext" family="mono" className="truncate">
                {change.name}
              </Text>
              <Text variant="subtext" theme="neutral" className="truncate">
                {change.detail}
              </Text>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

const Overview = ({
  branch,
  onOpenRollout,
}: {
  branch: TBranchOverview
  onOpenRollout: (groupId?: string) => void
}) => {
  const rollout = branch.rollout
  const hasPlan = branch.groups.length > 0

  return (
    <div className="flex flex-col gap-10 p-4 md:p-6">
      {rollout ? (
        <section className="grid items-start gap-6 lg:grid-cols-[20rem_minmax(0,1fr)]">
          <RunSourceCard
            source={rollout.source}
            title={rollout.title}
            sha={rollout.sha}
            shaUrl={`https://github.com/${branch.repo}/commit/${rollout.sha}`}
            author={rollout.author}
            status={rollout.state}
          />
          <Changes changes={rollout.changes} />
        </section>
      ) : (
        <Text variant="subtext" theme="neutral">
          {hasPlan
            ? 'No runs yet. Push a commit or start a run.'
            : 'This branch has no install groups yet. Every install updates at once.'}
        </Text>
      )}

      {hasPlan ? (
        <section className="flex flex-col gap-3">
          <SectionHeader
            title="Installs"
            description={rollout?.activity}
            actions={
              <Button variant="ghost" size="sm" onClick={() => onOpenRollout()}>
                View rollout
              </Button>
            }
          />
          <RolloutTiles
            groups={trackGroups(branch)}
            onSelectGroup={onOpenRollout}
          />
        </section>
      ) : null}
    </div>
  )
}

const Rollout = ({
  branch,
  groupId,
}: {
  branch: TBranchOverview
  groupId?: string
}) => {
  const rollout = branch.rollout
  return (
    <div className="flex flex-col gap-3 p-4 md:p-6">
      <SectionHeader
        title={rollout ? `Rollout #${rollout.number}` : 'Rollout'}
        description={rollout?.activity}
        actions={rollout ? <Link href={`#${rollout.id}`}>View run</Link> : null}
      />
      {branch.groups.length === 0 ? (
        <Text variant="subtext" theme="neutral">
          This branch has no install groups yet. Every install updates at once.
        </Text>
      ) : (
        <RolloutTrack
          key={`${rollout?.id}-${groupId}`}
          groups={trackGroups(branch)}
          initialGroupId={groupId}
        />
      )}
    </div>
  )
}

const Settings = ({ branch }: { branch: TBranchOverview }) => (
  <div className="flex flex-col gap-10 p-4 md:p-6">
    <section className="flex flex-col gap-3">
      <SectionHeader
        title="Deployment plan"
        description="Installs roll out in this order. Each group waits for the one before it."
        actions={
          <Button variant="primary">
            {branch.groups.length ? 'Edit plan' : 'Create deployment plan'}
          </Button>
        }
      />
      {branch.groups.length === 0 ? (
        <Text variant="subtext" theme="neutral">
          No install groups. Every install on this branch updates at once.
        </Text>
      ) : (
        <ol className="flex flex-col divide-y border-y">
          {branch.groups.map((group, idx) => (
            <li
              key={group.id}
              className="grid grid-cols-[1.5rem_10rem_minmax(0,1fr)_6rem_8rem_8rem] items-center gap-4 py-3"
            >
              <Text variant="subtext" family="mono" theme="neutral">
                {idx + 1}
              </Text>
              <Text variant="body" weight="strong" className="truncate">
                {group.name}
              </Text>
              <GroupRules group={group} />
              <Text variant="subtext" theme="neutral" nowrap>
                {pluralize(group.installCount, 'install')}
              </Text>
              <Text variant="subtext" theme="neutral" nowrap>
                Up to {group.maxParallel} at a time
              </Text>
              <Badge
                size="sm"
                theme={group.approval === 'manual' ? 'warn' : 'neutral'}
                className="justify-self-end"
              >
                {group.approval === 'manual'
                  ? 'Manual approval'
                  : 'Auto-approve'}
              </Badge>
            </li>
          ))}
        </ol>
      )}
    </section>

    <section className="flex flex-col gap-3">
      <SectionHeader title="Source" />
      <dl className="grid grid-cols-[8rem_1fr] gap-x-4 gap-y-2">
        <Text as="dt" variant="subtext" theme="neutral">
          Repository
        </Text>
        <dd>
          <Link href={`https://github.com/${branch.repo}`} isExternal>
            {branch.repo}
          </Link>
        </dd>
        <Text as="dt" variant="subtext" theme="neutral">
          Directory
        </Text>
        <Text as="dd" variant="subtext" family="mono">
          {branch.directory}
        </Text>
        <Text as="dt" variant="subtext" theme="neutral">
          Runs on
        </Text>
        <Text as="dd" variant="subtext">
          {branch.trigger}
        </Text>
      </dl>
    </section>
  </div>
)

const COLUMN_CLASS = [
  '',
  'grid-cols-1',
  'grid-cols-[minmax(0,1.5fr)_minmax(0,1fr)]',
  'grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)_minmax(0,1fr)]',
]

const DRILL_COLUMN_CLASS = [
  '',
  'grid-cols-[minmax(0,1fr)_auto]',
  'grid-cols-[minmax(0,1.5fr)_minmax(0,1fr)_auto]',
  'grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)_minmax(0,1fr)_auto]',
]

const DRILL_IN = new Set(['components', 'actions', 'runbooks', 'policies'])

const TemplateBackLink = ({
  label,
  onClick,
}: {
  label: string
  onClick: () => void
}) => (
  <button
    type="button"
    onClick={onClick}
    className="flex w-fit items-center gap-1.5 text-link hover:text-link-hover focus-visible:rounded"
  >
    <Icon variant="CaretLeftIcon" weight="bold" />
    <Text variant="base" weight="strong" className="text-inherit">
      {label}
    </Text>
  </button>
)

const RowLink = ({
  children,
  onClick,
}: {
  children: ReactNode
  onClick: () => void
}) => (
  <button
    type="button"
    onClick={onClick}
    className="w-fit truncate text-left text-link hover:text-link-hover hover:underline focus-visible:rounded"
  >
    {children}
  </button>
)

const COMPONENT_GRID =
  'grid grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)_minmax(0,1.2fr)_minmax(0,1fr)_auto] items-center gap-4'

const ComponentList = ({
  item,
  onSelect,
}: {
  item: TTemplateItem
  onSelect: (name: string) => void
}) => (
  <div className="flex flex-col gap-3 p-4 md:p-6">
    <SectionHeader title={item.label} description={item.description} />
    <div className="overflow-x-auto">
      <div className={cn(COMPONENT_GRID, 'border-b py-2')}>
        {['Component name', 'Type', 'Dependencies', 'Latest build', ''].map(
          (column) => (
            <Text key={column} variant="label" theme="neutral">
              {column}
            </Text>
          )
        )}
      </div>
      {item.entries.map((row) => {
        const component = row.detail?.component
        const build = row.detail?.builds?.at(0)
        const deps = row.detail?.dependencies ?? []
        return (
          <div
            key={row.name}
            className={cn(
              COMPONENT_GRID,
              'border-b py-3 hover:bg-black/[0.02] dark:hover:bg-white/[0.03]'
            )}
          >
            <span className="flex min-w-0 flex-col">
              <RowLink onClick={() => onSelect(row.name)}>
                <Text variant="body" className="text-inherit">
                  {row.name}
                </Text>
              </RowLink>
              {component ? <ID>{component.id}</ID> : null}
            </span>
            {component ? (
              <ComponentType
                type={component.type}
                variant="subtext"
                colorVariant="color"
              />
            ) : (
              <Text variant="subtext">{row.cells[0]}</Text>
            )}
            {deps.length ? (
              <span className="flex flex-wrap gap-1">
                {deps.map((dep) => (
                  <Badge key={dep} size="sm" variant="code" theme="neutral">
                    {dep}
                  </Badge>
                ))}
              </span>
            ) : (
              <Icon variant="MinusIcon" />
            )}
            {build ? (
              <Status variant="badge" status={build.status} />
            ) : (
              <Icon variant="MinusIcon" />
            )}
            <RowLink onClick={() => onSelect(row.name)}>
              <Text variant="subtext" className="text-inherit">
                View component
              </Text>
            </RowLink>
          </div>
        )
      })}
    </div>
  </div>
)

const ComponentPage = ({
  entry,
  onBack,
}: {
  entry: TTemplateEntry
  onBack: () => void
}) => {
  const detail = entry.detail
  const component = detail?.component
  if (!detail || !component) return null
  const build = detail.builds?.at(0)
  const dependents = component.dependents ?? []
  const dependencies = detail.dependencies ?? []

  return (
    <div className="flex flex-col gap-6 p-4 md:p-6">
      <TemplateBackLink label="Components" onClick={onBack} />
      <DetailHeader
        backLink={false}
        icon={
          <ComponentType
            type={component.type}
            displayVariant="icon-only"
            colorVariant="color"
            iconSize="24"
          />
        }
        title={entry.name}
        status={build ? <Status status={build.status} /> : null}
        id={component.id}
        actions={
          <>
            <HistoryPanelButton
              title="Previous builds"
              history={
                <ul className="flex flex-col divide-y">
                  {(detail.builds ?? []).map((b) => (
                    <li
                      key={b.sha}
                      className="flex items-center justify-between gap-4 py-3"
                    >
                      <Text variant="subtext" family="mono">
                        {b.sha}
                      </Text>
                      <Status status={b.status} />
                      <Text variant="subtext" theme="neutral">
                        {b.when}
                      </Text>
                    </li>
                  ))}
                </ul>
              }
            />
            <Button variant="primary">Build component</Button>
          </>
        }
      />

      <Card className="gap-4">
        <div className="flex items-start justify-between gap-4">
          <Text weight="strong">Build source</Text>
          <Link href="#runs">View app branch run</Link>
        </div>
        <BranchRunCommit
          status={build?.status}
          href="#builds"
          message={component.source.message}
          author={component.source.author}
          sha={component.source.sha}
          createdAt={component.source.createdAt}
        />
      </Card>

      <Card>
        <div className="flex flex-col gap-6">
          <Text weight="strong">Configuration</Text>
          <div className="flex items-start gap-6">
            <LabeledValue label="Version">{component.version}</LabeledValue>
            <LabeledValue label="Type">
              <ComponentType type={component.type} variant="subtext" />
            </LabeledValue>
          </div>
          <div className="grid grid-cols-2 gap-6 md:grid-cols-3 lg:grid-cols-4">
            <LabeledValue label="Build timeout">
              {component.buildTimeout}
            </LabeledValue>
            <LabeledValue label="Deploy timeout">
              {component.deployTimeout}
            </LabeledValue>
            {component.config.map((field) => (
              <LabeledValue key={field.label} label={field.label}>
                {field.value}
              </LabeledValue>
            ))}
          </div>
          {dependencies.length || dependents.length ? (
            <div className="flex flex-col gap-6 border-t pt-6">
              {dependencies.length ? (
                <DependencyList title="Dependencies" names={dependencies} />
              ) : null}
              {dependents.length ? (
                <DependencyList title="Dependents" names={dependents} />
              ) : null}
            </div>
          ) : null}
        </div>
      </Card>
    </div>
  )
}

const DependencyList = ({
  title,
  names,
}: {
  title: string
  names: string[]
}) => (
  <div className="flex flex-col gap-2">
    <Text variant="body" weight="strong" level={5}>
      {title}
    </Text>
    <span className="flex flex-wrap gap-1">
      {names.map((name) => (
        <Badge key={name} size="sm" variant="code" theme="neutral">
          {name}
        </Badge>
      ))}
    </span>
  </div>
)

const TemplateDetail = ({
  item,
  entry,
  onBack,
}: {
  item: TTemplateItem
  entry: TTemplateEntry
  onBack: () => void
}) => {
  const detail = entry.detail
  return (
    <div className="flex flex-col gap-8 p-4 md:p-6">
      <TemplateBackLink label={item.label} onClick={onBack} />
      <SectionHeader title={entry.name} description={detail?.summary} />
      {detail?.fields?.length ? (
        <dl className="grid grid-cols-[8rem_1fr] gap-x-4 gap-y-2">
          {detail.fields.map((field) => (
            <Fragment key={field.label}>
              <Text as="dt" variant="subtext" theme="neutral">
                {field.label}
              </Text>
              <Text as="dd" variant="subtext" family="mono">
                {field.value}
              </Text>
            </Fragment>
          ))}
        </dl>
      ) : null}
      {detail?.dependencies?.length ? (
        <section className="flex flex-col gap-2">
          <Text variant="body" weight="strong">
            Dependencies
          </Text>
          <span className="flex flex-wrap gap-2">
            {detail.dependencies.map((name) => (
              <Text key={name} variant="subtext" family="mono">
                {name}
              </Text>
            ))}
          </span>
        </section>
      ) : null}
      {detail?.steps?.length ? (
        <section className="flex flex-col gap-3">
          <Text variant="body" weight="strong">
            Steps
          </Text>
          <ol className="flex flex-col divide-y border-y">
            {detail.steps.map((step, index) => (
              <li
                key={step.name}
                className="grid grid-cols-[1.5rem_minmax(0,10rem)_minmax(0,1fr)] items-center gap-3 py-2.5"
              >
                <Text variant="subtext" family="mono" theme="neutral">
                  {index + 1}
                </Text>
                <Text variant="subtext" family="mono">
                  {step.name}
                </Text>
                <Text variant="subtext" theme="neutral">
                  {step.detail}
                </Text>
              </li>
            ))}
          </ol>
        </section>
      ) : null}
      {detail?.builds?.length ? (
        <section className="flex flex-col gap-3">
          <Text variant="body" weight="strong">
            Builds
          </Text>
          <ul className="flex flex-col divide-y border-y">
            {detail.builds.map((build) => (
              <li
                key={build.sha}
                className="grid grid-cols-[6rem_8rem_minmax(0,1fr)] items-center gap-4 py-2.5"
              >
                <Text variant="subtext" family="mono">
                  {build.sha}
                </Text>
                <Status status={build.status} />
                <Text variant="subtext" theme="neutral">
                  {build.when}
                </Text>
              </li>
            ))}
          </ul>
        </section>
      ) : null}
    </div>
  )
}

const TemplatePart = ({
  item,
  entry,
  onSelect,
  onBack,
}: {
  item: TTemplateItem
  entry?: TTemplateEntry
  onSelect: (name: string) => void
  onBack: () => void
}) => {
  if (item.id === 'components') {
    return entry?.detail?.component ? (
      <ComponentPage entry={entry} onBack={onBack} />
    ) : (
      <ComponentList item={item} onSelect={onSelect} />
    )
  }

  if (entry?.detail) {
    return <TemplateDetail item={item} entry={entry} onBack={onBack} />
  }

  const canDrill = DRILL_IN.has(item.id)
  const gridClass = (canDrill ? DRILL_COLUMN_CLASS : COLUMN_CLASS)[
    item.columns.length
  ]
  return (
    <div className="flex flex-col gap-3 p-4 md:p-6">
      <SectionHeader title={item.label} description={item.description} />
      <div className="overflow-x-auto">
        <div className={cn('grid items-center gap-4 border-b py-2', gridClass)}>
          {item.columns.map((column) => (
            <Text key={column} variant="label" theme="neutral">
              {column}
            </Text>
          ))}
          {canDrill ? <span /> : null}
        </div>
        {item.entries.map((row) => (
          <div
            key={row.name}
            className={cn(
              'grid items-center gap-4 border-b py-2.5',
              gridClass,
              canDrill && 'hover:bg-black/[0.02] dark:hover:bg-white/[0.03]'
            )}
          >
            {canDrill ? (
              <RowLink onClick={() => onSelect(row.name)}>
                <Text variant="subtext" family="mono" className="text-inherit">
                  {row.name}
                </Text>
              </RowLink>
            ) : (
              <Text variant="subtext" family="mono" className="truncate">
                {row.name}
              </Text>
            )}
            {row.cells.map((cell, index) => (
              <Text
                key={`${row.name}-${index}`}
                variant="subtext"
                theme="neutral"
                className="truncate"
              >
                {cell}
              </Text>
            ))}
            {canDrill ? (
              <RowLink onClick={() => onSelect(row.name)}>
                <Text variant="subtext" className="text-inherit">
                  View
                </Text>
              </RowLink>
            ) : null}
          </div>
        ))}
      </div>
    </div>
  )
}

const TEMPLATE_ICONS = {
  components: 'CardsIcon',
  inputs: 'ListChecksIcon',
  actions: 'TerminalWindowIcon',
  runbooks: 'BookIcon',
  sandboxes: 'ShippingContainerIcon',
  policies: 'ShieldCheckIcon',
  roles: 'FileLockIcon',
  labels: 'TagIcon',
  readme: 'BookOpenIcon',
} as const

const NavButton = ({
  icon,
  label,
  isActive,
  count,
  onClick,
}: {
  icon:
    | (typeof TEMPLATE_ICONS)[keyof typeof TEMPLATE_ICONS]
    | 'GraphIcon'
    | 'ListIcon'
    | 'StackIcon'
    | 'GearIcon'
  label: string
  isActive: boolean
  count?: number
  onClick: () => void
}) => (
  <button
    type="button"
    onClick={onClick}
    aria-current={isActive ? 'page' : undefined}
    className={cn(
      'flex w-full items-center gap-3 rounded-md px-3 py-2 text-left',
      isActive
        ? 'bg-primary-200 text-primary-800 dark:bg-primary-600/25 dark:text-primary-400'
        : 'text-cool-grey-800 hover:bg-black/5 dark:text-cool-grey-400 dark:hover:bg-white/10'
    )}
  >
    <Icon variant={icon} size={16} />
    <Text variant="subtext" className="flex-1 truncate">
      {label}
    </Text>
    {count != null ? (
      <Text variant="label" theme="neutral">
        {count}
      </Text>
    ) : null}
  </button>
)

type TView = 'overview' | 'runs' | 'rollout' | 'settings' | 'template'

export interface IBranchOverviewPlayground {
  branch: TBranchOverview
  initialView?: TView
}

export const BranchOverviewPlayground = ({
  branch,
  initialView = 'overview',
}: IBranchOverviewPlayground) => {
  const [view, setView] = useState<TView>(initialView)
  const [templateId, setTemplateId] = useState(branch.template[0]?.id)
  const [templateEntryName, setTemplateEntryName] = useState<string>()
  const [rolloutGroupId, setRolloutGroupId] = useState<string>()
  const openRollout = (groupId?: string) => {
    setRolloutGroupId(groupId)
    setView('rollout')
  }
  const template =
    branch.template.find((item) => item.id === templateId) ?? branch.template[0]
  const templateEntry = template?.entries.find(
    (entry) => entry.name === templateEntryName
  )
  const isOverview = view === 'overview'

  return (
    <div className="flex h-full min-h-[44rem]">
      <aside className="flex w-60 shrink-0 flex-col border-r">
        <nav aria-label="Branch" className="flex flex-col gap-1 p-3">
          <NavButton
            icon="GraphIcon"
            label="Overview"
            isActive={view === 'overview'}
            onClick={() => setView('overview')}
          />
          <NavButton
            icon="ListIcon"
            label="Runs"
            isActive={view === 'runs'}
            onClick={() => setView('runs')}
          />
          <NavButton
            icon="StackIcon"
            label="Rollout"
            isActive={view === 'rollout'}
            onClick={() => openRollout()}
          />
          <NavButton
            icon="GearIcon"
            label="Settings"
            isActive={view === 'settings'}
            onClick={() => setView('settings')}
          />
        </nav>
        <nav
          aria-label="App template"
          className="mt-auto flex flex-col gap-1 border-t p-3"
        >
          <Text
            variant="label"
            theme="neutral"
            className="px-3 py-1 uppercase tracking-wider"
          >
            App template
          </Text>
          {branch.template.map((item) => (
            <NavButton
              key={item.id}
              icon={
                TEMPLATE_ICONS[item.id as keyof typeof TEMPLATE_ICONS] ??
                'CardsIcon'
              }
              label={item.label}
              count={item.entries.length}
              isActive={view === 'template' && template?.id === item.id}
              onClick={() => {
                setTemplateId(item.id)
                setTemplateEntryName(undefined)
                setView('template')
              }}
            />
          ))}
        </nav>
      </aside>
      <div className="flex min-w-0 flex-1 flex-col overflow-y-auto">
        <DetailHeader
          variant="page"
          backLink={false}
          title={branch.appName}
          identity={
            <BranchHeaderMeta
              configuration={
                <Text
                  as="span"
                  variant="subtext"
                  weight="strong"
                  flex
                  nowrap
                  className="gap-1"
                >
                  {branch.branchName}
                  <Icon variant="CaretUpDownIcon" size={12} />
                </Text>
              }
              repo={branch.repo}
              gitBranch={branch.gitBranch}
              directory={branch.directory}
              trigger={branch.trigger}
            />
          }
          actions={
            branch.groups.length === 0 ? (
              <Button variant="primary" onClick={() => setView('settings')}>
                Create deployment plan
              </Button>
            ) : (
              <BranchDetailActionsComponent
                isTriggerPending={false}
                onTriggerRun={() => {}}
                onTriggerPreviewModal={() => {}}
              />
            )
          }
        />
        <div className="border-t">
          {view === 'settings' ? <Settings branch={branch} /> : null}
          {view === 'runs' ? (
            <div className="p-4 md:p-6">
              <RecentRuns branch={branch} />
            </div>
          ) : null}
          {view === 'template' && template ? (
            <TemplatePart
              item={template}
              entry={templateEntry}
              onSelect={setTemplateEntryName}
              onBack={() => setTemplateEntryName(undefined)}
            />
          ) : null}
          {view === 'rollout' ? (
            <Rollout branch={branch} groupId={rolloutGroupId} />
          ) : null}
          {isOverview ? (
            <Overview branch={branch} onOpenRollout={openRollout} />
          ) : null}
        </div>
      </div>
    </div>
  )
}
