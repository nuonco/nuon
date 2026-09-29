import { useMemo } from 'react'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Menu } from '@/components/common/Menu'
import { SearchInput } from '@/components/common/SearchInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

export type TBreadcrumbQuickNavItem = {
  id: string
  name: string
  href: string
}

export interface IBreadcrumbQuickNav {
  id: string
  label: string
  title: string
  items: TBreadcrumbQuickNavItem[]
  currentId: string
  isLoading: boolean
  searchTerm: string
  onSearch: (value: string) => void
  onOpenChange?: (isOpen: boolean) => void
}

export const BreadcrumbQuickNav = ({
  id,
  label,
  title,
  items,
  currentId,
  isLoading,
  searchTerm,
  onSearch,
  onOpenChange,
}: IBreadcrumbQuickNav) => {
  const visibleItems = useMemo(
    () => items.filter((item) => item.id !== currentId),
    [currentId, items]
  )

  return (
    <Dropdown
      id={id}
      variant="ghost"
      position="below"
      alignment="left"
      onOpenChange={onOpenChange}
      icon={<Icon variant="CaretDownIcon" size={12} />}
      buttonClassName="!p-0 !h-fit !border-0 !bg-transparent gap-1 max-w-48"
      buttonText={
        <span className="truncate text-sm font-semibold">{label}</span>
      }
    >
      <Menu className="w-72 max-h-[400px]">
        <Text>{title}</Text>
        <SearchInput
          className="md:!min-w-full md:!w-full"
          labelClassName="md:!min-w-full md:!w-full"
          placeholder={`Search ${title.toLowerCase()}...`}
          value={searchTerm}
          onChange={onSearch}
        />
        <div className="flex flex-col gap-0.5 overflow-y-auto mt-2">
          {isLoading ? (
            Array.from({ length: 4 }).map((_, index) => (
              <Skeleton key={index} height="32px" />
            ))
          ) : visibleItems.length ? (
            visibleItems.map((item) => (
              <Link
                key={item.id}
                href={item.href}
                variant="ghost"
                className={cn(
                  '!p-2 text-sm !leading-none h-8 w-full !rounded-md',
                  'items-center justify-between'
                )}
              >
                <span className="truncate">{item.name}</span>
              </Link>
            ))
          ) : (
            <Text
              variant="subtext"
              theme="neutral"
              className="px-2 py-4 text-center"
            >
              No other results
            </Text>
          )}
        </div>
      </Menu>
    </Dropdown>
  )
}
