import { useEffect, useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Avatar } from '@/components/common/Avatar'
import { Duration } from '@/components/common/Duration'
import {
  CommitLink,
  PullRequestLink,
} from '@/components/common/GitReferenceLink'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { Time } from '@/components/common/Time'
import { SectionHeader } from '@/components/layout/SectionHeader'
import {
  PreviewPanel,
  type TPreviewRun,
} from '@/components/installs/DeploymentDetail/DeploymentDetail.preview'
import { Link } from '@/components/common/Link'
import { Panel } from '@/components/surfaces/Panel'
import { useInstallHref } from '@/hooks/use-install-path'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TWorkflowStep } from '@/types'
import { cn } from '@/utils/classnames'
import { humanize } from '@/utils/string-utils'
import type {
  TConfigChange,
  TDeploymentGroup,
  TDeploymentOverview,
  TGroupInstall,
  TGroupInstallStatus,
} from './fixtures'

const STATUS_SUMMARY: { status: TGroupInstallStatus; label: string }[] = [
  { status: 'success', label: 'completed' },
  { status: 'failed', label: 'failed' },
  { status: 'in-progress', label: 'in progress' },
  { status: 'awaiting-approval', label: 'awaiting approval' },
  { status: 'pending', label: 'pending' },
]

const countStatus = (installs: TGroupInstall[], status: TGroupInstallStatus) =>
  installs.filter((install) => install.status === status).length

const groupColumnClass =
  'items-center gap-3 px-4 md:grid-cols-[minmax(8rem,1fr)_minmax(8rem,1fr)_minmax(12rem,1.4fr)_7rem]'

const GroupColumnHeader = () => (
  <li
    className={cn('hidden py-2 md:grid', groupColumnClass)}
    aria-hidden="true"
  >
    <Text variant="subtext" theme="neutral">
      Group
    </Text>
    <Text variant="subtext" theme="neutral">
      Last version
    </Text>
    <Text variant="subtext" theme="neutral">
      Status
    </Text>
    <span />
  </li>
)

const ChangeList = ({
  changes,
  empty = 'No config changes.',
}: {
  changes: TConfigChange[]
  empty?: string
}) => {
  if (!changes.length) {
    return (
      <Text variant="subtext" theme="neutral">
        {empty}
      </Text>
    )
  }

  return (
    <ul className="flex flex-col gap-3">
      {changes.map((change) => (
        <li
          key={change.id}
          className="flex flex-wrap items-center gap-x-3 gap-y-1"
        >
          <Badge size="sm">{change.kind}</Badge>
          <Text family="mono" variant="subtext">
            {change.field ? `${change.name}.${change.field}` : change.name}
          </Text>
          <Text family="mono" variant="subtext" theme="neutral">
            {change.before}
          </Text>
          <Text variant="subtext" theme="neutral">
            →
          </Text>
          <Text family="mono" variant="subtext">
            {change.after}
          </Text>
        </li>
      ))}
    </ul>
  )
}

const installActivity = (install: TGroupInstall) => {
  if (install.status === 'failed') return install.error || 'Install failed'
  if (install.status === 'success') return 'Install deployed'
  if (install.status === 'awaiting-approval') return 'Waiting for approval'
  if (install.status === 'in-progress') return 'Applying install'
  return 'Waiting to start'
}

