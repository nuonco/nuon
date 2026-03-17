import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { TInstall, TAppBranchInstallGroup } from '@/types'

interface IGroupConfigPanel {
  group?: TAppBranchInstallGroup
  availableInstalls: TInstall[]
  onUpdate: (updates: Partial<TAppBranchInstallGroup>) => void
  onDelete: () => void
}

export const GroupConfigPanel = ({
  group,
  availableInstalls,
  onUpdate,
  onDelete,
}: IGroupConfigPanel) => {
  if (!group) {
    return (
      <div className="w-80 flex flex-col border-l dark:border-gray-700 pl-6">
        <div className="text-center py-12">
          <Icon variant="Settings" size={48} className="mx-auto mb-4 opacity-30" />
          <Text variant="base" theme="neutral">
            Select a group to configure
          </Text>
        </div>
      </div>
    )
  }

  const groupInstalls = availableInstalls.filter((i) =>
    group.install_ids?.includes(i.id)
  )

  return (
    <div className="w-80 flex flex-col border-l dark:border-gray-700 pl-6 space-y-6">
      <div>
        <Text variant="h4" weight="strong">
          Configure Group
        </Text>
      </div>

      {/* Group Name */}
      <div className="space-y-2">
        <Text variant="base" weight="strong">
          Name
        </Text>
        <Input
          type="text"
          value={group.name || ''}
          onChange={(e) => onUpdate({ name: e.target.value })}
          placeholder="Enter group name"
        />
      </div>

      {/* Max Parallel */}
      <div className="space-y-2">
        <Text variant="base" weight="strong">
          Max Parallel Deployments
        </Text>
        <Input
          type="number"
          min={1}
          value={group.max_parallel || 1}
          onChange={(e) =>
            onUpdate({ max_parallel: parseInt(e.target.value) || 1 })
          }
        />
        <Text variant="subtext" theme="neutral">
          Maximum number of installs to deploy in parallel
        </Text>
      </div>

      {/* Options */}
      <div className="space-y-3">
        <label className="flex items-center gap-3 cursor-pointer">
          <input
            type="checkbox"
            checked={group.requires_approval || false}
            onChange={(e) => onUpdate({ requires_approval: e.target.checked })}
            className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
          />
          <div className="flex-1">
            <Text variant="base" weight="strong">
              Requires Approval
            </Text>
            <Text variant="subtext" theme="neutral">
              Wait for manual approval before deploying this group
            </Text>
          </div>
        </label>

        <label className="flex items-center gap-3 cursor-pointer">
          <input
            type="checkbox"
            checked={group.rollback_on_failure || false}
            onChange={(e) => onUpdate({ rollback_on_failure: e.target.checked })}
            className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
          />
          <div className="flex-1">
            <Text variant="base" weight="strong">
              Rollback on Failure
            </Text>
            <Text variant="subtext" theme="neutral">
              Automatically rollback if any deployment in this group fails
            </Text>
          </div>
        </label>
      </div>

      {/* Installs in Group */}
      <div className="space-y-2">
        <Text variant="base" weight="strong">
          Installs in Group ({groupInstalls.length})
        </Text>
        {groupInstalls.length > 0 ? (
          <div className="space-y-1 max-h-40 overflow-y-auto">
            {groupInstalls.map((install) => (
              <div
                key={install.id}
                className="flex items-center gap-2 px-3 py-2 bg-gray-50 dark:bg-gray-900 rounded-md"
              >
                <Icon variant="Cloud" size={14} />
                <Text variant="subtext" className="flex-1 truncate">
                  {install.name}
                </Text>
              </div>
            ))}
          </div>
        ) : (
          <Text variant="subtext" theme="neutral">
            No installs in this group
          </Text>
        )}
      </div>

      {/* Delete Group */}
      <div className="pt-6 border-t dark:border-gray-700">
        <Button
          onClick={onDelete}
          variant="danger"
          size="sm"
          className="w-full"
        >
          <Icon variant="Trash" size={16} />
          Delete Group
        </Button>
      </div>
    </div>
  )
}
