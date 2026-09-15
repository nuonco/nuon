import { useMemo, useState, useEffect } from 'react'
import { keepPreviousData, useQueries, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getApps, getComponents } from '@/lib'
import type { TApp, TComponent } from '@/types'
import { AppSelect } from './AppSelect'
import {
  appInstallBadge,
  hasRunnerConfig,
  latestComponentIds,
  type AppInstallBadge,
} from './app-install-readiness'

interface AppSelectContainerProps {
  onSelectApp: (app: TApp) => void
  onClose: () => void
}

export const AppSelectContainer = ({ onSelectApp, onClose }: AppSelectContainerProps) => {
  const { org } = useOrg()
  const [allApps, setAllApps] = useState<TApp[]>([])
  const [currentPage, setCurrentPage] = useState(0)
  const [searchQuery, setSearchQuery] = useState('')
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const [hasMorePages, setHasMorePages] = useState(true)
  const limit = 5

  const handleSearchChange = (query: string) => {
    setSearchQuery(query)
    setCurrentPage(0)
    setAllApps([])
    setHasMorePages(true)
  }

  const {
    data: apps,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['apps', org?.id, currentPage, limit, searchQuery],
    queryFn: () => getApps({
      orgId: org.id,
      offset: currentPage * limit,
      limit,
      q: searchQuery || undefined,
    }),
    enabled: !!org?.id,
    placeholderData: keepPreviousData,
  })

  useEffect(() => {
    if (apps) {
      const appData = apps.data
      if (currentPage === 0) {
        setAllApps(appData)
      } else {
        setAllApps((prev) => {
          const existingIds = new Set(prev.map((app) => app.id))
          const newApps = appData.filter((app) => !existingIds.has(app.id))
          return [...prev, ...newApps]
        })
      }

      setHasMorePages(appData.length === limit)

      if (isLoadingMore) {
        setTimeout(() => setIsLoadingMore(false), 800)
      }
    }
  }, [apps, currentPage, isLoadingMore])

  const appsNeedingComponents = useMemo(
    () =>
      allApps.filter(
        (app) => hasRunnerConfig(app) && latestComponentIds(app).length > 0
      ),
    [allApps]
  )

  const componentQueries = useQueries({
    queries: appsNeedingComponents.map((app) => ({
      queryKey: ['components', org?.id, app.id, 'create-install-gate'],
      queryFn: () =>
        getComponents({ orgId: org.id, appId: app.id!, limit: 100 }),
      enabled: !!org?.id && !!app.id,
    })),
  })

  const badges = useMemo(() => {
    const componentsByAppId: Record<string, TComponent[] | undefined> = {}
    appsNeedingComponents.forEach((app, index) => {
      if (app.id) componentsByAppId[app.id] = componentQueries[index]?.data?.data
    })

    const next: Record<string, AppInstallBadge | undefined> = {}
    for (const app of allApps) {
      if (!app.id) continue
      next[app.id] = appInstallBadge(app, componentsByAppId[app.id])
    }
    return next
  }, [allApps, appsNeedingComponents, componentQueries])

  return (
    <AppSelect
      apps={allApps}
      badges={badges}
      isLoading={isLoading}
      isLoadingMore={isLoadingMore}
      hasMorePages={hasMorePages}
      error={error}
      searchQuery={searchQuery}
      onSearchChange={handleSearchChange}
      onLoadMore={() => {
        setIsLoadingMore(true)
        setCurrentPage((prev) => prev + 1)
      }}
      onSelectApp={onSelectApp}
      onClose={onClose}
    />
  )
}
