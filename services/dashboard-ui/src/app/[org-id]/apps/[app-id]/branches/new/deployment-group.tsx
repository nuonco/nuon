'use client'

import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { InstallCard, IInstall } from './install-card'
import { cn } from '@/utils/classnames'

interface IDeploymentGroupProps {
  groupIndex: number
  installs: IInstall[]
  onRemoveInstall: (installId: string) => void
  onDeleteGroup?: () => void
  canDelete?: boolean
}

export const DeploymentGroup = ({
  groupIndex,
  installs,
  onRemoveInstall,
  onDeleteGroup,
  canDelete = true,
}: IDeploymentGroupProps) => {
  return (
    <div
      className={cn(
        'border-2 border-dashed rounded-lg p-4',
        installs.length > 0
          ? 'border-primary-300 dark:border-primary-700 bg-primary-50/30 dark:bg-primary-950/20'
          : 'border-cool-grey-300 dark:border-dark-grey-600 bg-cool-grey-50/50 dark:bg-dark-grey-800/50'
      )}
    >
      {/* Group Header */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <div
            className={cn(
              'w-8 h-8 rounded-full flex items-center justify-center text-sm font-strong',
              installs.length > 0
                ? 'bg-primary-600 text-white'
                : 'bg-cool-grey-300 dark:bg-dark-grey-600 text-cool-grey-700 dark:text-cool-grey-300'
            )}
          >
            {groupIndex + 1}
          </div>
          <div>
            <Text variant="sm" weight="strong">
              Group {groupIndex + 1}
            </Text>
            <Text
              variant="xs"
              className="text-cool-grey-600 dark:text-cool-grey-400"
            >
              {installs.length} install{installs.length !== 1 ? 's' : ''} •
              Deploy in parallel
            </Text>
          </div>
        </div>

        {canDelete && (
          <Button
            size="sm"
            variant="ghost"
            onClick={onDeleteGroup}
            title="Delete group"
          >
            <Icon variant="Trash" size={16} />
          </Button>
        )}
      </div>

      {/* Installs in Group */}
      {installs.length > 0 ? (
        <div className="space-y-2">
          {installs.map((install) => (
            <InstallCard
              key={install.id}
              install={install}
              isInGroup
              onMoveToUngrouped={() => onRemoveInstall(install.id)}
            />
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center py-8 text-center">
          <Icon
            variant="Package"
            size={32}
            className="text-cool-grey-400 dark:text-cool-grey-600 mb-2"
          />
          <Text
            variant="sm"
            className="text-cool-grey-600 dark:text-cool-grey-400"
          >
            No installs in this group
          </Text>
          <Text
            variant="xs"
            className="text-cool-grey-500 dark:text-cool-grey-500 mt-1"
          >
            Use the arrows on ungrouped installs below to add them here
          </Text>
        </div>
      )}
    </div>
  )
}
