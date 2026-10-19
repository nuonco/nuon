import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAppBranchInstallGroup, TInstall } from '@/types'
import { labelChangesForGroupMove } from './label-changes'

const selectorEntries = (group?: TAppBranchInstallGroup) =>
  Object.entries(group?.label_selector?.match_labels ?? {})

const SelectorLabels = ({ group }: { group?: TAppBranchInstallGroup }) => {
  const entries = selectorEntries(group)
  if (entries.length === 0) return null
  return (
    <div className="flex flex-wrap gap-1">
      {entries.map(([key, value]) => (
        <LabelBadge key={key} labelKey={key} labelValue={value} size="sm" />
      ))}
    </div>
  )
}

interface IChangeInstallGroupModal extends Omit<IModal, 'onSubmit'> {
  install: TInstall
  branchName?: string
  currentGroup?: string
  groups: TAppBranchInstallGroup[]
  targetGroup: string
  isPending: boolean
  isLoading: boolean
  loadFailed: boolean
  onSelectGroup: (group: string) => void
  onConfirm: () => void
}

export const ChangeInstallGroupModal = ({
  install,
  branchName,
  currentGroup,
  groups,
  targetGroup,
  isPending,
  isLoading,
  loadFailed,
  onSelectGroup,
  onConfirm,
  ...props
}: IChangeInstallGroupModal) => {
  const destinations = groups.filter(
    (group) => group.name && group.name !== currentGroup
  )
  const current = groups.find((group) => group.name === currentGroup)
  const selected = groups.find((group) => group.name === targetGroup)
  const labelChanges = labelChangesForGroupMove(install.labels, selected)
  const canMove = !!targetGroup && targetGroup !== currentGroup

  const confirmLabel = isPending ? (
    <span className="flex items-center gap-2">
      <Icon variant="Loading" />
      Moving install
    </span>
  ) : (
    'Move group'
  )

  return (
    <Modal
      size="lg"
      heading={
        <Text flex className="gap-3" variant="h3" weight="strong">
          <Icon variant="UsersThreeIcon" size="24" />
          Change install group
        </Text>
      }
      primaryActionTrigger={{
        children: confirmLabel,
        disabled: !canMove || isPending || isLoading || loadFailed,
        tooltipProps:
          !isPending && !isLoading && !loadFailed && !canMove
            ? { tipContent: 'Choose an install group' }
            : undefined,
        onClick: onConfirm,
        variant: 'primary',
      }}
      secondaryActionTrigger={{
        children: 'Cancel',
        disabled: isPending,
        onClick: () => props.onClose?.(),
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        <Banner theme="info">
          {labelChanges.length > 0 ? (
            <>
              Moving to <strong>{selected?.name}</strong> updates the labels
              that group matches. Other labels stay, and the install is
              reconciled immediately.
            </>
          ) : (
            <>
              This pins {install.name} to one group on{' '}
              {branchName ? <strong>{branchName}</strong> : 'this branch'}.
              Labels stay unchanged, and the install is reconciled immediately.
            </>
          )}
        </Banner>

        <div className="flex flex-col gap-2">
          <Text variant="base" weight="strong">
            Current group
          </Text>
          <div className="flex flex-col gap-2 px-3 py-2.5 rounded-md bg-cool-grey-50 dark:bg-dark-grey-700">
            <div className="flex items-center gap-2">
              <Icon
                variant="UsersThreeIcon"
                size={14}
                className="text-cool-grey-400"
              />
              <Text variant="body" family="mono">
                {currentGroup || 'None'}
              </Text>
            </div>
            {selectorEntries(current).length > 0 && (
              <div className="flex flex-col gap-1.5">
                <Text variant="subtext" theme="neutral">
                  Group selector
                </Text>
                <SelectorLabels group={current} />
              </div>
            )}
            {Object.keys(install.labels ?? {}).length > 0 && (
              <div className="flex flex-col gap-1.5">
                <Text variant="subtext" theme="neutral">
                  Install labels
                </Text>
                <div className="flex flex-wrap gap-1">
                  {Object.entries(install.labels ?? {}).map(([key, value]) => (
                    <LabelBadge
                      key={key}
                      labelKey={key}
                      labelValue={value}
                      size="sm"
                    />
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>

        {isLoading && (
          <div className="flex items-center gap-2">
            <Icon variant="Loading" size={14} />
            <Text variant="subtext" theme="neutral">
              Loading install groups
            </Text>
          </div>
        )}

        {loadFailed && (
          <Banner theme="error">
            Install groups could not be loaded. Try again in a moment.
          </Banner>
        )}

        {!isLoading && !loadFailed && destinations.length > 0 && (
          <div className="flex flex-col gap-2">
            <Text variant="base" weight="strong">
              Move to
            </Text>
            <div className="flex flex-col gap-2">
              {destinations.map((group) => {
                const isSelected = targetGroup === group.name
                return (
                  <button
                    key={group.id ?? group.name}
                    type="button"
                    disabled={isPending}
                    onClick={() => group.name && onSelectGroup(group.name)}
                    className={`flex items-center justify-between gap-3 px-3 py-2.5 rounded-md text-left w-full cursor-pointer focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-500 ${
                      isSelected
                        ? 'bg-primary-50 dark:bg-primary-900/30 ring-1 ring-primary-500'
                        : 'bg-cool-grey-50 dark:bg-dark-grey-700 hover:bg-cool-grey-100 dark:hover:bg-dark-grey-600'
                    }`}
                  >
                    <div className="flex min-w-0 flex-col gap-1.5">
                      <div className="flex items-center gap-2">
                        <Text
                          variant="body"
                          weight="strong"
                          className="truncate"
                        >
                          {group.name}
                        </Text>
                        {group.default && (
                          <Text variant="subtext" theme="neutral">
                            Default group
                          </Text>
                        )}
                      </div>
                      <SelectorLabels group={group} />
                    </div>
                    {isSelected && (
                      <span className="flex shrink-0 items-center gap-1 text-xs text-primary-700 dark:text-primary-300">
                        <Icon variant="CheckIcon" size={14} />
                        Selected
                      </span>
                    )}
                  </button>
                )
              })}
            </div>
          </div>
        )}

        {labelChanges.length > 0 && (
          <div className="flex flex-col gap-2">
            <Text variant="base" weight="strong">
              Label changes
            </Text>
            <div className="flex flex-col gap-2">
              {labelChanges.map((change) => (
                <div key={change.key} className="flex items-center gap-2">
                  {change.from ? (
                    <LabelBadge
                      labelKey={change.key}
                      labelValue={change.from}
                      size="sm"
                    />
                  ) : (
                    <Text variant="subtext" theme="neutral">
                      Add {change.key}
                    </Text>
                  )}
                  <Icon
                    variant="ArrowRightIcon"
                    size={14}
                    className="text-cool-grey-400"
                  />
                  <LabelBadge
                    labelKey={change.key}
                    labelValue={change.to}
                    size="sm"
                  />
                </div>
              ))}
            </div>
          </div>
        )}

        {!isLoading && !loadFailed && destinations.length === 0 && (
          <Banner theme="warn">
            {groups.length === 0
              ? 'This branch has no install groups. Add a group before moving this install.'
              : 'This branch has no other install groups.'}
          </Banner>
        )}
      </div>
    </Modal>
  )
}
