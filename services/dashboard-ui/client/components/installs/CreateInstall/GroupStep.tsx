import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import type { TAppBranchConfig, TAppBranchInstallGroup } from '@/types'

export type TGroupStepSelection =
  | { mode: 'none'; group: TAppBranchInstallGroup | null }
  | {
      mode: 'default' | 'labels' | 'explicit'
      group: TAppBranchInstallGroup
    }

export const appBranchGroupFromSelection = (
  selection: TGroupStepSelection
): string | undefined =>
  selection.mode === 'explicit' ? selection.group.name || undefined : undefined

export const labelsFromGroupSelection = (
  selection: TGroupStepSelection
): Record<string, string> =>
  selection.mode === 'labels'
    ? (selection.group.label_selector?.match_labels ?? {})
    : {}

interface IGroupStep {
  config: TAppBranchConfig
  selected: TGroupStepSelection
  onSelect: (group: TGroupStepSelection) => void
}

interface IGroupRowProps {
  group: TAppBranchInstallGroup
  selected: boolean
  selectedMode?: TGroupStepSelection['mode']
  onSelectGroup: (group: TAppBranchInstallGroup) => void
  onSelectMode: (selection: TGroupStepSelection) => void
}

const GroupRow = ({
  group,
  selected,
  selectedMode,
  onSelectGroup,
  onSelectMode,
}: IGroupRowProps) => {
  const labelEntries = Object.entries(group.label_selector?.match_labels ?? {})

  return (
    <div className="flex flex-col gap-3">
      <button
        type="button"
        onClick={() => onSelectGroup(group)}
        className={`flex items-center justify-between gap-3 px-3 py-2.5 rounded-md text-left w-full ${
          selected
            ? 'bg-primary-50 dark:bg-primary-900/30 ring-1 ring-primary-500'
            : 'bg-cool-grey-50 dark:bg-dark-grey-700 hover:bg-cool-grey-100 dark:hover:bg-dark-grey-600'
        }`}
      >
        <div className="flex items-center gap-2 flex-wrap min-w-0">
          <Text variant="body" weight="strong">
            {group.name}
          </Text>
          {labelEntries.map(([k, v]) => (
            <LabelBadge key={k} labelKey={k} labelValue={v} size="sm" />
          ))}
          {group.default && (
            <Text variant="subtext" theme="neutral">
              Default group
            </Text>
          )}
        </div>
        {selected && (
          <span className="flex shrink-0 items-center gap-1 text-xs text-primary-700 dark:text-primary-300">
            <Icon variant="CheckIcon" size={14} />
            Selected
          </span>
        )}
      </button>

      {selected && !group.default && (
        <div className="flex flex-col gap-2 pl-3">
          <Text variant="subtext" theme="neutral">
            Choose how this install joins {group.name}.
          </Text>
          <div className="flex gap-2">
            <Button
              variant={selectedMode === 'explicit' ? 'primary' : 'secondary'}
              size="sm"
              onClick={() => onSelectMode({ mode: 'explicit', group })}
            >
              Pin to group
            </Button>
            {labelEntries.length > 0 && (
              <Button
                variant={selectedMode === 'labels' ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => onSelectMode({ mode: 'labels', group })}
              >
                Add labels
              </Button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

export const GroupStep = ({ config, selected, onSelect }: IGroupStep) => {
  const groups = config.install_groups ?? []

  if (groups.length === 0) {
    return (
      <div className="flex flex-col gap-3">
        <Text variant="subtext" theme="neutral">
          This branch has no deployment groups. Add a default group before
          creating an install on this branch.
        </Text>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <Text variant="subtext" theme="neutral">
        Select the deployment group this install should join.
      </Text>

      <div className="flex flex-col gap-3">
        {groups.map((group, idx) => {
          const isSelected = selected.group?.name === group.name
          return (
            <GroupRow
              key={group.id || idx}
              group={group}
              selected={isSelected}
              selectedMode={isSelected ? selected.mode : undefined}
              onSelectGroup={(nextGroup) =>
                onSelect(
                  nextGroup.default
                    ? { mode: 'default', group: nextGroup }
                    : { mode: 'none', group: nextGroup }
                )
              }
              onSelectMode={onSelect}
            />
          )
        })}
      </div>
    </div>
  )
}
