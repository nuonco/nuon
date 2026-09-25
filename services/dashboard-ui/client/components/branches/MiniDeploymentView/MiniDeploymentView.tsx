import { HealthBars, type IHealthBar } from '@/components/common/HealthBars'
import { Text } from '@/components/common/Text'
import {
  PreviewBadge,
  type TBranchPlanGroup,
} from '@/components/branches/BranchCards/BranchPlanDots'
import { getStatusTheme } from '@/utils/status-utils'
import { humanize } from '@/utils/string-utils'

export type TMiniDeployInstall = {
  id: string
  name: string
  group: string
  href?: string
  runStatus?: string
  health?: string
  rolledOut?: boolean
}

export interface IMiniDeploymentView {
  groups: TBranchPlanGroup[]
  installs: TMiniDeployInstall[]
  showRollout?: boolean
}

const BAR_CLASSES: Record<string, string> = {
  success: 'bg-green-600 dark:bg-green-500',
  error: 'bg-red-600 dark:bg-red-500',
  warn: 'bg-orange-500 dark:bg-orange-400',
  info: 'bg-blue-600 dark:bg-blue-500',
  brand: 'bg-primary-600 dark:bg-primary-400',
  neutral: 'bg-cool-grey-300 dark:bg-dark-grey-600',
}

const barClass = (install: TMiniDeployInstall) =>
  BAR_CLASSES[getStatusTheme(install.runStatus ?? 'unknown')] ??
  BAR_CLASSES.neutral

const rolloutLabel = (install: TMiniDeployInstall) =>
  install.rolledOut ? 'Rolled out' : 'Not rolled out yet'

const summarize = (installs: TMiniDeployInstall[]) => {
  const rolledOut = installs.filter((install) => install.rolledOut).length
  return `${rolledOut}/${installs.length} installs updated`
}

const orderGroups = (
  groups: TBranchPlanGroup[],
  installs: TMiniDeployInstall[]
) => {
  const byGroup = new Map<string, TMiniDeployInstall[]>()
  for (const install of installs) {
    const existing = byGroup.get(install.group)
    if (existing) existing.push(install)
    else byGroup.set(install.group, [install])
  }

  const named = groups.map((group) => ({
    name: group.name,
    hasSelector: group.hasSelector,
    isPreview: group.isPreview,
    installs: byGroup.get(group.name) ?? [],
  }))

  const extras = [...byGroup.keys()]
    .filter((name) => !groups.some((group) => group.name === name))
    .map((name) => ({
      name,
      hasSelector: false,
      isPreview: false,
      installs: byGroup.get(name) ?? [],
    }))

  return [...named, ...extras]
}

const InstallTooltip = ({
  install,
  showRollout,
}: {
  install: TMiniDeployInstall
  showRollout: boolean
}) => (
  <div className="flex flex-col gap-0.5 w-44">
    <Text variant="subtext" weight="strong" family="mono">
      {install.name}
    </Text>
    {install.runStatus ? (
      <Text variant="label" theme="neutral">
        Run: {humanize(install.runStatus)}
      </Text>
    ) : null}
    {install.health ? (
      <Text variant="label" theme="neutral">
        Health: {humanize(install.health)}
      </Text>
    ) : null}
    {showRollout ? (
      <Text variant="label" theme="neutral">
        {rolloutLabel(install)}
      </Text>
    ) : null}
  </div>
)

const GroupBars = ({
  installs,
  showRollout,
}: {
  installs: TMiniDeployInstall[]
  showRollout: boolean
}) => {
  const bars: IHealthBar[] = installs.map((install) => ({
    key: install.id,
    colorClass: barClass(install),
    ariaLabel: `${install.name}: ${humanize(install.runStatus ?? 'unknown')}`,
    tooltip: <InstallTooltip install={install} showRollout={showRollout} />,
  }))

  return (
    <HealthBars
      grow
      bars={bars}
      className={installs.length === 1 ? 'w-12' : 'w-full'}
      barClassName="h-5 rounded-xs"
      emptyMessage="No installs in this group"
    />
  )
}

const GroupLabel = ({
  name,
  hasSelector,
  isPreview,
}: {
  name: string
  hasSelector: boolean
  isPreview?: boolean
}) => (
  <div className="flex items-center gap-2 min-w-0">
    <Text variant="label" theme="neutral" family="mono" className="truncate">
      {name}
    </Text>
    {isPreview ? <PreviewBadge /> : null}
    {hasSelector ? (
      <Text variant="label" theme="neutral">
        matched by label
      </Text>
    ) : null}
  </div>
)

export const MiniDeploymentView = ({
  groups,
  installs,
  showRollout = false,
}: IMiniDeploymentView) => {
  const ordered = orderGroups(groups, installs)
  if (ordered.length === 0) return null

  return (
    <div className="flex flex-col gap-4 min-w-0">
      {ordered.map((group) => (
        <div
          key={group.name}
          className="flex flex-col gap-1.5 min-w-0"
          aria-label={`${group.name}: ${summarize(group.installs)}`}
        >
          <div className="flex items-center justify-between gap-3">
            <GroupLabel
              name={group.name}
              hasSelector={group.hasSelector}
              isPreview={group.isPreview}
            />
            {group.installs.length > 0 ? (
              <Text
                variant="label"
                theme="neutral"
                className="shrink-0 text-right"
              >
                {summarize(group.installs)}
              </Text>
            ) : null}
          </div>
          <GroupBars installs={group.installs} showRollout={showRollout} />
        </div>
      ))}
    </div>
  )
}
