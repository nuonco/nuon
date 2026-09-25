import { useEffect, useMemo, useRef, useState } from 'react'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Dropdown } from '@/components/common/Dropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import { Input } from '@/components/common/form/Input'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import type { TInstall } from '@/types'
import { cn } from '@/utils/classnames'
import { matchesSelector } from '@/components/match/matches'
import { parseLabelsQuery } from '@/components/match/parse'
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
  availableInstalls: TInstall[]
  resolvedInstalls?: TInstall[]
  labelColors?: Record<string, string>
  disabled?: boolean
  autoFocusName?: boolean
  nameError?: string
  contentError?: string
  deleteDisabledReason?: string
  onUpdate: (updates: Partial<IInstallGroup>) => void
  onMoveUp: () => void
  onMoveDown: () => void
  onDelete: () => void
}

export const GroupEditor = ({
  group,
  index,
  totalGroups,
  availableInstalls,
  resolvedInstalls = availableInstalls,
  labelColors,
  disabled,
  autoFocusName,
  nameError,
  contentError,
  deleteDisabledReason,
  onUpdate,
  onMoveUp,
  onMoveDown,
  onDelete,
}: IGroupEditor) => {
  const nameRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!autoFocusName || disabled) return
    nameRef.current?.focus({ preventScroll: true })
    nameRef.current?.select()
  }, [autoFocusName, disabled])

  return (
    <Card className="!p-0 !gap-0 overflow-hidden">
      <div className="flex items-center gap-2 px-5 py-3.5 bg-cool-grey-50 dark:bg-dark-grey-800">
        <div className="flex-1 min-w-0">
          <Input
            ref={nameRef}
            id={`group-name-${group.id}`}
            type="text"
            aria-label={`Group ${index + 1} name`}
            value={group.name}
            onChange={(e) => onUpdate({ name: e.target.value })}
            placeholder={`Group ${index + 1}`}
            disabled={disabled}
            className="!font-bold"
            error={!!nameError}
            errorMessage={nameError}
          />
        </div>

        <div className="flex items-center gap-1">
          <ToggleButton<InstallSelectionMode>
            options={[
              { value: 'labels', label: 'Labels' },
              { value: 'pinned', label: 'Pinned only' },
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
              <Button
                isMenuButton
                variant="danger"
                onClick={onDelete}
                disabled={!!deleteDisabledReason}
                tooltipProps={
                  deleteDisabledReason
                    ? { tipContent: deleteDisabledReason }
                    : undefined
                }
              >
                Delete group
                <Icon variant="TrashIcon" />
              </Button>
            </Menu>
          </Dropdown>
        </div>
      </div>

      <div className="flex flex-col gap-4 p-5">
        {group.selection_mode === 'labels' ? (
          <LabelSelectorEditor
            groupId={group.id}
            labelSelector={group.label_selector}
            availableInstalls={availableInstalls}
            resolvedInstalls={resolvedInstalls}
            labelColors={labelColors}
            disabled={disabled}
            onUpdate={(ls) => onUpdate({ label_selector: ls })}
          />
        ) : group.is_default ? (
          <DefaultGroupSummary installCount={resolvedInstalls.length} />
        ) : (
          <PinnedGroupSummary installCount={resolvedInstalls.length} />
        )}

        {group.selection_mode === 'labels' && contentError && (
          <Text variant="subtext" theme="error">
            {contentError}
          </Text>
        )}

        {group.selection_mode === 'labels' && group.is_default && (
          <DefaultGroupSummary installCount={resolvedInstalls.length} />
        )}

        <div className="border-t pt-3">
          <CheckboxInput
            id={`group-default-${group.id}`}
            checked={group.is_default}
            disabled={disabled}
            onChange={(e) => onUpdate({ is_default: e.target.checked })}
            labelProps={{
              labelText: (
                <span className="flex flex-col gap-1">
                  <Text
                    variant="subtext"
                    weight="strong"
                    className="!leading-none"
                  >
                    Default group
                  </Text>
                  <Text
                    variant="subtext"
                    theme="neutral"
                    className="!leading-none"
                  >
                    Installs that match no other group deploy here.
                  </Text>
                </span>
              ),
              className: '!p-1 !gap-1.5 items-start',
              labelTextProps: { as: 'div' },
            }}
          />
        </div>
      </div>
    </Card>
  )
}

const DefaultGroupSummary = ({ installCount }: { installCount: number }) => (
  <Text variant="subtext" theme="neutral">
    {installCount === 0
      ? 'No installs yet — installs join this group as they are created.'
      : `${installCount} install${installCount === 1 ? '' : 's'} in this group today.`}
  </Text>
)

const PinnedGroupSummary = ({ installCount }: { installCount: number }) => (
  <div className="flex flex-col gap-1">
    <Text variant="subtext" theme="neutral">
      Only installs manually pinned to this group are included.
    </Text>
    <Text variant="subtext" theme="neutral">
      {installCount === 0
        ? 'No installs are pinned to this group.'
        : `${installCount} install${installCount === 1 ? '' : 's'} pinned.`}
    </Text>
  </div>
)

const LabelSelectorEditor = ({
  groupId,
  labelSelector,
  availableInstalls,
  resolvedInstalls,
  labelColors,
  disabled,
  onUpdate,
}: {
  groupId: string
  labelSelector?: ILabelSelector | null
  availableInstalls: TInstall[]
  resolvedInstalls: TInstall[]
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

  const matchedInstalls = hasSelector
    ? resolvedInstalls.filter((install) =>
        matchesSelector(install.labels, labelSelector)
      )
    : []

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
