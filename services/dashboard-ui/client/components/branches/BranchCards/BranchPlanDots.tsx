import { Badge } from '@/components/common/Badge'

export type TBranchPlanGroup = {
  name: string
  installs: number
  hasSelector: boolean
  isPreview?: boolean
}

const MAX_VISIBLE_GROUPS = 4

export const PreviewBadge = () => (
  <Badge size="xs" theme="info" className="shrink-0">
    Preview
  </Badge>
)

export const BranchPlanDots = ({ groups }: { groups: TBranchPlanGroup[] }) => {
  const visibleGroups = groups.slice(0, MAX_VISIBLE_GROUPS)
  const hiddenGroups = groups.length - visibleGroups.length

  return (
    <div className="flex items-center gap-1.5 min-w-0">
      {visibleGroups.map((group, index) => (
        <div
          key={`${group.name}-${index}`}
          className="flex items-center gap-1.5 min-w-0"
          title={
            group.hasSelector
              ? `${group.name}: installs matched by label`
              : `${group.name}: ${group.installs} ${
                  group.installs === 1 ? 'install' : 'installs'
                }`
          }
        >
          {index > 0 ? (
            <span className="h-px w-3 shrink-0 bg-cool-grey-300 dark:bg-dark-grey-600" />
          ) : null}
          <span className="h-2 w-2 shrink-0 rounded-full bg-primary-500" />
          <span className="truncate text-xs text-cool-grey-700 dark:text-cool-grey-300">
            {group.name}
          </span>
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
