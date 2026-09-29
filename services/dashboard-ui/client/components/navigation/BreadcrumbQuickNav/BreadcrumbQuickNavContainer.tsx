import { useEffect, useMemo, useState } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useLocation } from 'react-router'
import { getAppBranches, getInstalls } from '@/lib'
import {
  branchQuickNavHref,
  installQuickNavHref,
} from '@/lib/breadcrumb-quick-nav'
import type { TNavLink } from '@/types'
import {
  BreadcrumbQuickNav,
  type TBreadcrumbQuickNavItem,
} from './BreadcrumbQuickNav'

const LIMIT = 100

const useDebouncedValue = (value: string) => {
  const [debounced, setDebounced] = useState(value)

  useEffect(() => {
    const timeout = setTimeout(() => setDebounced(value), 250)
    return () => clearTimeout(timeout)
  }, [value])

  return debounced
}

const BranchQuickNav = ({
  label,
  config,
}: {
  label: string
  config: NonNullable<TNavLink['quickNav']>
}) => {
  const { pathname } = useLocation()
  const [isOpen, setIsOpen] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')
  const query = useDebouncedValue(searchTerm)
  const { data, isFetching } = useQuery({
    queryKey: [
      'app-branches',
      config.orgId,
      config.appId,
      'breadcrumb-quick-nav',
      query,
    ],
    queryFn: () =>
      getAppBranches({
        orgId: config.orgId,
        appId: config.appId,
        q: query || undefined,
        limit: LIMIT,
        offset: 0,
      }),
    placeholderData: keepPreviousData,
    enabled: isOpen,
  })

  const items = useMemo<TBreadcrumbQuickNavItem[]>(
    () =>
      (data?.data ?? []).flatMap((branch) => {
        if (!branch.id || !branch.name) return []
        const targetBasePath = `/${config.orgId}/apps/${config.appId}/branches/${branch.id}`
        return [
          {
            id: branch.id,
            name: branch.name,
            href: branchQuickNavHref({
              pathname,
              currentBasePath: config.basePath,
              targetBasePath,
            }),
          },
        ]
      }),
    [config, data?.data, pathname]
  )

  return (
    <BreadcrumbQuickNav
      id="breadcrumb-branch-quick-nav"
      label={label}
      title="Switch branch"
      items={items}
      currentId={config.resourceId}
      isLoading={isFetching}
      searchTerm={searchTerm}
      onSearch={setSearchTerm}
      onOpenChange={setIsOpen}
    />
  )
}

const InstallQuickNav = ({
  label,
  config,
}: {
  label: string
  config: NonNullable<TNavLink['quickNav']>
}) => {
  const { pathname } = useLocation()
  const [isOpen, setIsOpen] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')
  const query = useDebouncedValue(searchTerm)
  const { data, isFetching } = useQuery({
    queryKey: ['installs', config.orgId, 'breadcrumb-quick-nav', query],
    queryFn: () =>
      getInstalls({
        orgId: config.orgId,
        q: query || undefined,
        limit: LIMIT,
        offset: 0,
      }),
    placeholderData: keepPreviousData,
    enabled: isOpen,
  })

  const items = useMemo<TBreadcrumbQuickNavItem[]>(
    () =>
      (data?.data ?? []).flatMap((install) => {
        if (!install.id || !install.name || !install.app_id) return []
        const targetBasePath = `/${config.orgId}/apps/${install.app_id}/installs/${install.id}`
        return [
          {
            id: install.id,
            name: install.name,
            href: installQuickNavHref({
              pathname,
              currentBasePath: config.basePath,
              targetBasePath,
            }),
          },
        ]
      }),
    [config, data?.data, pathname]
  )

  return (
    <BreadcrumbQuickNav
      id="breadcrumb-install-quick-nav"
      label={label}
      title="Switch install"
      items={items}
      currentId={config.resourceId}
      isLoading={isFetching}
      searchTerm={searchTerm}
      onSearch={setSearchTerm}
      onOpenChange={setIsOpen}
    />
  )
}

export const BreadcrumbQuickNavContainer = ({
  label,
  config,
}: {
  label: string
  config: NonNullable<TNavLink['quickNav']>
}) =>
  config.type === 'branch' ? (
    <BranchQuickNav label={label} config={config} />
  ) : (
    <InstallQuickNav label={label} config={config} />
  )
