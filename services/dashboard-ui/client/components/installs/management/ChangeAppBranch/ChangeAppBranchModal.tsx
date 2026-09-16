import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { matchesSelector } from '@/components/match/matches'
import type { TAppBranch, TInstall } from '@/types'

const branchHasMatchingGroup = (
  branch: TAppBranch,
  installLabels: Record<string, string>
): boolean => {
  const groups = branch.configs?.at(0)?.install_groups ?? []
  return groups.some((g) => {
    if (g.all_installs) return true
    const matchLabels = g.label_selector?.match_labels ?? {}
    if (Object.keys(matchLabels).length > 0) {
      return matchesSelector(installLabels, g.label_selector)
    }
    return false
  })
}

interface IChangeAppBranchModal extends Omit<IModal, 'onSubmit'> {
  install: TInstall
  targetBranch: TAppBranch | null
  branches: TAppBranch[]
  isPending: boolean
  onSelectBranch: (branch: TAppBranch) => void
  onConfirm: () => void
}

export const ChangeAppBranchModal = ({
  install,
  targetBranch,
  branches,
  isPending,
  onSelectBranch,
  onConfirm,
  ...props
}: IChangeAppBranchModal) => {
  const installLabels = install.labels ?? {}
  const noMatchingGroup =
    targetBranch !== null &&
    !branchHasMatchingGroup(targetBranch, installLabels)

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
        disabled: !targetBranch || isPending,
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

        {noMatchingGroup && targetBranch && (
          <Banner theme="warn">
            <strong>Warning:</strong> The current labels on this install
            don&apos;t match any install group on{' '}
            <strong>{targetBranch.name}</strong>. This install won&apos;t
            receive deployments until it&apos;s added to a group.
            {Object.keys(installLabels).length > 0 && (
              <div className="flex flex-wrap gap-1 mt-2">
                {Object.entries(installLabels).map(([k, v]) => (
                  <LabelBadge key={k} labelKey={k} labelValue={v} size="sm" />
                ))}
              </div>
            )}
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
                Remove this install from the current branch&apos;s deployment
                plan
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
