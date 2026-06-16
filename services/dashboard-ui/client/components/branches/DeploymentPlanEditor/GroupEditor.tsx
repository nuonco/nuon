import { useMemo, useState } from 'react'
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
import type { IInstallGroup, ILabelSelector } from './types'

interface IGroupEditor {
  group: IInstallGroup
  index: number
  totalGroups: number
  unassignedInstalls: TInstall[]
  availableInstalls: TInstall[]
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
  unassignedInstalls,
  availableInstalls,
  disabled,
  nameError,
  onUpdate,
  onAddInstalls,
  onRemoveInstall,
  onMoveUp,
  onMoveDown,
  onDelete,
}: IGroupEditor) => {
  const installs = useMemo(() => {
    const byId = new Map(availableInstalls.map((i) => [i.id, i]))
    return group.install_ids.map((id) => byId.get(id)).filter((i): i is TInstall => !!i)
  }, [group.install_ids, availableInstalls])

  return (
    <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-lg bg-white dark:bg-dark-grey-800">
      {/* Header row: group name + menu */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-cool-grey-200 dark:border-dark-grey-700">
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

      {/* Settings row */}
      <div className="flex flex-wrap items-center gap-x-4 gap-y-2 px-4 py-3 border-b border-cool-grey-200 dark:border-dark-grey-700">
        <div className="flex items-center gap-1 rounded-md bg-cool-grey-100 dark:bg-dark-grey-700 p-0.5">
          <button
            type="button"
            className={cn(
              'text-xs font-medium px-2 py-1 rounded transition-colors',
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
              'text-xs font-medium px-2 py-1 rounded transition-colors',
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
          id={`group-preview-${group.id}`}
          checked={group.use_for_previews ?? false}
          onChange={(e) =>
            onUpdate({ use_for_previews: e.target.checked })
          }
          disabled={disabled}
          labelProps={{ labelText: 'Use for previews' }}
        />
      </div>

      {/* Installs / labels content */}
      <div className="flex flex-col gap-2 p-4">
        {group.selection_mode === 'labels' ? (
          <LabelSelectorEditor
            groupId={group.id}
            labelSelector={group.label_selector}
            availableInstalls={availableInstalls}
            disabled={disabled}
            onUpdate={(ls) => onUpdate({ label_selector: ls })}
          />
        ) : (
          <>
            {installs.length > 0 ? (
              <div className="flex flex-col gap-1.5">
                {installs.map((install) => (
                  <div
                    key={install.id}
                    className="flex items-center justify-between gap-2 px-2 py-2 rounded-md bg-cool-grey-50 dark:bg-dark-grey-900"
                  >
                    <Text variant="body" className="truncate">
                      {install.name || install.id}
                    </Text>
                    <Button
                      variant="ghost"
                      size="xs"
                      onClick={() => onRemoveInstall(install.id)}
                      disabled={disabled}
                      title={`Remove ${install.name || install.id}`}
                      className="!p-1 shrink-0"
                    >
                      <Icon variant="XIcon" size={14} />
                    </Button>
                  </div>
                ))}
              </div>
            ) : (
              <div className="px-3 py-3 rounded-md border border-dashed border-cool-grey-300 dark:border-dark-grey-600 text-center">
                <Text variant="subtext" theme="neutral">
                  Use Add install to assign installs to this group
                </Text>
              </div>
            )}

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
  )
}

const LabelSelectorEditor = ({
  groupId,
  labelSelector,
  availableInstalls,
  disabled,
  onUpdate,
}: {
  groupId: string
  labelSelector?: ILabelSelector | null
  availableInstalls: TInstall[]
  disabled?: boolean
  onUpdate: (ls: ILabelSelector) => void
}) => {
  const [newKey, setNewKey] = useState('')
  const [newValue, setNewValue] = useState('')

  const labels = labelSelector?.match_labels ?? {}
  const entries = Object.entries(labels)

  const suggestedLabels = useMemo(() => {
    const seen = new Set<string>()
    const result: Array<{ key: string; value: string }> = []
    for (const install of availableInstalls) {
      for (const [k, v] of Object.entries(install.labels ?? {})) {
        const token = `${k}=${v}`
        if (!seen.has(token)) {
          seen.add(token)
          result.push({ key: k, value: v })
        }
      }
    }
    return result
  }, [availableInstalls])

  const addLabel = () => {
    const key = newKey.trim()
    const value = newValue.trim()
    if (!key) return
    onUpdate({ match_labels: { ...labels, [key]: value } })
    setNewKey('')
    setNewValue('')
  }

  const toggleSuggestion = (key: string, value: string) => {
    if (labels[key] === value) {
      const next = { ...labels }
      delete next[key]
      onUpdate({ match_labels: next })
    } else {
      onUpdate({ match_labels: { ...labels, [key]: value } })
    }
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

      {suggestedLabels.length > 0 && (
        <div className="flex flex-col gap-1.5">
          <Text variant="subtext" theme="neutral">
            Labels from your installs
          </Text>
          <div className="flex flex-wrap gap-1.5">
            {suggestedLabels.map(({ key, value }) => {
              const isActive = labels[key] === value
              return (
                <button
                  key={`${key}=${value}`}
                  type="button"
                  onClick={() => toggleSuggestion(key, value)}
                  disabled={disabled}
                  className={cn(
                    'inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-mono border transition-colors',
                    isActive
                      ? 'bg-primary-100 border-primary-400 text-primary-800 dark:bg-primary-900/30 dark:border-primary-500 dark:text-primary-200'
                      : 'bg-cool-grey-100 border-cool-grey-300 text-cool-grey-700 hover:border-cool-grey-400 dark:bg-dark-grey-600 dark:border-dark-grey-400 dark:text-cool-grey-200 dark:hover:border-cool-grey-300'
                  )}
                >
                  {isActive && <Icon variant="CheckIcon" size={11} />}
                  {key}={value}
                </button>
              )
            })}
          </div>
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

      {entries.length === 0 && suggestedLabels.length === 0 && (
        <div className="px-3 py-3 rounded-md border border-dashed border-cool-grey-300 dark:border-dark-grey-600 text-center">
          <Text variant="subtext" theme="neutral">
            Add label selectors to match installs dynamically
          </Text>
        </div>
      )}
    </div>
  )
}
