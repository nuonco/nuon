import { useEffect, useMemo, useState } from 'react'
import { useInfiniteQuery } from '@tanstack/react-query'
import { useLocation } from 'react-router'
import { getAppBranches } from '@/lib'
import { useApp } from '../../../providers/app-provider'
import { useAppBranch } from '../../../providers/app-branch-provider'
import { useOrg } from '../../../providers/org-provider'
import { appBranchHref } from '../../../utils/hrefs'
import { useDropdown } from '../../atoms/Dropdown'
import { BranchSwitcher } from './BranchSwitcher'

const PAGE_SIZE = 5

export const BranchSwitcherContainer = () => {
  const { pathname } = useLocation()
  const { orgId } = useOrg()
  const { appId } = useApp()
  const { branch, branchId } = useAppBranch()
  const dropdown = useDropdown()
  const isOpen = dropdown?.isOpen ?? true
  const [search, setSearch] = useState('')

  useEffect(() => {
    if (!isOpen) setSearch('')
  }, [isOpen])

  const query = useInfiniteQuery({
    queryKey: ['app-branches', orgId, appId, 'switcher', search],
    queryFn: ({ pageParam }) =>
      getAppBranches({
        orgId: orgId!,
        appId: appId!,
        limit: PAGE_SIZE,
        offset: pageParam,
        q: search || undefined,
      }),
    initialPageParam: 0,
    getNextPageParam: (lastPage) =>
      lastPage?.pagination?.hasNext
        ? lastPage.pagination.offset + lastPage.pagination.limit
        : undefined,
    enabled: !!orgId && !!appId && isOpen,
  })
  const branches = useMemo(
    () => query.data?.pages.flatMap((page) => page?.data ?? []) ?? [],
    [query.data]
  )
  const currentBase =
    orgId && appId && branchId ? appBranchHref(orgId, appId, branchId) : ''
  const sectionPath =
    currentBase && pathname.startsWith(currentBase)
      ? pathname.slice(currentBase.length)
      : ''

  return (
    <BranchSwitcher
      branches={branches}
      currentBranch={branch}
      search={search}
      onSearchChange={setSearch}
      onLoadMore={() => void query.fetchNextPage()}
      getBranchHref={(nextBranchId) =>
        orgId && appId
          ? `${appBranchHref(orgId, appId, nextBranchId)}${sectionPath}`
          : ''
      }
      isLoading={query.isLoading}
      isLoadingMore={query.isFetchingNextPage}
      hasMore={query.hasNextPage}
      hasError={query.isError}
    />
  )
}
