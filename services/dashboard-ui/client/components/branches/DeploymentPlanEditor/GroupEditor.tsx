import { useMemo, useState } from 'react'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Dropdown } from '@/components/common/Dropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import { Tooltip } from '@/components/common/Tooltip'
import { Input } from '@/components/common/form/Input'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import type { TInstall } from '@/types'
import { cn } from '@/utils/classnames'
import { matchesSelector } from '@/components/match/matches'
import { parseLabelsQuery } from '@/components/match/parse'
import { AddInstallPicker } from './AddInstallPicker'
import { InstallRow } from './InstallRow'
import type {
  IInstallGroup,
  ILabelSelector,
  InstallSelectionMode,
} from './types'

interface IGroupEditor {
  group: IInstallGroup
  index: number
  totalGroups: number
  unassignedInstalls: TInstall[]
  availableInstalls: TInstall[]
  labelColors?: Record<string, string>
  disabled?: boolean
  nameError?: string
  contentError?: string
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
  labelColors,
  disabled,
  nameError,
  contentError,
  onUpdate,
  onAddInstalls,
  onRemoveInstall,
  onMoveUp,
  onMoveDown,
  onDelete,
}: IGroupEditor) => {
  const installs = useMemo(() => {
    const byId = new Map(availableInstalls.map((i) => [i.id, i]))
    return group.install_ids
      .map((id) => byId.get(id))
      .filter((i): i is TInstall => !!i)
  }, [group.install_ids, availableInstalls])

  return (
    <Card className="!p-0 !gap-0 overflow-hidden">
      <div className="flex items-center gap-2 px-4 py-2.5 bg-cool-grey-50 dark:bg-dark-grey-800">
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

        <div className="flex items-center gap-1">
          <ToggleButton<InstallSelectionMode>
            options={[
              { value: 'manual', label: 'Manual' },
              { value: 'labels', label: 'Labels' },
            ]}
            value={group.selection_mode}
            onChange={(mode) => onUpdate({ selection_mode: mode })}
            size="sm"
            className={cn(disabled && 'pointer-events-none opacity-50')}
          />

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
              <Button isMenuButton variant="danger" onClick={onDelete}>
                Delete group
                <Icon variant="TrashIcon" />
              </Button>
            </Menu>
          </Dropdown>
        </div>
      </div>

      <div className="flex flex-col gap-3 p-4">
        {group.selection_mode === 'labels' ? (
          <LabelSelectorEditor
            groupId={group.id}
            labelSelector={group.label_selector}
            availableInstalls={availableInstalls}
            labelColors={labelColors}
            disabled={disabled}
            onUpdate={(ls) => onUpdate({ label_selector: ls })}
          />
        ) : (
          <>
            {installs.length > 0 ? (
              <div className="flex flex-col gap-1.5">
                {installs.map((install) => (
                  <InstallRow
                    key={install.id}
                    install={install}
                    labelColors={labelColors}
                    onRemove={() => onRemoveInstall(install.id)}
                    disabled={disabled}
                  />
                ))}
              </div>
            ) : (
              <EmptyState
                variant="table"
                size="sm"
                emptyTitle="No installs"
                emptyMessage="Add an install below, or delete this group — a group can't be saved empty."
                action={
                  <Button
                    variant="danger"
                    onClick={onDelete}
                    disabled={disabled}
                  >
                    <Icon variant="TrashIcon" size={16} />
                    Delete group
                  </Button>
                }
              />
            )}

            <Tooltip
              isOpen={!!contentError}
              disableHover
              position="right"
              tipContent={<Text variant="subtext">{contentError}</Text>}
            >
              <AddInstallPicker
                groupId={group.id}
                unassignedInstalls={unassignedInstalls}
                disabled={disabled}
                onAdd={onAddInstalls}
              />
            </Tooltip>
          </>
        )}

        {group.selection_mode === 'labels' && contentError && (
          <Text variant="subtext" theme="error">
            {contentError}
          </Text>
        )}
      </div>
    </Card>
  )
}

const LabelSelectorEditor = ({
  groupId,
  labelSelector,
  availableInstalls,
  labelColors,
  disabled,
  onUpdate,
}: {
  groupId: string
  labelSelector?: ILabelSelector | null
  availableInstalls: TInstall[]
  labelColors?: Record<string, string>
  disabled?: boolean
  onUpdate: (ls: ILabelSelector) => void
}) => {
  const [draft, setDraft] = useState('')

  const labels = labelSelector?.match_labels ?? {}
  const entries = Object.entries(labels)
  const hasSelector = entries.length > 0

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

  const matchedInstalls = useMemo(
    () =>
      hasSelector
        ? availableInstalls.filter((i) =>
            matchesSelector(i.labels, labelSelector)
          )
        : [],
    [hasSelector, availableInstalls, labelSelector]
  )

  const commitDraft = () => {
    const parsed = parseLabelsQuery(draft)
    setDraft('')
    if (Object.keys(parsed).length === 0) return
    onUpdate({ match_labels: { ...labels, ...parsed } })
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
        Installs matching all labels are included at deploy time.
      </Text>

      {hasSelector && (
        <div className="flex flex-wrap gap-1.5">
          {entries.map(([key, value]) => (
            <LabelBadge
              key={key}
              labelKey={key}
              labelValue={value}
              size="sm"
              disabled={disabled}
              customColor={labelColors?.[key]}
              onRemove={() => removeLabel(key)}
              removeAriaLabel={`Remove ${key}=${value}`}
            />
          ))}
        </div>
      )}

      <Input
        id={`label-input-${groupId}`}
        type="text"
        size="sm"
        placeholder="env=prod — press Enter to add"
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            e.preventDefault()
            commitDraft()
          }
        }}
        onBlur={commitDraft}
        disabled={disabled}
      />

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
                  className="disabled:opacity-50"
                >
                  <LabelBadge
                    labelKey={key}
                    labelValue={value}
                    size="sm"
                    keyTheme={isActive ? 'brand' : 'neutral'}
                    theme={isActive ? 'brand' : 'default'}
                  />
                </button>
              )
            })}
          </div>
        </div>
      )}

      {hasSelector ? (
        matchedInstalls.length > 0 ? (
          <div className="flex flex-col gap-1.5">
            <Text variant="subtext" theme="neutral">
              {matchedInstalls.length}{' '}
              {matchedInstalls.length === 1
                ? 'install matches'
                : 'installs match'}
            </Text>
            {matchedInstalls.map((install) => (
              <InstallRow
                key={install.id}
                install={install}
                labelColors={labelColors}
              />
            ))}
          </div>
        ) : (
          <EmptyState
            variant="search"
            size="sm"
            emptyTitle="No matches"
            emptyMessage="No installs match this selector."
          />
        )
      ) : (
        <EmptyState
          variant="search"
          size="sm"
          emptyTitle="No labels yet"
          emptyMessage="Add a label above or pick one from your installs."
        />
      )}
    </div>
  )
}
