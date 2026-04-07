import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getApps } from '@/lib'
import type { TApp } from '@/types'
import { AppSelect } from './AppSelect'

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

  return (
    <AppSelect
      apps={allApps}
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