const installPreviewRun = (
  group: TDeploymentGroup,
  install: TGroupInstall
): TPreviewRun => {
  const status =
    install.status === 'failed'
      ? 'error'
      : install.status === 'success'
        ? 'success'
        : 'in-progress'
  const states =
    install.status === 'pending'
      ? ['pending', 'pending', 'pending', 'pending']
      : install.status === 'awaiting-approval'
        ? ['success', 'success', 'approval-awaiting', 'pending']
        : install.status === 'in-progress'
          ? ['success', 'success', 'in-progress', 'pending']
          : install.status === 'failed'
            ? ['success', 'success', 'error', 'not-started']
            : ['success', 'success', 'success', 'success']
  const names = [
    'Sync configuration',
    'Plan install',
    'Deploy install',
    'Verify install health',
  ]
  const outcomeStatus =
    install.status === 'success'
      ? 'success'
      : install.status === 'failed'
        ? 'error'
        : install.status === 'in-progress'
          ? 'in-progress'
          : 'not-started'

  return {
    id: `wf-${install.id}`,
    title: install.name,
    type: 'component_deploy',
    resources: [install.name],
    status,
    scope: 'components',
    activity: installActivity(install),
    changesApplied: install.status === 'failed',
    created_at: group.startedAt ?? '2026-10-15T13:40:00.000Z',
    outcomes: [
      {
        category: 'Components',
        name: install.name,
        status: outcomeStatus,
        detail: installActivity(install),
      },
    ],
    steps: names.map(
      (name, index) =>
        ({
          id: `${install.id}-step-${index}`,
          name,
          execution_type: 'system',
          group_idx: index,
          finished: ['success', 'error'].includes(states[index]),
          execution_time: ['success', 'error'].includes(states[index])
            ? (index + 1) * 2_000_000_000
            : undefined,
          status: { status: states[index] },
        }) as TWorkflowStep
    ),
  }
}

const useInstallWorkflowOpener = () => {
  const { addPanel, panels, removePanel } = useSurfaces()

  return (group: TDeploymentGroup, install: TGroupInstall) => {
    const panelKey = `install-deployment-${install.id}`
    if (panels.some((panel) => panel.key === panelKey && panel.isVisible)) {
      return
    }
    panels
      .filter(
        (panel) =>
          panel.key?.startsWith('install-deployment-') && panel.isVisible
      )
      .forEach((panel) => removePanel(panel.id, panel.key))
    addPanel(
      <PreviewPanel panelKey={panelKey} run={installPreviewRun(group, install)} />,
      panelKey
    )
  }
}

const InstallList = ({
  group,
  installs,
  onApprove,
}: {
  group: TDeploymentGroup
  installs: TGroupInstall[]
  onApprove: (id: string) => void
}) => {
  const openWorkflow = useInstallWorkflowOpener()
  const installHref = useInstallHref()

  if (!installs.length) {
    return (
      <Text variant="subtext" theme="neutral">
        This group has not started deploying.
      </Text>
    )
  }

  return (
    <ul className="flex flex-col divide-y overflow-hidden rounded-md border">
      {installs.map((install) => (
          <li key={install.id} className="flex flex-col gap-2 px-4 py-3">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <Link
                href={installHref({ orgId: 'org-acme', installId: install.id })}
                className="min-w-0 truncate font-mono"
              >
                {install.name}
              </Link>
              <span className="flex items-center gap-3">
                <Status status={install.status} />
                {install.status === 'awaiting-approval' ? (
                  <Button
                    size="sm"
                    variant="primary"
                    onClick={() => onApprove(install.id)}
                  >
                    Approve
                  </Button>
                ) : null}
                <Button
                  size="sm"
                  variant="secondary"
                  onClick={() => openWorkflow(group, install)}
                >
                  View details
                </Button>
              </span>
            </div>
            {install.error ? (
              <Text variant="subtext" theme="error">
                {install.error}
              </Text>
            ) : null}
          </li>
      ))}
    </ul>
  )
}

