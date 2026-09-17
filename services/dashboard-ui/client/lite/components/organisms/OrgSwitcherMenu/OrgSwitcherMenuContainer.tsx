import { useEffect, useMemo, useState } from 'react'
import { useInfiniteQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { getPaginatedOrgs } from '@/lib'
import { useDropdown } from '../../atoms/Dropdown'
import { OrgSwitcherMenu } from './OrgSwitcherMenu'

const PAGE_SIZE = 5

export const OrgSwitcherMenuContainer = () => {
  const { orgId } = useParams<{ orgId: string }>()
  const dropdown = useDropdown()
  const isOpen = dropdown?.open ?? true
  const [search, setSearch] = useState('')

  useEffect(() => {
    if (isOpen) setSearch('')
  }, [isOpen])

  const query = useInfiniteQuery({
    queryKey: ['orgs', 'switcher', search],
    queryFn: ({ pageParam }) =>
      getPaginatedOrgs({
        limit: PAGE_SIZE,
        offset: pageParam,
        q: search || undefined,
      }),
    initialPageParam: 0,
    getNextPageParam: (lastPage) =>
      lastPage?.pagination?.hasNext
        ? lastPage.pagination.offset + lastPage.pagination.limit
        : undefined,
    enabled: isOpen,
  })
  const orgs = useMemo(
    () => query.data?.pages.flatMap((page) => page?.data ?? []) ?? [],
    [query.data]
  )

  return (
    <OrgSwitcherMenu
      orgs={orgs}
      currentOrgId={orgId}
      search={search}
      onSearchChange={setSearch}
      onLoadMore={() => void query.fetchNextPage()}
      loading={query.isLoading}
      loadingMore={query.isFetchingNextPage}
      hasMore={query.hasNextPage}
      hasError={query.isError}
    />
  )
}
