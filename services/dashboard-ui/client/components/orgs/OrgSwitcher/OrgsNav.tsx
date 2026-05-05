import { useRef, type KeyboardEvent } from 'react'
import { Avatar } from '@/components/common/Avatar'
import { Button } from '@/components/common/Button'
import { Link } from '@/components/common/Link'
import { SearchInput } from '@/components/common/SearchInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import type { TOrg } from '@/types'
import { OrgSummary } from './OrgSummary'

const LoadingOrgSummary = () => {
  return (
    <div className="flex gap-4 items-center p-2 w-full">
      <Avatar size="xl" isLoading />
      <div className="flex flex-col gap-1 transition-all w-full">
        <Skeleton height="14px" width="80%" />
        <Skeleton height="14px" width="40%" />
      </div>
    </div>
  )
}

interface IOrgsNav {
  orgs?: TOrg[]
  isLoading: boolean
  searchTerm: string
  onSearchChange: (value: string) => void
  onLoadMore: () => void
  showSearch: boolean
  showLoadMore: boolean
}

export const OrgsNav = ({
  orgs,
  isLoading,
  searchTerm,
  onSearchChange,
  onLoadMore,
  showSearch,
  showLoadMore,
}: IOrgsNav) => {
  const searchInputRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLDivElement>(null)

  const focusLinkAt = (index: number) => {
    const links = listRef.current?.querySelectorAll<HTMLAnchorElement>('a')
    if (!links?.length) return
    const wrapped = ((index % links.length) + links.length) % links.length
    links[wrapped]?.focus()
  }

  const handleSearchKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      e.stopPropagation()
      focusLinkAt(0)
    }
  }

  const handleLinkKeyDown = (
    e: KeyboardEvent<HTMLAnchorElement>,
    idx: number
  ) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      e.stopPropagation()
      focusLinkAt(idx + 1)
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      e.stopPropagation()
      if (idx === 0) {
        searchInputRef.current?.focus()
      } else {
        focusLinkAt(idx - 1)
      }
    }
  }

  return (
    <div className="w-full">
      {showSearch ? (
        <div className="p-2 w-full">
          <SearchInput
            ref={searchInputRef}
            labelClassName="md:!min-w-full md:!w-full"
            className="md:!min-w-full md:!w-full"
            placeholder="Search org by name..."
            value={searchTerm}
            onChange={onSearchChange}
            onKeyDown={handleSearchKeyDown}
          />
        </div>
      ) : null}
      {isLoading ? (
        Array.from({ length: 5 }).map((_, i) => <LoadingOrgSummary key={i} />)
      ) : orgs?.length ? (
        <div ref={listRef}>
          {orgs?.map((o, idx) => (
            <Link
              key={o?.id}
              className="!h-fit !block w-full"
              href={`/${o?.id}/apps`}
              variant="ghost"
              onKeyDown={(e) => handleLinkKeyDown(e, idx)}
            >
              <OrgSummary org={o} />
            </Link>
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center text-center w-full px-2 py-4">
          <Text variant="base" weight="strong">
            No org found
          </Text>
          <Text variant="subtext" theme="neutral">
            Clear your search and try again
          </Text>
        </div>
      )}
      {showLoadMore ? (
        <Button
          className="w-full justify-center mt-4"
          onClick={onLoadMore}
          variant="ghost"
        >
          Load more
        </Button>
      ) : null}
    </div>
  )
}