const GroupDrawer = ({
  group,
  initialTab = 'installs',
  onInstallsChange,
}: {
  group: TDeploymentGroup
  initialTab?: 'changes' | 'installs'
  onInstallsChange: (installs: TGroupInstall[]) => void
}) => {
  const [installs, setInstalls] = useState(group.installs)
  const awaiting = countStatus(installs, 'awaiting-approval')

  const approve = (id?: string) => {
    const next = installs.map((install) =>
      install.status === 'awaiting-approval' &&
      (id == null || install.id === id)
        ? { ...install, status: 'in-progress' as const }
        : install
    )
    setInstalls(next)
    onInstallsChange(next)
  }

  return (
    <Tabs
      naturalHeight
      initActiveTab={initialTab}
      actions={
        <Button
          size="sm"
          variant="primary"
          disabled={awaiting === 0}
          tooltipProps={
            awaiting === 0
              ? {
                  position: 'bottom',
                  tipContent: (
                    <Text variant="subtext">
                      Cannot approve — no installs are awaiting approval
                    </Text>
                  ),
                }
              : undefined
          }
          onClick={() => approve()}
        >
          Approve all
        </Button>
      }
      tabs={{
        changes: (
          <ChangeList
            changes={group.changes}
            empty="No config changes for this group."
          />
        ),
          installs: (
            <InstallList
              group={group}
              installs={installs}
              onApprove={approve}
            />
          ),
      }}
    />
  )
}

const groupAnchorId = (groupId: string) => `deployment-group-${groupId}`

const scrollToGroup = (groupId: string) => {
  const reduceMotion = window.matchMedia(
    '(prefers-reduced-motion: reduce)'
  ).matches
  document.getElementById(groupAnchorId(groupId))?.scrollIntoView({
    behavior: reduceMotion ? 'auto' : 'smooth',
    block: 'center',
  })
}

const isGroupComplete = (group: TDeploymentGroup) =>
  group.plannedCount > 0 &&
  countStatus(group.installs, 'success') === group.plannedCount

const GroupRow = ({
  group,
  initialTab,
  autoOpen,
  pending,
  current,
  onInstallsChange,
}: {
  group: TDeploymentGroup
  initialTab?: 'changes' | 'installs'
  autoOpen?: boolean
  pending?: boolean
  current?: boolean
  onInstallsChange: (installs: TGroupInstall[]) => void
}) => {
  const { addPanel, panels, removePanel } = useSurfaces()
  const panelKey = `deployment-group-${group.id}`
  const canOpen = group.installs.length > 0 || group.changes.length > 0

  const open = () => {
    if (panels.some((panel) => panel.key === panelKey && panel.isVisible)) {
      return
    }
    panels
      .filter(
        (panel) =>
          panel.key?.startsWith('deployment-group-') &&
          panel.key !== panelKey &&
          panel.isVisible
      )
      .forEach((panel) => removePanel(panel.id, panel.key))
    addPanel(
      <Panel
        panelKey={panelKey}
        size="half"
        heading={
          <div className="flex flex-col gap-0.5">
            <Text variant="subtext" theme="neutral">
              Group information
            </Text>
            <span className="flex items-baseline gap-2">
              <Text variant="base" weight="strong" family="mono">
                {group.name}
              </Text>
              <Text variant="subtext" theme="neutral">
                {group.label}
              </Text>
            </span>
          </div>
        }
      >
        <GroupDrawer
          group={group}
          initialTab={initialTab}
          onInstallsChange={onInstallsChange}
        />
      </Panel>,
      panelKey
    )
  }

  useEffect(() => {
    if (!autoOpen || !canOpen) return
    // Stories that start with a group open the drawer once. The timeout is
    // cleared if this effect re-runs before it fires, so only one panel opens.
    const timer = window.setTimeout(open, 0)
    return () => window.clearTimeout(timer)
  }, [])

  const deploying = group.installs.some(
    (install) => install.status === 'in-progress'
  )
  const awaiting = countStatus(group.installs, 'awaiting-approval')
  const currentStatus =
    !deploying && awaiting > 0 ? 'awaiting-approval' : 'in-progress'

  return (
    <li
      className={cn(
        'grid grid-cols-1 py-3',
        groupColumnClass,
        current && 'opacity-60'
      )}
      id={groupAnchorId(group.id)}
      data-group={group.name}
      data-current={current ? 'true' : undefined}
    >
      <span className="flex min-w-0 items-baseline gap-2">
        <Text family="mono" variant="body" weight="strong">
          {group.name}
        </Text>
        <Text variant="subtext" theme="neutral" className="truncate">
          {group.label}
        </Text>
      </span>
      {group.version ? (
        <span className="flex min-w-0 items-baseline gap-2">
          <Text family="mono" variant="subtext">
            {group.version}
          </Text>
          <Text variant="subtext" theme="neutral">
            {group.attempts === 1
              ? '1 attempt'
              : `${group.attempts ?? 0} attempts`}
          </Text>
        </span>
      ) : (
        <Text variant="subtext" theme="neutral">
          Not started
        </Text>
      )}
      <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
        {current ? (
          <Status status={currentStatus} />
        ) : (
          STATUS_SUMMARY.map((item) => {
            const count = countStatus(group.installs, item.status)
            if (!count) return null
            return (
              <Status key={item.status} status={item.status}>
                <span>
                  {count} {item.label}
                </span>
              </Status>
            )
          })
        )}
        {pending && !current ? <Status status="pending" /> : null}
      </span>
      {canOpen ? (
        <Button
          className="justify-self-start md:justify-self-end"
          size="sm"
          variant="secondary"
          onClick={open}
        >
          View details
        </Button>
      ) : (
        <span />
      )}
    </li>
  )
}

