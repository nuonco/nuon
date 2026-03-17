import { useDroppable } from '@dnd-kit/core'
import { Badge } from '@/components/common/Badge'
import { Text } from '@/components/common/Text'
import { Icon } from '@/components/common/Icon'
import { Button } from '@/components/common/Button'
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
          'border-blue-400 dark:border-blue-500 bg-blue-50 dark:bg-blue-900/20': isSelected,
          'border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 hover:border-gray-400 dark:hover:border-gray-500': !isSelected,
          'border-green-400 dark:border-green-500 bg-green-50 dark:bg-green-900/20': isOver,
        }
      )}
    >
      {/* Group header */}
      <div className="flex items-center justify-between mb-3">
        <Text variant="base" weight="strong">
          {index + 1}. {group.name}
        </Text>
        <div className="flex items-center gap-2">
          {group.requires_approval && (
            <Badge theme="warning" size="sm">
              Requires Approval
            </Badge>
          )}
          {group.rollback_on_failure && (
            <Badge theme="info" size="sm">
              Rollback on Failure
            </Badge>
          )}
          <Text variant="subtext" theme="neutral">
            Max {group.max_parallel} parallel
          </Text>
        </div>
      </div>

      {/* Install cards */}
      <div className="space-y-2">
        {installs.length > 0 ? (
          installs.slice(0, 5).map((install) => (
            <button
              key={install.id}
              onClick={onClick}
              className={cn(
                'w-full flex items-center justify-between gap-3 px-4 py-2.5',
                'rounded-md border-2 transition-all text-left',
                'cursor-pointer select-none',
                'hover:-translate-y-0.5 hover:shadow-md',
                {
                  'border-blue-300 dark:border-blue-600 bg-blue-50 dark:bg-blue-900/30': isSelected,
                  'border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900': !isSelected,
                }
              )}
            >
              <div className="flex items-center gap-3 flex-1 min-w-0">
                <Icon variant="Cloud" size={16} />
                <Text variant="base" className="truncate">
                  {install.name}
                </Text>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={(e) => {
                  e.stopPropagation()
                  onRemoveInstall(install.id)
                }}
              >
                <Icon variant="X" size={14} />
              </Button>
            </button>
          ))
        ) : (
          <div className="px-4 py-8 bg-gray-50 dark:bg-gray-900 rounded-md text-center border-2 border-dashed border-gray-300 dark:border-gray-600">
            <Text variant="subtext" theme="neutral">
              Drag installs here to add them to this group
            </Text>
          </div>
        )}

        {installs.length > 5 && (
          <div className="px-4 py-2.5 bg-gray-100 dark:bg-gray-800 rounded-md text-center">
            <Text variant="subtext">+{installs.length - 5} more</Text>
          </div>
        )}
      </div>
    </div>
  )
}
