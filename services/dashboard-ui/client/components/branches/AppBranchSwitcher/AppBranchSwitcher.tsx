import { useMemo, useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Menu } from '@/components/common/Menu'
import { SearchInput } from '@/components/common/SearchInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import type { TAppBranch } from '@/types'

const SEARCH_THRESHOLD = 6

export interface IAppBranchSwitcher {
  branches: TAppBranch[]
  currentBranch: TAppBranch
  orgId: string
  appId: string
  isLoading: boolean
}

export const AppBranchSwitcher = ({
  branches,
  currentBranch,
  orgId,
  appId,
  isLoading,
}: IAppBranchSwitcher) => {
  const [searchTerm, setSearchTerm] = useState('')

  const filteredBranches = useMemo(() => {
    if (!searchTerm) return branches
    const q = searchTerm.toLowerCase()
    return branches.filter((b) => b.name?.toLowerCase().includes(q))
  }, [branches, searchTerm])

  const showSearch = branches.length > SEARCH_THRESHOLD

  return (
    <Dropdown
      id="app-branch-switcher"
      variant="ghost"
      position="below"
      alignment="left"
      hideIcon
      buttonClassName="!p-0 !h-fit !border-0 !bg-transparent !rounded-full"
      buttonText={
        <Badge
          size="sm"
          theme="brand"
          className="cursor-pointer transition-colors hover:!border-primary-300 dark:hover:!border-[#4A2D69]"
        >
          <Icon variant="GitBranchIcon" size={13} />
          {currentBranch.name}
          <Icon variant="CaretUpDownIcon" size={12} />
        </Badge>
      }
    >
      <Menu className="w-64 max-h-[400px] overflow-y-auto">
        <Text>Switch branch</Text>
        {showSearch ? (
          <SearchInput
            className="md:!min-w-full md:!w-full"
            labelClassName="md:!min-w-full md:!w-full"
            placeholder="Search branches..."
            value={searchTerm}
            onChange={setSearchTerm}
          />
        ) : null}
        {isLoading ? (
          <div className="flex flex-col gap-1 p-1">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} height="32px" />
            ))}
          </div>
        ) : filteredBranches.length ? (
          filteredBranches.map((b) => {
            const isCurrent = b.id === currentBranch.id
            return (
              <Link
                key={b.id}
                href={`/${orgId}/apps/${appId}/branches/${b.id}`}
                variant="ghost"
                className={cn('items-center', {
                  '!text-primary-600 dark:!text-primary-400': isCurrent,
                })}
              >
                <span className="flex items-center gap-2 min-w-0">
                  <Icon variant="GitBranchIcon" size={14} className="shrink-0" />
                  <span className="truncate">{b.name}</span>
                </span>
                {isCurrent ? (
                  <Icon variant="CheckIcon" size={14} className="shrink-0" />
                ) : null}
              </Link>
            )
          })
        ) : (
          <div className="px-2 py-4 text-center">
            <Text variant="subtext" theme="neutral">
              No branches found
            </Text>
          </div>
        )}
      </Menu>
    </Dropdown>
  )
}