const INSTALL_SEGMENT_CLASS: Record<TGroupInstallStatus, string> = {
  success: 'h-1 bg-green-600 dark:bg-green-500',
  failed: 'h-1 bg-red-600 dark:bg-red-500',
  'in-progress': 'h-2 bg-blue-600 dark:bg-blue-500',
  'awaiting-approval': 'h-1 bg-orange-600 dark:bg-orange-500',
  pending: 'h-1 bg-cool-grey-200 dark:bg-dark-grey-700',
}

const InstallProgress = ({ group }: { group: TDeploymentGroup }) => {
  const openWorkflow = useInstallWorkflowOpener()
  const unstarted = Math.max(group.plannedCount - group.installs.length, 0)
  const slots = [
    ...group.installs.map((install) => ({
      key: install.id,
      install,
    })),
    ...Array.from({ length: unstarted }, (_, index) => ({
      key: `unstarted-${index}`,
      install: undefined,
    })),
  ]

  return (
    <ol className="flex min-w-0 gap-1" aria-label={`${group.name} installs`}>
      {slots.map((slot) => {
        const status = slot.install?.status ?? 'pending'
        const name = slot.install?.name
        return (
          <li key={slot.key} className="flex min-w-0 flex-1">
            <Tooltip
              position="top"
              className="block"
              style={{ width: '100%' }}
              tipContent={
                slot.install ? (
                  <span className="flex flex-col gap-1">
                    <Text variant="subtext" family="mono">
                      {slot.install.name}
                    </Text>
                    <Status status={slot.install.status} />
                    {slot.install.error ? (
                      <Text variant="subtext">{slot.install.error}</Text>
                    ) : null}
                  </span>
                ) : (
                  <Text variant="subtext">Not started</Text>
                )
              }
            >
              <button
                type="button"
                className={cn(
                  'flex h-6 w-full items-center rounded-sm hover:bg-cool-grey-500/8 focus-visible:outline-none focus-visible:shadow-[0_0_0_1px_var(--background),0_0_0_3px_var(--focus-ring)] dark:hover:bg-white/8',
                  slot.install ? 'cursor-pointer' : 'cursor-default'
                )}
                aria-label={
                  name ? `${name}, ${humanize(status)}` : 'Not started'
                }
                onClick={() => {
                  if (slot.install) openWorkflow(group, slot.install)
                }}
              >
                <span
                  className={cn(
                    'w-full rounded-full',
                    INSTALL_SEGMENT_CLASS[status]
                  )}
                />
              </button>
            </Tooltip>
          </li>
        )
      })}
    </ol>
  )
}

