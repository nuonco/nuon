import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Checkbox } from '@/components/common/form/CheckboxInput'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { TInstall, TAppBranchInstallGroup } from '@/types'

interface IGroupConfigSection {
  group?: TAppBranchInstallGroup
  availableInstalls: TInstall[]
  onUpdate: (updates: Partial<TAppBranchInstallGroup>) => void
  onDelete: () => void
}

export const GroupConfigSection = ({
  group,
  availableInstalls,
  onUpdate,
  onDelete,
}: IGroupConfigSection) => {
  if (!group) {
    return (
      <div className="w-80 flex flex-col border-l dark:border-dark-grey-700 pl-6">
        <div className="text-center py-12">
          <Icon variant="Gear" size={48} className="mx-auto mb-4 opacity-30" />
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
    <div className="w-80 flex flex-col border-l dark:border-dark-grey-700 pl-6 space-y-6">
      <div>
        <Text variant="h3" weight="strong">
          Configure group
        </Text>
      </div>

      <Input
        id="group-name"
        type="text"
        value={group.name || ''}
        onChange={(e) => onUpdate({ name: e.target.value })}
        placeholder="Enter group name"
        labelProps={{ labelText: 'Name' }}
      />

      <Input
        id="group-max-parallel"
        type="number"
        min={1}
        value={group.max_parallel || 1}
        onChange={(e) => onUpdate({ max_parallel: parseInt(e.target.value) || 1 })}
        labelProps={{ labelText: 'Max parallel deployments' }}
        helperText="Maximum number of installs to deploy in parallel"
      />

      <div className="space-y-1">
        <label className="flex gap-3 items-start rounded-md p-2 hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer">
          <Checkbox
            id="requires-approval"
            checked={group.requires_approval || false}
            onChange={(e) => onUpdate({ requires_approval: e.target.checked })}
            className="mt-1.5 shrink-0"
          />
          <div className="flex flex-col gap-0.5">
            <Text variant="body" weight="strong">Requires approval</Text>
            <Text variant="subtext" theme="neutral">Wait for manual approval before deploying this group</Text>
          </div>
        </label>
        <label className="flex gap-3 items-start rounded-md p-2 hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer">
          <Checkbox
            id="rollback-on-failure"
            checked={group.rollback_on_failure || false}
            onChange={(e) => onUpdate({ rollback_on_failure: e.target.checked })}
            className="mt-1.5 shrink-0"
          />
          <div className="flex flex-col gap-0.5">
            <Text variant="body" weight="strong">Rollback on failure</Text>
            <Text variant="subtext" theme="neutral">Automatically rollback if any deployment in this group fails</Text>
          </div>
        </label>
      </div>

      <div className="flex flex-col gap-2">
        <Text variant="base" weight="strong" className="block">
          Installs in group ({groupInstalls.length})
        </Text>
        {groupInstalls.length > 0 ? (
          <div className="space-y-1 max-h-40 overflow-y-auto">
            {groupInstalls.map((install) => (
              <div
                key={install.id}
                className="flex items-center gap-2 px-3 py-2 dark:bg-dark-grey-800 rounded-md"
              >
                <Icon variant="Cloud" size={14} />
                <Text variant="subtext" className="flex-1 truncate">
                  {install.name}
                </Text>
              </div>
            ))}
          </div>
        ) : (
          <Text variant="subtext" theme="neutral" className="block">
            No installs in this group
          </Text>
        )}
      </div>

      <div className="!mt-auto pt-6 border-t dark:border-dark-grey-700">
        <Button
          onClick={onDelete}
          variant="danger"
          className="w-full items-center"
        >
          <Icon variant="Trash" size={16} />
          Delete group
        </Button>
      </div>
    </div>
  )
}
