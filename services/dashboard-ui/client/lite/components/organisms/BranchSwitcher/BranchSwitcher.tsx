import type { TAppBranch } from '@/types'
import { Button } from '../../atoms/Button'
import { Dropdown } from '../../atoms/Dropdown'
import { Icon } from '../../atoms/Icon'
import { Text } from '../../atoms/Text'
import { SwitcherMenu } from '../../molecules/SwitcherMenu'

export interface IBranchSwitcher {
  branches: TAppBranch[]
  currentBranch?: TAppBranch
  search: string
  onSearchChange: (value: string) => void
  onLoadMore: () => void
  getBranchHref: (branchId: string) => string
  isLoading?: boolean
  isLoadingMore?: boolean
  hasMore?: boolean
  hasError?: boolean
}

export const BranchSwitcher = ({
  branches,
  currentBranch,
  getBranchHref,
  ...props
}: IBranchSwitcher) => (
  <Dropdown
    align="end"
    trigger={
      <Button
        variant="secondary"
        icon={<Icon variant="GitBranchIcon" size={16} />}
        loading={!currentBranch && props.isLoading}
      >
        <Text family="mono" color="inherit">
          {currentBranch?.name ?? 'Select branch'}
        </Text>
      </Button>
    }
  >
    <SwitcherMenu
      {...props}
      items={branches.flatMap((branch) =>
        branch?.id && branch?.name
          ? [
              {
                id: branch.id,
                label: branch.name,
                href: getBranchHref(branch.id),
              },
            ]
          : []
      )}
      selectedId={currentBranch?.id}
      searchLabel="Search branches"
      searchPlaceholder="Search branches..."
      emptyTitle="No branches found"
      errorTitle="Branches failed to load"
    />
  </Dropdown>
)
