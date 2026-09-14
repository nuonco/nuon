import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import type { TAppBranch } from '@/types'

export type TBranchGroupSelection = {
  branchId: string
  groupId: string
}

interface IBranchGroupPicker {
  branches: TAppBranch[]
  selected: TBranchGroupSelection | null
  onSelect: (selection: TBranchGroupSelection) => void
}

const groupKind = (
  group: NonNullable<NonNullable<TAppBranch['configs']>[number]['install_groups']>[number]
) => {
  if (group.all_installs) return 'All installs'
  const labels = Object.entries(group.label_selector?.match_labels ?? {})
  if (labels.length > 0) return null
  const count = group.install_ids?.length ?? 0
  return `${count} install${count !== 1 ? 's' : ''} by ID`
}

export const BranchGroupPicker = ({
  branches,
  selected,
  onSelect,
}: IBranchGroupPicker) => {
  const hasAnyGroup = branches.some(
    (branch) => (branch.configs?.at(0)?.install_groups ?? []).length > 0
  )

  if (branches.length === 0) {
    return (
      <Banner theme="error">
        This app has no app branches. Refresh and try again.
      </Banner>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <Text variant="subtext" theme="neutral">
        Choose an app branch and install group. The install will use that
        branch&apos;s app config.
      </Text>

      {!hasAnyGroup ? (
        <Banner theme="error">
          No install groups are configured. Add an install group to an app
          branch before creating an install.
        </Banner>
      ) : null}

      <div className="flex flex-col gap-3">
        {branches.map((branch) => {
          const latestConfig = branch.configs?.at(0)
          const groups = latestConfig?.install_groups ?? []
          const branchId = branch.id || ''

          return (
            <Expand
              key={branchId}
              id={`create-install-branch-${branchId}`}
              isOpen
              heading={
                <div className="flex items-center gap-2">
                  <Icon variant="GitBranchIcon" size={14} />
                  <Text variant="body" weight="strong">
                    {branch.name}
                  </Text>
                  <Badge size="sm" theme="info">
                    {groups.length} group{groups.length !== 1 ? 's' : ''}
                  </Badge>
                </div>
              }
              headerClassName="!px-3"
              className="border rounded-md"
            >
              <div className="flex flex-col gap-2 p-3 border-t">
                {groups.length === 0 ? (
                  <Text variant="subtext" theme="neutral">
                    No install groups in this branch
                  </Text>
                ) : (
                  groups.map((group, idx) => {
                    const groupId = group.id || ''
                    const isSelected =
                      selected?.branchId === branchId &&
                      selected?.groupId === groupId
                    const labels = Object.entries(
                      group.label_selector?.match_labels ?? {}
                    )
                    const kind = groupKind(group)

                    return (
                      <button
                        key={groupId || idx}
                        type="button"
                        onClick={() =>
                          groupId && onSelect({ branchId, groupId })
                        }
                        disabled={!groupId}
                        className={`flex items-center justify-between gap-3 px-3 py-2.5 rounded-md text-left ${
                          isSelected
                            ? 'bg-primary-50 dark:bg-primary-900/30 ring-1 ring-primary-500'
                            : 'bg-cool-grey-50 dark:bg-dark-grey-700'
                        }`}
                      >
                        <div className="flex items-center gap-2 flex-wrap min-w-0">
                          <Text variant="body" weight="strong">
                            {group.name}
                          </Text>
                          {labels.map(([k, v]) => (
                            <LabelBadge
                              key={k}
                              labelKey={k}
                              labelValue={v}
                              size="sm"
                            />
                          ))}
                          {kind ? (
                            <Text variant="subtext" theme="neutral">
                              {kind}
                            </Text>
                          ) : null}
                        </div>
                        {isSelected ? (
                          <span className="flex shrink-0 items-center gap-1 text-xs text-primary-700 dark:text-primary-300">
                            <Icon variant="CheckIcon" size={14} />
                            Selected
                          </span>
                        ) : null}
                      </button>
                    )
                  })
                )}
              </div>
            </Expand>
          )
        })}
      </div>
    </div>
  )
}
