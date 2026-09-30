import { Badge } from '@/components/common/Badge'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { cn } from '@/utils/classnames'
import { getStatusTheme, type TStatusTheme } from '@/utils/status-utils'
import { humanize } from '@/utils/string-utils'

export type TBranchPlanGroup = {
  name: string
  installs: number
  hasSelector: boolean
  isPreview?: boolean
  status?: string
}

const MAX_VISIBLE_GROUPS = 4

const DOT_THEME_CLASSES: Record<TStatusTheme, string> = {
  success: 'bg-green-600 dark:bg-green-500',
  error: 'bg-red-600 dark:bg-red-500',
  warn: 'bg-orange-600 dark:bg-orange-500',
  info: 'bg-blue-600 dark:bg-blue-500',
  brand: 'bg-primary-600 dark:bg-primary-400',
  neutral: 'bg-cool-grey-400 dark:bg-dark-grey-500',
}

export const PreviewBadge = () => (
  <Badge size="xs" theme="info" className="shrink-0">
    Preview
  </Badge>
)

const installSummary = (group: TBranchPlanGroup) =>
  group.hasSelector
    ? 'Installs matched by label'
    : `${group.installs} ${group.installs === 1 ? 'install' : 'installs'}`

const GroupTip = ({ group }: { group: TBranchPlanGroup }) => (
  <div className="flex flex-col gap-0.5">
    <Text variant="subtext" weight="strong" family="mono">
      {group.name}
    </Text>
    {group.status ? (
      <Text variant="label" theme="neutral">
        Status: {humanize(group.status)}
      </Text>
    ) : null}
    <Text variant="label" theme="neutral">
      {installSummary(group)}
    </Text>
  </div>
)

const GroupDot = ({ group }: { group: TBranchPlanGroup }) => {
  const dot = (
    <span
      className={cn(
        'block h-2 w-2 shrink-0 rounded-full',
        group.status
          ? DOT_THEME_CLASSES[getStatusTheme(group.status)]
          : 'bg-primary-500'
      )}
    />
  )
  const label = (
    <span className="truncate text-xs text-cool-grey-700 dark:text-cool-grey-300">
      {group.name}
    </span>
  )

  if (!group.status) {
    return (
      <span
        className="flex items-center gap-1.5 min-w-0"
        title={`${group.name}: ${installSummary(group).toLowerCase()}`}
      >
        {dot}
        {label}
      </span>
    )
  }

  return (
    <Tooltip
      position="top"
      tipContent={<GroupTip group={group} />}
      className="flex items-center gap-1.5 min-w-0"
    >
      {dot}
      {label}
    </Tooltip>
  )
}

export const BranchPlanDots = ({ groups }: { groups: TBranchPlanGroup[] }) => {
  const visibleGroups = groups.slice(0, MAX_VISIBLE_GROUPS)
  const hiddenGroups = groups.length - visibleGroups.length

  return (
    <div className="flex items-center gap-1.5 min-w-0">
      {visibleGroups.map((group, index) => (
        <div
          key={`${group.name}-${index}`}
          className="flex items-center gap-1.5 min-w-0"
        >
          {index > 0 ? (
            <span className="h-px w-3 shrink-0 bg-cool-grey-300 dark:bg-dark-grey-600" />
          ) : null}
          <GroupDot group={group} />
          {group.isPreview ? <PreviewBadge /> : null}
        </div>
      ))}
      {hiddenGroups > 0 ? (
        <>
          <span className="h-px w-3 shrink-0 bg-cool-grey-300 dark:bg-dark-grey-600" />
          <span
            className="shrink-0 text-xs text-cool-grey-500"
            title={groups
              .slice(MAX_VISIBLE_GROUPS)
              .map((group) => group.name)
              .join(', ')}
          >
            +{hiddenGroups} more
          </span>
        </>
      ) : null}
    </div>
  )
}
