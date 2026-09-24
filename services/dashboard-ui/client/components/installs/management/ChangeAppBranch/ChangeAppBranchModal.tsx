import { Banner } from '@/components/common/Banner'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Select } from '@/components/common/form/Select'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAppBranch, TInstall } from '@/types'

export type TBranchGroupAssignmentMode = 'default' | 'labels' | 'explicit'

interface IChangeAppBranchModal extends Omit<IModal, 'onSubmit'> {
  install: TInstall
  targetBranch: TAppBranch | null
  targetGroup: string
  assignmentMode: TBranchGroupAssignmentMode | null
  branches: TAppBranch[]
  isPending: boolean
  onSelectBranch: (branch: TAppBranch) => void
  onSelectGroup: (group: string) => void
  onSelectAssignmentMode: (mode: TBranchGroupAssignmentMode) => void
  onConfirm: () => void
}

export const ChangeAppBranchModal = ({
  install,
  targetBranch,
  targetGroup,
  assignmentMode,
  branches,
  isPending,
  onSelectBranch,
  onSelectGroup,
  onSelectAssignmentMode,
  onConfirm,
  ...props
}: IChangeAppBranchModal) => {
  const targetGroups = targetBranch?.configs?.at(0)?.install_groups ?? []
  const selectedGroup = targetGroups.find((group) => group.name === targetGroup)
  const selectedLabels = Object.entries(
    selectedGroup?.label_selector?.match_labels ?? {}
  )
  const assignmentIncomplete =
    !selectedGroup ||
    (!selectedGroup.default &&
      assignmentMode !== 'labels' &&
      assignmentMode !== 'explicit')

  const confirmLabel = isPending ? (
    <span className="flex items-center gap-2">
      <Icon variant="Loading" />
      Moving install
    </span>
  ) : (
    'Move and reconcile'
  )

  return (
    <Modal
      size="lg"
      heading={
        <Text flex className="gap-3" variant="h3" weight="strong">
          <Icon variant="GitBranchIcon" size="24" />
          Change app branch
        </Text>
      }
      primaryActionTrigger={{
        children: confirmLabel,
        disabled:
          !targetBranch ||
          isPending ||
          targetGroups.length === 0 ||
          assignmentIncomplete,
        tooltipProps:
          targetBranch && !isPending && assignmentIncomplete
            ? { tipContent: 'Choose how to assign this install to a group' }
            : targetBranch && !isPending && targetGroups.length === 0
              ? { tipContent: 'This branch has no deployment groups' }
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
        <Banner theme="warn">
          <strong>Warning:</strong> Moving this install to a different branch
          will change its app config and affect all future deployments. The
          install will be reconciled against the new branch immediately.
        </Banner>

        <div className="flex flex-col gap-2">
          <Text variant="base" weight="strong">
            Current branch
          </Text>
          <div className="flex items-center gap-2 px-3 py-2.5 rounded-md bg-cool-grey-50 dark:bg-dark-grey-700">
            <Icon
              variant="GitBranchIcon"
              size={14}
              className="text-cool-grey-400"
            />
            <Text variant="body">
              {install.app_branch?.name ?? install.app_branch_id ?? 'None'}
            </Text>
          </div>
        </div>

        <div className="flex flex-col gap-2">
          <Text variant="base" weight="strong">
            Select destination branch
          </Text>
          <div className="flex flex-col gap-2">
            {branches
              .filter((b) => b.id !== install.app_branch_id)
              .map((branch) => {
                const isSelected = targetBranch?.id === branch.id
                return (
                  <button
                    key={branch.id}
                    type="button"
                    disabled={isPending}
                    onClick={() => onSelectBranch(branch)}
                    className={`flex items-center justify-between gap-3 px-3 py-2.5 rounded-md text-left w-full ${
                      isSelected
                        ? 'bg-primary-50 dark:bg-primary-900/30 ring-1 ring-primary-500'
                        : 'bg-cool-grey-50 dark:bg-dark-grey-700 hover:bg-cool-grey-100 dark:hover:bg-dark-grey-600'
                    }`}
                  >
                    <div className="flex items-center gap-2 min-w-0">
                      <Icon
                        variant="GitBranchIcon"
                        size={14}
                        className="shrink-0 text-cool-grey-400"
                      />
                      <Text variant="body" weight="strong" className="truncate">
                        {branch.name}
                      </Text>
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
            {branches.filter((b) => b.id !== install.app_branch_id).length ===
              0 && (
              <Text variant="subtext" theme="neutral">
                No other branches are configured for this app.
              </Text>
            )}
          </div>
        </div>

        {targetBranch && targetGroups.length > 0 && (
          <div className="flex flex-col gap-3">
            <Text variant="base" weight="strong">
              Group assignment
            </Text>
            <Select
              id="app-branch-group"
              value={targetGroup}
              onChange={onSelectGroup}
              labelProps={{ labelText: 'Deployment group' }}
              options={[
                {
                  value: '',
                  label: 'Select a group',
                },
                ...targetGroups
                  .filter((group) => !!group.name)
                  .map((group) => ({
                    value: group.name!,
                    label: group.name!,
                    description: group.default ? 'Default group' : undefined,
                  })),
              ]}
            />

            {selectedGroup?.default && (
              <Text variant="subtext" theme="neutral">
                This install will join the default group without changing its
                labels.
              </Text>
            )}

            {selectedGroup && !selectedGroup.default && (
              <div className="flex flex-col gap-3">
                <Text variant="subtext" theme="neutral">
                  Choose how this install joins {selectedGroup.name}.
                </Text>
                <RadioInput
                  name="group-assignment"
                  value="explicit"
                  checked={assignmentMode === 'explicit'}
                  onChange={() => onSelectAssignmentMode('explicit')}
                  labelProps={{
                    labelText: 'Pin install to this group',
                  }}
                />
                {selectedLabels.length > 0 && (
                  <RadioInput
                    name="group-assignment"
                    value="labels"
                    checked={assignmentMode === 'labels'}
                    onChange={() => onSelectAssignmentMode('labels')}
                    labelProps={{
                      labelText: 'Add group labels to install',
                    }}
                  />
                )}

                {assignmentMode === 'labels' && selectedLabels.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {selectedLabels.map(([key, value]) => (
                      <LabelBadge
                        key={key}
                        labelKey={key}
                        labelValue={value}
                        size="sm"
                      />
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        {targetBranch && targetGroups.length === 0 && (
          <Banner theme="warn">
            <strong>Warning:</strong> <strong>{targetBranch.name}</strong> has
            no install groups. Add a group before moving this install.
          </Banner>
        )}

        <div className="flex flex-col gap-2">
          <Text variant="subtext" theme="neutral">
            Moving to a new branch will:
          </Text>
          <ul className="flex flex-col gap-1 pl-4 list-disc">
            <li>
              <Text variant="subtext" theme="neutral">
                Switch the app config this install uses
              </Text>
            </li>
            <li>
              <Text variant="subtext" theme="neutral">
                Remove this install from the current branch
              </Text>
            </li>
            <li>
              <Text variant="subtext" theme="neutral">
                Trigger a reconciliation workflow immediately
              </Text>
            </li>
          </ul>
        </div>
      </div>
    </Modal>
  )
}
