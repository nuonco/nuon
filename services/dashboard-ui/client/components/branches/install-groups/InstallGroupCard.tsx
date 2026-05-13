import { useDroppable } from '@dnd-kit/core'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Icon } from '@/components/common/Icon'
import { cn } from '@/utils/classnames'
import type { TInstall, TAppBranchInstallGroup } from '@/types'

interface IInstallGroupCard {
  group: TAppBranchInstallGroup
  installs: TInstall[]
  isSelected: boolean
  onClick: () => void
  onRemoveInstall: (installId: string) => void
  index: number
}

export const InstallGroupCard = ({
  group,
  installs,
  isSelected,
  onClick,
  onRemoveInstall,
  index,
}: IInstallGroupCard) => {
  const { setNodeRef, isOver } = useDroppable({
    id: group.id || '',
  })

  return (
    <div
      ref={setNodeRef}
      className={cn(
        'p-4 rounded-lg border-2 transition-all',
        {
          'border-primary-600 dark:border-primary-500 bg-primary-900/10 dark:bg-primary-900/20': isSelected,
          'border-cool-grey-300 dark:border-dark-grey-600 bg-white dark:bg-dark-grey-800 hover:border-cool-grey-400 dark:hover:border-dark-grey-500': !isSelected,
          'border-green-500 dark:border-green-400 bg-green-900/10': isOver,
        }
      )}
    >
      <div className="flex items-center justify-between mb-3">
        <Text variant="base" weight="strong">
          {index + 1}. {group.name}
        </Text>
        <div className="flex items-center gap-2">
          {group.requires_approval && (
            <Badge theme="warn" size="sm">
              Requires approval
            </Badge>
          )}
          {group.rollback_on_failure && (
            <Badge theme="info" size="sm">
              Rollback on failure
            </Badge>
          )}
          <Text variant="subtext" theme="neutral">
            Max {group.max_parallel} parallel
          </Text>
        </div>
      </div>

      <div className="space-y-2">
        {installs.length > 0 ? (
          installs.slice(0, 5).map((install) => (
            <Button
              key={install.id}
              variant="ghost"
              onClick={onClick}
              className={cn(
                'w-full flex items-center justify-between gap-3 px-4 py-2.5',
                'rounded-md border-2 transition-all',
                'hover:-translate-y-0.5 hover:shadow-md',
                {
                  'border-primary-400 dark:border-primary-600 bg-primary-900/10 dark:bg-primary-900/30': isSelected,
                  'border-cool-grey-200 dark:border-dark-grey-700 bg-cool-grey-50 dark:bg-dark-grey-900': !isSelected,
                }
              )}
            >
              <div className="flex items-center gap-3 flex-1 min-w-0">
                <Icon variant="CloudIcon" size={16} />
                <Text variant="base" className="truncate">
                  {install.name}
                </Text>
              </div>
              <Button
                variant="ghost"
                size="xs"
                onClick={(e) => {
                  e.stopPropagation()
                  onRemoveInstall(install.id)
                }}
              >
                <Icon variant="XIcon" size={14} />
              </Button>
            </Button>
          ))
        ) : (
          <div className="px-4 py-8 dark:bg-dark-grey-900/50 rounded-md text-center border-2 border-dashed dark:border-dark-grey-600">
            <Text variant="subtext" theme="neutral">
              Drag installs here to add them to this group
            </Text>
          </div>
        )}

        {installs.length > 5 && (
          <div className="px-4 py-2.5 dark:bg-dark-grey-800 rounded-md text-center">
            <Text variant="subtext">+{installs.length - 5} more</Text>
          </div>
        )}
      </div>
    </div>
  )
}