const PlanStages = ({ groups }: { groups: TDeploymentGroup[] }) => {
  const currentIndex = groups.findIndex((group) => !isGroupComplete(group))
  const doneGroups = currentIndex === -1 ? groups.length : currentIndex
  const total = groups.reduce((sum, group) => sum + group.plannedCount, 0)
  const completed = countStatus(
    groups.flatMap((group) => group.installs),
    'success'
  )

  return (
    <div className="flex min-w-0 flex-1 flex-col gap-2 md:flex-row md:items-center md:gap-4">
      <ol
        className="flex min-w-0 flex-1 gap-1"
        aria-label={`Deployment plan, ${doneGroups} of ${groups.length} groups complete`}
      >
        {groups.map((group, index) => {
          const state =
            index < doneGroups
              ? 'complete'
              : index === currentIndex
                ? 'current'
                : 'waiting'
          return (
            <li key={group.id} className="flex min-w-0 flex-1">
              <Tooltip
                position="top"
                className="block"
                style={{ width: '100%' }}
                tipContent={
                  <Text variant="subtext">
                    {group.name} · {group.label}
                  </Text>
                }
              >
                <button
                  type="button"
                  className="flex h-6 w-full cursor-pointer items-center rounded-sm hover:bg-cool-grey-500/8 focus-visible:outline-none focus-visible:shadow-[0_0_0_1px_var(--background),0_0_0_3px_var(--focus-ring)] dark:hover:bg-white/8"
                  aria-current={state === 'current' ? 'step' : undefined}
                  aria-label={`${group.name} ${group.label}, ${state}`}
                  onClick={() => scrollToGroup(group.id)}
                >
                  <span
                    className={cn(
                      'w-full rounded-full',
                      state === 'current' ? 'h-2' : 'h-1',
                      state === 'complete' && 'bg-green-600 dark:bg-green-500',
                      state === 'current' && 'bg-blue-600 dark:bg-blue-500',
                      state === 'waiting' &&
                        'bg-cool-grey-200 dark:bg-dark-grey-700'
                    )}
                  />
                </button>
              </Tooltip>
            </li>
          )
        })}
      </ol>
      <Text variant="subtext" theme="neutral" className="shrink-0">
        {doneGroups} of {groups.length} groups · {completed} of {total} installs
      </Text>
    </div>
  )
}

