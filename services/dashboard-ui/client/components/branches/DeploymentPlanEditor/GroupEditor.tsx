import { useDroppable } from '@dnd-kit/core'
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import type { TInstall } from '@/types'
import { cn } from '@/utils/classnames'
import { AddInstallPicker } from './AddInstallPicker'
import { SortableInstallRow } from './SortableInstallRow'
import type { IInstallGroup } from './types'

interface IGroupEditor {
  group: IInstallGroup
  index: number
  totalGroups: number
  installs: TInstall[]
  unassignedInstalls: TInstall[]
  disabled?: boolean
  nameError?: string
  onUpdate: (updates: Partial<IInstallGroup>) => void
  onAddInstalls: (installIds: string[]) => void
  onRemoveInstall: (installId: string) => void
  onMoveUp: () => void
  onMoveDown: () => void
  onDelete: () => void
}

export const GroupEditor = ({
  group,
  index,
  totalGroups,
  installs,
  unassignedInstalls,
  disabled,
  nameError,
  onUpdate,
  onAddInstalls,
  onRemoveInstall,
  onMoveUp,
  onMoveDown,
  onDelete,
}: IGroupEditor) => {
  const { setNodeRef, isOver } = useDroppable({
    id: group.id,
    data: { containerId: group.id, type: 'container' },
  })

  return (
    <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-lg bg-white dark:bg-dark-grey-800">
      <div className="grid grid-cols-1 md:grid-cols-[minmax(220px,280px)_1fr] divide-y md:divide-y-0 md:divide-x divide-cool-grey-200 dark:divide-dark-grey-700">
        <div className="flex flex-col gap-4 p-4">
          <div className="flex items-start gap-2">
            <div className="flex-1 min-w-0">
              <Input
                id={`group-name-${group.id}`}
                type="text"
                value={group.name}
                onChange={(e) => onUpdate({ name: e.target.value })}
                placeholder={`Group ${index + 1}`}
                disabled={disabled}
                size="sm"
                className="!font-bold"
                error={!!nameError}
                errorMessage={nameError}
              />
            </div>

            <Dropdown
              id={`group-menu-${group.id}`}
              variant="ghost"
              alignment="right"
              hideIcon
              disabled={disabled}
              buttonClassName="!p-2"
              buttonText={<Icon variant="DotsThreeVerticalIcon" size={16} />}
            >
              <Menu>
                <Button isMenuButton onClick={onMoveUp} disabled={index === 0}>
                  Move up
                  <Icon variant="ArrowUpIcon" />
                </Button>
                <Button
                  isMenuButton
                  onClick={onMoveDown}
                  disabled={index === totalGroups - 1}
                >
                  Move down
                  <Icon variant="ArrowDownIcon" />
                </Button>
                <hr />
                <Button
                  isMenuButton
                  className="!text-red-800 dark:!text-red-500"
                  onClick={onDelete}
                >
                  Delete group
                  <Icon variant="TrashIcon" />
                </Button>
              </Menu>
            </Dropdown>
          </div>

          <div className="flex flex-col gap-2">
            <div className="flex items-center gap-2">
              <Text variant="subtext" theme="neutral">
                Max parallel
              </Text>
              <Input
                id={`group-max-parallel-${group.id}`}
                type="number"
                min={1}
                value={group.max_parallel ?? 1}
                onChange={(e) =>
                  onUpdate({ max_parallel: parseInt(e.target.value) || 1 })
                }
                disabled={disabled}
                size="sm"
                className="!w-16"
              />
            </div>

            <CheckboxInput
              id={`group-requires-approval-${group.id}`}
              checked={group.requires_approval ?? false}
              onChange={(e) =>
                onUpdate({ requires_approval: e.target.checked })
              }
              disabled={disabled}
              labelProps={{ labelText: 'Requires approval' }}
            />

            <CheckboxInput
              id={`group-rollback-${group.id}`}
              checked={group.rollback_on_failure ?? false}
              onChange={(e) =>
                onUpdate({ rollback_on_failure: e.target.checked })
              }
              disabled={disabled}
              labelProps={{ labelText: 'Rollback on failure' }}
            />
          </div>
        </div>

        <div
          ref={setNodeRef}
          className={cn(
            'flex flex-col gap-2 p-4 transition-colors',
            isOver && 'bg-primary-50/40 dark:bg-primary-900/10'
          )}
        >
          <SortableContext
            items={group.install_ids}
            strategy={verticalListSortingStrategy}
          >
            {installs.length > 0 ? (
              <div className="flex flex-col gap-1.5">
                {installs.map((install) => (
                  <SortableInstallRow
                    key={install.id}
                    installId={install.id}
                    installName={install.name || install.id}
                    containerId={group.id}
                    disabled={disabled}
                    showRemove
                    onRemove={() => onRemoveInstall(install.id)}
                  />
                ))}
              </div>
            ) : (
              <div className="px-3 py-3 rounded-md border border-dashed border-cool-grey-300 dark:border-dark-grey-600 text-center">
                <Text variant="subtext" theme="neutral">
                  Drop installs here or use Add install
                </Text>
              </div>
            )}
          </SortableContext>

          <AddInstallPicker
            groupId={group.id}
            unassignedInstalls={unassignedInstalls}
            disabled={disabled}
            onAdd={onAddInstalls}
          />
        </div>
      </div>
    </div>
  )
}
