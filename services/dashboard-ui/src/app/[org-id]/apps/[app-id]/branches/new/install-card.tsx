'use client'

import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { cn } from '@/utils/classnames'

export interface IInstall {
  id: string
  name: string
  region: string
  status: 'active' | 'inactive'
  platform: 'aws' | 'azure' | 'gcp'
}

interface IInstallCardProps {
  install: IInstall
  onMoveToGroup?: (groupId: string) => void
  onMoveToUngrouped?: () => void
  availableGroups?: string[]
  availableGroupNames?: string[]
  isInGroup?: boolean
  showActions?: boolean
}

export const InstallCard = ({
  install,
  onMoveToGroup,
  onMoveToUngrouped,
  availableGroups = [],
  availableGroupNames = [],
  isInGroup = false,
  showActions = true,
}: IInstallCardProps) => {
  return (
    <div
      className={cn(
        'flex items-center gap-3 p-3 border rounded-lg bg-white dark:bg-dark-grey-800',
        'hover:border-primary-400 dark:hover:border-primary-500 transition-colors'
      )}
    >
      {/* Drag Handle Icon */}
      <div className="flex-shrink-0 text-cool-grey-400 dark:text-cool-grey-600">
        <Icon variant="DotsSixVertical" size={20} />
      </div>

      {/* Install Info */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <Icon variant="AWS" size={16} className="flex-shrink-0" />
          <Text variant="sm" weight="strong" className="truncate">
            {install.name}
          </Text>
        </div>
        <div className="flex items-center gap-2">
          <Icon
            variant="MapPin"
            size={12}
            className="flex-shrink-0 text-cool-grey-500"
          />
          <Text
            variant="xs"
            className="text-cool-grey-600 dark:text-cool-grey-400"
          >
            {install.region}
          </Text>
        </div>
      </div>

      {/* Status Badge */}
      <div className="flex-shrink-0">
        <Badge
          size="sm"
          theme={install.status === 'active' ? 'success' : 'neutral'}
        >
          {install.status}
        </Badge>
      </div>

      {/* Action Buttons */}
      {showActions && (
        <div className="flex-shrink-0">
          {isInGroup ? (
            <Button
              size="sm"
              variant="ghost"
              onClick={onMoveToUngrouped}
              title="Remove from group"
            >
              <Icon variant="X" size={16} />
            </Button>
          ) : (
            <div className="flex items-center gap-1">
              {availableGroups.map((groupId, index) => (
                <Button
                  key={groupId}
                  size="sm"
                  variant="ghost"
                  onClick={() => onMoveToGroup?.(groupId)}
                  title={`Move to ${availableGroupNames[index] || `Group ${index + 1}`}`}
                >
                  <Icon variant="ArrowRight" size={14} />
                  <Text variant="xs">{index + 1}</Text>
                </Button>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}