const ActiveGroupCard = ({
  group,
  index,
  total,
  initialTab,
  autoOpen,
  onInstallsChange,
}: {
  group: TDeploymentGroup
  index: number
  total: number
  initialTab?: 'changes' | 'installs'
  autoOpen?: boolean
  onInstallsChange: (installs: TGroupInstall[]) => void
}) => {
  const { addPanel, panels, removePanel } = useSurfaces()
  const panelKey = `deployment-group-${group.id}`
  const completed = countStatus(group.installs, 'success')
  const awaiting = countStatus(group.installs, 'awaiting-approval')
  const deploying = group.installs.find(
    (install) => install.status === 'in-progress'
  )
  const blocked = !deploying && awaiting > 0
  const position = `Group ${index + 1} of ${total}`

  const open = () => {
    if (panels.some((panel) => panel.key === panelKey && panel.isVisible)) {
      return
    }
    panels
      .filter(
        (panel) =>
          panel.key?.startsWith('deployment-group-') &&
          panel.key !== panelKey &&
          panel.isVisible
      )
      .forEach((panel) => removePanel(panel.id, panel.key))
    addPanel(
      <Panel
        panelKey={panelKey}
        size="half"
        heading={
          <div className="flex flex-col gap-0.5">
            <Text variant="subtext" theme="neutral">
              Group information
            </Text>
            <span className="flex items-baseline gap-2">
              <Text variant="base" weight="strong" family="mono">
                {group.name}
              </Text>
              <Text variant="subtext" theme="neutral">
                {group.label}
              </Text>
            </span>
          </div>
        }
      >
        <GroupDrawer
          group={group}
          initialTab={initialTab}
          onInstallsChange={onInstallsChange}
        />
      </Panel>,
      panelKey
    )
  }

  useEffect(() => {
    if (!autoOpen) return
    const timer = window.setTimeout(open, 0)
    return () => window.clearTimeout(timer)
  }, [])

  const approveAll = () => {
    onInstallsChange(
      group.installs.map((install) =>
        install.status === 'awaiting-approval'
          ? { ...install, status: 'in-progress' as const }
          : install
      )
    )
  }

  return (
    <Card
      elevation="1"
      className="!gap-4 !p-4"
      data-group={group.name}
      data-current="true"
      aria-current="step"
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex min-w-0 flex-col gap-1">
          <Text
            variant="subtext"
            weight="strong"
            theme="info"
            aria-live="polite"
          >
            {blocked ? 'Waiting for approval' : 'Deploying'}
            <Text as="span" variant="subtext" theme="neutral">
              {' '}
              · {position}
            </Text>
          </Text>
          <span className="flex min-w-0 flex-wrap items-baseline gap-2">
            <Text family="mono" variant="body" weight="strong">
              {group.name}
            </Text>
            <Text variant="subtext" theme="neutral">
              {group.label}
            </Text>
          </span>
        </div>
        <Status status={blocked ? 'awaiting-approval' : 'in-progress'} />
      </div>
      <div className="flex flex-col gap-2">
        <span className="flex flex-wrap items-baseline gap-x-3 gap-y-1">
          <Text variant="subtext">
            {completed} of {group.plannedCount} installs completed
          </Text>
          {group.version ? (
            <Text family="mono" variant="subtext" theme="neutral">
              {group.version}
            </Text>
          ) : null}
          {group.startedAt ? (
            <Text variant="subtext" theme="neutral">
              Elapsed{' '}
              <Duration beginTime={group.startedAt} unitDisplay="short" />
            </Text>
          ) : null}
        </span>
        <InstallProgress group={group} />
        <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
          {STATUS_SUMMARY.filter((item) => item.status !== 'success').map(
            (item) => {
              const count = countStatus(group.installs, item.status)
              if (!count) return null
              return (
                <Status key={item.status} status={item.status}>
                  <span>
                    {count} {item.label}
                  </span>
                </Status>
              )
            }
          )}
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
          {awaiting > 0 ? (
            <Button size="sm" variant="primary" onClick={approveAll}>
              Approve all
            </Button>
          ) : null}
          <Button size="sm" variant="secondary" onClick={open}>
            View group details
          </Button>
        </span>
      </div>
    </Card>
  )
}

const GroupTable = ({
  groups,
  currentId,
  initialTab,
  initialGroupId,
  onInstallsChange,
}: {
  groups: TDeploymentGroup[]
  currentId?: string
  initialTab?: 'changes' | 'installs'
  initialGroupId?: string
  onInstallsChange: (groupId: string, installs: TGroupInstall[]) => void
}) => {
  if (!groups.length) return null

  const currentIndex = groups.findIndex((group) => group.id === currentId)

  return (
    <ul className="flex flex-col divide-y overflow-hidden rounded-md border">
      <GroupColumnHeader />
      {groups.map((group, index) => (
        <GroupRow
          key={group.id}
          group={group}
          current={group.id === currentId}
          pending={currentIndex >= 0 && index > currentIndex}
          initialTab={initialTab}
          autoOpen={group.id === initialGroupId && group.id !== currentId}
          onInstallsChange={(installs) => onInstallsChange(group.id, installs)}
        />
      ))}
    </ul>
  )
}

