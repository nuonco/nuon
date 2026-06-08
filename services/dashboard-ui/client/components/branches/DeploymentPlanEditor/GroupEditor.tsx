import { useState } from 'react'
import { useDroppable } from '@dnd-kit/core'
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable'
import { Badge } from '@/components/common/Badge'
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
import type { IInstallGroup, InstallSelectionMode } from './types'

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
            <div className="flex items-center gap-1 rounded-md bg-cool-grey-100 dark:bg-dark-grey-700 p-0.5">
              <button
                type="button"
                className={cn(
                  'flex-1 text-xs font-medium px-2 py-1 rounded transition-colors',
                  group.selection_mode === 'manual'
                    ? 'bg-white dark:bg-dark-grey-600 shadow-sm'
                    : 'text-cool-grey-600 dark:text-dark-grey-400 hover:text-cool-grey-900'
                )}
                onClick={() => onUpdate({ selection_mode: 'manual' })}
                disabled={disabled}
              >
                Manual
              </button>
              <button
                type="button"
                className={cn(
                  'flex-1 text-xs font-medium px-2 py-1 rounded transition-colors',
                  group.selection_mode === 'labels'
                    ? 'bg-white dark:bg-dark-grey-600 shadow-sm'
                    : 'text-cool-grey-600 dark:text-dark-grey-400 hover:text-cool-grey-900'
                )}
                onClick={() => onUpdate({ selection_mode: 'labels' })}
                disabled={disabled}
              >
                By labels
              </button>
            </div>

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

            <CheckboxInput
              id={`group-preview-${group.id}`}
              checked={group.use_for_previews ?? false}
              onChange={(e) =>
                onUpdate({ use_for_previews: e.target.checked })
              }
              disabled={disabled}
              labelProps={{ labelText: 'Use for previews' }}
            />
          </div>
        </div>

        <div
          ref={setNodeRef}
          className={cn(
            'flex flex-col gap-2 p-4 transition-colors',
            isOver && group.selection_mode === 'manual' && 'bg-primary-50/40 dark:bg-primary-900/10'
          )}
        >
          {group.selection_mode === 'labels' ? (
            <LabelSelectorEditor
              groupId={group.id}
              labelSelector={group.label_selector}
              disabled={disabled}
              onUpdate={(ls) => onUpdate({ label_selector: ls })}
            />
          ) : (
            <>
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
            </>
          )}
        </div>
      </div>
    </div>
  )
}

const LabelSelectorEditor = ({
  groupId,
  labelSelector,
  disabled,
  onUpdate,
}: {
  groupId: string
  labelSelector?: { match_labels: Record<string, string> } | null
  disabled?: boolean
  onUpdate: (ls: { match_labels: Record<string, string> }) => void
}) => {
  const [newKey, setNewKey] = useState('')
  const [newValue, setNewValue] = useState('')

  const labels = labelSelector?.match_labels ?? {}
  const entries = Object.entries(labels)

  const addLabel = () => {
    const key = newKey.trim()
    const value = newValue.trim()
    if (!key) return
    onUpdate({ match_labels: { ...labels, [key]: value } })
    setNewKey('')
    setNewValue('')
  }

  const removeLabel = (key: string) => {
    const next = { ...labels }
    delete next[key]
    onUpdate({ match_labels: next })
  }

  return (
    <div className="flex flex-col gap-3">
      <Text variant="subtext" theme="neutral">
        Installs matching all labels will be included at deploy time.
      </Text>

      {entries.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {entries.map(([key, value]) => (
            <Badge key={key} variant="code" size="md">
              <span className="inline-flex items-center gap-1">
                {key}={value}
                <button
                  type="button"
                  onClick={() => removeLabel(key)}
                  disabled={disabled}
                  className="ml-0.5 hover:text-red-600"
                >
                  <Icon variant="XIcon" size={12} />
                </button>
              </span>
            </Badge>
          ))}
        </div>
      )}

      <div className="flex items-end gap-2">
        <div className="flex-1">
          <Input
            id={`label-key-${groupId}`}
            type="text"
            size="sm"
            placeholder="Key"
            value={newKey}
            onChange={(e) => setNewKey(e.target.value)}
            disabled={disabled}
          />
        </div>
        <div className="flex-1">
          <Input
            id={`label-value-${groupId}`}
            type="text"
            size="sm"
            placeholder="Value"
            value={newValue}
            onChange={(e) => setNewValue(e.target.value)}
            disabled={disabled}
          />
        </div>
        <Button
          variant="secondary"
          size="sm"
          onClick={addLabel}
          disabled={disabled || !newKey.trim()}
        >
          <Icon variant="PlusIcon" size={14} />
          Add
        </Button>
      </div>

      {entries.length === 0 && (
        <div className="px-3 py-3 rounded-md border border-dashed border-cool-grey-300 dark:border-dark-grey-600 text-center">
          <Text variant="subtext" theme="neutral">
            Add label selectors to match installs dynamically
          </Text>
        </div>
      )}
    </div>
  )
}
