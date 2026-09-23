import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import type { TAppBranchConfig, TAppBranchInstallGroup } from '@/types'

export type TGroupStepSelection = TAppBranchInstallGroup | null

/** Concrete match_labels: entries where the value is not a wildcard. */
export const concreteMatchLabels = (
  group: TAppBranchInstallGroup
): Record<string, string> =>
  Object.fromEntries(
    Object.entries(group.label_selector?.match_labels ?? {}).filter(
      ([, v]) => v !== '*'
    )
  )

/** True when any concrete match_label conflicts with the provided install labels. */
export const hasLabelConflict = (
  group: TAppBranchInstallGroup,
  installLabels: Record<string, string>
): boolean => {
  const concrete = concreteMatchLabels(group)
  return Object.entries(concrete).some(
    ([k, v]) => installLabels[k] !== undefined && installLabels[k] !== v
  )
}

/** Merge group's concrete match_labels on top of existing install labels. */
export const mergeGroupLabels = (
  existingLabels: Record<string, string>,
  group: TAppBranchInstallGroup | null
): Record<string, string> => {
  if (!group) return existingLabels
  return { ...existingLabels, ...concreteMatchLabels(group) }
}

type GroupKind = 'label' | 'default'

const resolveGroupKind = (group: TAppBranchInstallGroup): GroupKind => {
  const labels = group.label_selector?.match_labels ?? {}
  if (Object.keys(labels).length > 0) return 'label'
  return 'default'
}

const hasWildcard = (group: TAppBranchInstallGroup): boolean => {
  const labels = group.label_selector?.match_labels ?? {}
  return Object.values(labels).some((v) => v === '*')
}

interface IGroupStep {
  config: TAppBranchConfig
  installLabels: Record<string, string>
  selected: TGroupStepSelection
  onSelect: (group: TGroupStepSelection) => void
}

interface IGroupRowProps {
  group: TAppBranchInstallGroup
  installLabels: Record<string, string>
  isSelected: boolean
  onSelect: (group: TAppBranchInstallGroup) => void
}

const GroupRow = ({
  group,
  installLabels,
  isSelected,
  onSelect,
}: IGroupRowProps) => {
  const kind = resolveGroupKind(group)
  const labelEntries = Object.entries(group.label_selector?.match_labels ?? {})
  const concrete = concreteMatchLabels(group)
  const conflict = hasLabelConflict(group, installLabels)
  const wildcard = hasWildcard(group)

  const isDisabled = kind !== 'label' || wildcard || conflict

  let disabledReason: string | null = null
  if (kind === 'default') {
    disabledReason =
      'This group includes all remaining installs automatically — it already includes this install.'
  } else if (wildcard) {
    disabledReason =
      'This group uses a wildcard selector and cannot be joined by applying labels.'
  } else if (conflict) {
    disabledReason =
      'This group conflicts with labels already set on the install. Update those labels before selecting it.'
  }

  return (
    <div className="flex flex-col gap-2">
      <button
        type="button"
        disabled={isDisabled}
        onClick={() => !isDisabled && onSelect(group)}
        className={`flex items-center justify-between gap-3 px-3 py-2.5 rounded-md text-left w-full ${
          isDisabled
            ? 'opacity-60 cursor-not-allowed bg-cool-grey-50 dark:bg-dark-grey-700'
            : isSelected
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
          {kind === 'default' && (
            <Text variant="subtext" theme="neutral">
              All remaining installs
            </Text>
          )}
        </div>

        <div className="flex shrink-0 items-center gap-2">
          {isSelected && (
            <span className="flex items-center gap-1 text-xs text-primary-700 dark:text-primary-300">
              <Icon variant="CheckIcon" size={14} />
              Selected
            </span>
          )}
        </div>
      </button>

      {disabledReason && (
        <Text variant="subtext" theme="neutral" className="pl-3">
          {disabledReason}
        </Text>
      )}

      {!isDisabled && !isSelected && Object.keys(concrete).length > 0 && (
        <Text variant="subtext" theme="neutral" className="pl-3">
          Joining this group will apply:{' '}
          {Object.entries(concrete)
            .map(([k, v]) => `${k}=${v}`)
            .join(', ')}
        </Text>
      )}
    </div>
  )
}

export const GroupStep = ({
  config,
  installLabels,
  selected,
  onSelect,
}: IGroupStep) => {
  const groups = config.install_groups ?? []

  const labelGroups = groups.filter(
    (g) => resolveGroupKind(g) === 'label' && !hasWildcard(g)
  )
  const otherGroups = groups.filter(
    (g) => resolveGroupKind(g) !== 'label' || hasWildcard(g)
  )

  if (groups.length === 0) {
    return (
      <div className="flex flex-col gap-3">
        <Text variant="subtext" theme="neutral">
          This branch has no install groups. The install will be created on this
          branch without a group assignment.
        </Text>
      </div>
    )
  }

  if (labelGroups.length === 0) {
    return (
      <div className="flex flex-col gap-3">
        <Text variant="subtext" theme="neutral">
          This branch has no label-selector groups. The install will be created
          on this branch without a group assignment.
        </Text>
        {otherGroups.map((group, idx) => (
          <GroupRow
            key={group.id ?? idx}
            group={group}
            installLabels={installLabels}
            isSelected={false}
            onSelect={() => {}}
          />
        ))}
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <Text variant="subtext" theme="neutral">
        Optionally join an install group. Joining a label-selector group will
        apply its labels to this install and include it in branch deployments.
        You can skip this and join a group later.
      </Text>

      <div className="flex flex-col gap-3">
        {/* Selectable label groups */}
        {labelGroups.map((group, idx) => (
          <GroupRow
            key={group.id || idx}
            group={group}
            installLabels={installLabels}
            isSelected={selected?.id === group.id}
            onSelect={onSelect}
          />
        ))}

        {/* Disabled non-label groups */}
        {otherGroups.length > 0 && (
          <div className="flex flex-col gap-3 border-t pt-3">
            <Text variant="subtext" theme="neutral" weight="strong">
              Other groups (not selectable)
            </Text>
            {otherGroups.map((group, idx) => (
              <GroupRow
                key={group.id || idx}
                group={group}
                installLabels={installLabels}
                isSelected={false}
                onSelect={() => {}}
              />
            ))}
          </div>
        )}

        {/* "No group" option */}
        <button
          type="button"
          onClick={() => onSelect(null)}
          className={`flex items-center gap-2 px-3 py-2.5 rounded-md text-left w-full ${
            selected === null
              ? 'bg-primary-50 dark:bg-primary-900/30 ring-1 ring-primary-500'
              : 'bg-cool-grey-50 dark:bg-dark-grey-700 hover:bg-cool-grey-100 dark:hover:bg-dark-grey-600'
          }`}
        >
          <div className="flex items-center gap-2 flex-1 min-w-0">
            <Icon
              variant="MinusCircleIcon"
              size={14}
              className="shrink-0 text-cool-grey-400"
            />
            <div className="flex flex-col min-w-0">
              <Text variant="body" weight="strong">
                No group
              </Text>
              <Text variant="subtext" theme="neutral">
                Create the install on this branch without joining a group now.
              </Text>
            </div>
          </div>
          {selected === null && (
            <span className="flex shrink-0 items-center gap-1 text-xs text-primary-700 dark:text-primary-300">
              <Icon variant="CheckIcon" size={14} />
              Selected
            </span>
          )}
        </button>
      </div>
    </div>
  )
}