export const BranchDeploymentOverview = ({
  overview,
  initialGroupId,
  initialTab = 'installs',
}: {
  overview: TDeploymentOverview
  initialGroupId?: string
  initialTab?: 'changes' | 'installs'
}) => {
  const { addPanel, panels, removePanel, updatePanel } = useSurfaces()
  const [groups, setGroups] = useState(() =>
    overview.groups.map((group) => ({
      ...group,
      installs: group.installs.map((install) => ({ ...install })),
      changes: group.changes.map((change) => ({ ...change })),
    }))
  )

  const openChanges = () => {
    const group =
      groups.find((item) => !isGroupComplete(item) && item.changes.length) ??
      groups.find((item) => item.changes.length)
    if (!group) return

    const panelKey = `deployment-group-${group.id}`
    const content = (
      <Panel
        panelKey={panelKey}
        size="half"
        heading={
          <div className="flex flex-col gap-0.5">
            <Text variant="subtext" theme="neutral">
              Group information
            </Text>
            <span className="flex items-baseline gap-2">
              <Text variant="base" weight="strong" family="mono">
                {group.name}
              </Text>
              <Text variant="subtext" theme="neutral">
                {group.label}
              </Text>
            </span>
          </div>
        }
      >
        <GroupDrawer
          group={group}
          initialTab="changes"
          onInstallsChange={(installs) =>
            setGroups((current) =>
              current.map((item) =>
                item.id === group.id ? { ...item, installs } : item
              )
            )
          }
        />
      </Panel>
    )
    const existing = panels.find(
      (panel) => panel.key === panelKey && panel.isVisible
    )
    if (existing) {
      updatePanel(existing.id, content)
      return
    }
    panels
      .filter(
        (panel) => panel.key?.startsWith('deployment-group-') && panel.isVisible
      )
      .forEach((panel) => removePanel(panel.id, panel.key))
    addPanel(content, panelKey)
  }

  const currentIndex = groups.findIndex((group) => !isGroupComplete(group))
  const current = currentIndex >= 0 ? groups[currentIndex] : undefined
  const onInstallsChange = (groupId: string, installs: TGroupInstall[]) =>
    setGroups((existing) =>
      existing.map((item) =>
        item.id === groupId ? { ...item, installs } : item
      )
    )

  return (
    <div className="flex h-full flex-col gap-8 overflow-auto p-4 md:p-6">
      <section className="flex flex-col gap-6">
        <SectionHeader title="Overview" />
        <Card elevation="1" className="!gap-0 !p-4">
          <div className="flex flex-wrap items-center gap-4">
            <div className="grid min-w-0 flex-1 items-center gap-x-8 gap-y-4 sm:grid-cols-2 lg:grid-cols-[minmax(0,2fr)_repeat(3,minmax(0,1fr))]">
              <span className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
                <PullRequestLink
                  number={overview.commit.pullRequest.number}
                  repo={overview.commit.repo}
                  label={overview.commit.pullRequest.title}
                  textVariant="body"
                  weight="strong"
                />
                <Text variant="subtext" theme="neutral">
                  into{' '}
                  <Text as="span" family="mono" variant="subtext">
                    {overview.commit.baseBranch}
                  </Text>
                </Text>
              </span>
              <LabeledValue label="Author">
                <span className="flex items-center gap-2">
                  <Avatar
                    name={overview.commit.author}
                    alt=""
                    size="xs"
                    shape="circle"
                  />
                  <Text variant="subtext">{overview.commit.author}</Text>
                </span>
              </LabeledValue>
              <LabeledValue label="Commit">
                <CommitLink
                  sha={overview.commit.sha}
                  repo={overview.commit.repo}
                />
              </LabeledValue>
              <LabeledValue label="Committed">
                <Time time={overview.commit.createdAt} format="relative" />
              </LabeledValue>
            </div>
            <Button
              size="sm"
              variant="secondary"
              className="shrink-0"
              onClick={openChanges}
            >
              View changes
            </Button>
          </div>
        </Card>
      </section>

      <section className="flex flex-col gap-4">
        {current ? (
          <ActiveGroupCard
            group={current}
            index={currentIndex}
            total={groups.length}
            initialTab={initialTab}
            autoOpen={current.id === initialGroupId}
            onInstallsChange={(installs) =>
              onInstallsChange(current.id, installs)
            }
          />
        ) : null}
        <div className="flex flex-col gap-2 md:flex-row md:items-center md:gap-4">
          <Text variant="base" weight="strong" className="shrink-0">
            Deployment plan
          </Text>
          <PlanStages groups={groups} />
        </div>
        <GroupTable
          groups={groups}
          currentId={current?.id}
          initialTab={initialTab}
          initialGroupId={initialGroupId}
          onInstallsChange={onInstallsChange}
        />
      </section>
    </div>
  )
}
