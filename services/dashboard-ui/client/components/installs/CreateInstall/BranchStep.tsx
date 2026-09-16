import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TAppBranch } from '@/types'

interface IBranchStep {
  branches: TAppBranch[]
  selected: TAppBranch | null
  onSelect: (branch: TAppBranch) => void
  onSkip: () => void
}

export const BranchStep = ({
  branches,
  selected,
  onSelect,
  onSkip,
}: IBranchStep) => {
  if (branches.length === 0) {
    return (
      <Banner theme="error">
        This app has no app branches. Refresh and try again.
      </Banner>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <Text variant="subtext" theme="neutral">
        Select the app branch for this install. The branch determines which app
        config and deployment plan this install belongs to.
      </Text>

      <div className="flex flex-col gap-2">
        {branches.map((branch) => {
          const branchId = branch.id || ''
          const isSelected = selected?.id === branchId
          const groupCount =
            branch.configs?.at(0)?.install_groups?.length ?? 0

          return (
            <button
              key={branchId}
              type="button"
              onClick={() => branch.id && onSelect(branch)}
              disabled={!branch.id}
              className={`flex items-center justify-between gap-3 px-3 py-2.5 rounded-md text-left w-full ${
                isSelected
                  ? 'bg-primary-50 dark:bg-primary-900/30 ring-1 ring-primary-500'
                  : 'bg-cool-grey-50 dark:bg-dark-grey-700 hover:bg-cool-grey-100 dark:hover:bg-dark-grey-600'
              }`}
            >
              <div className="flex items-center gap-2 min-w-0">
                <Icon
                  variant="GitBranchIcon"
                  size={14}
                  className="shrink-0 text-cool-grey-400"
                />
                <Text variant="body" weight="strong" className="truncate">
                  {branch.name}
                </Text>
                {groupCount > 0 && (
                  <Text variant="subtext" theme="neutral" className="shrink-0">
                    {groupCount} group{groupCount !== 1 ? 's' : ''}
                  </Text>
                )}
              </div>
              {isSelected && (
                <span className="flex shrink-0 items-center gap-1 text-xs text-primary-700 dark:text-primary-300">
                  <Icon variant="CheckIcon" size={14} />
                  Selected
                </span>
              )}
            </button>
          )
        })}
      </div>

      <div className="flex flex-col items-start gap-1 border-t border-cool-grey-200 pt-4 dark:border-dark-grey-600">
        <Button variant="secondary" onClick={onSkip}>
          Skip app branch
        </Button>
        <Text variant="subtext" theme="neutral">
          This install will use the most recent config from{' '}
          <code>nuon apps sync</code> and will not belong to an app branch.
        </Text>
      </div>
    </div>
  )
}
