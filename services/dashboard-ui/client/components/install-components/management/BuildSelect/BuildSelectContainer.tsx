import { useState, useEffect, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getComponentBuilds } from '@/lib'
import type { TBuild } from '@/types'
import { BuildSelect } from './BuildSelect'

interface IBuildSelectContainer {
  componentId: string
  componentType?: string
  selectedBuildId?: string
  currentBuildId?: string
  currentDeployStatus?: string
  onSelectBuild: (buildId: string) => void
  onClose: () => void
}

export const BuildSelectContainer = ({
  componentId,
  componentType,
  selectedBuildId,
  currentBuildId,
  currentDeployStatus,
  onSelectBuild,
  onClose,
}: IBuildSelectContainer) => {
  const { org } = useOrg()
  const [allBuilds, setAllBuilds] = useState<TBuild[]>([])
  const [currentPage, setCurrentPage] = useState(0)
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const [hasMorePages, setHasMorePages] = useState(true)
  const limit = 25

  const {
    data: buildsResult,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['component-builds-select', org?.id, componentId, currentPage],
    queryFn: () =>
      getComponentBuilds({
        orgId: org.id,
        componentId,
        offset: currentPage * limit,
        limit,
      }),
    enabled: !!org?.id && !!componentId,
  })

  useEffect(() => {
    if (buildsResult) {
      const builds = buildsResult.data
      if (currentPage === 0) {
        setAllBuilds(builds)
      } else {
        setAllBuilds((prev) => {
          const existingIds = new Set(prev.map((build) => build.id))
          const newBuilds = builds.filter((build) => !existingIds.has(build.id))
          return [...prev, ...newBuilds]
        })
      }

      setHasMorePages(buildsResult.pagination?.hasNext ?? false)

      if (isLoadingMore) {
        setTimeout(() => setIsLoadingMore(false), 800)
      }
    }
  }, [buildsResult, currentPage, isLoadingMore])

  const handleScroll = useCallback(
    (e: React.UIEvent<HTMLDivElement>) => {
      const { scrollTop, scrollHeight, clientHeight } = e.currentTarget
      const isNearBottom = scrollTop + clientHeight >= scrollHeight - 100

      if (isNearBottom && !isLoading && !isLoadingMore && hasMorePages) {
        setIsLoadingMore(true)
        setCurrentPage((prev) => prev + 1)
      }
    },
    [isLoading, isLoadingMore, hasMorePages]
  )

  return (
    <BuildSelect
      componentType={componentType}
      selectedBuildId={selectedBuildId}
      currentBuildId={currentBuildId}
      currentDeployStatus={currentDeployStatus}
      onSelectBuild={onSelectBuild}
      builds={allBuilds}
      isLoading={isLoading}
      isLoadingMore={isLoadingMore}
      hasMorePages={hasMorePages}
      error={error as any}
      onScroll={handleScroll}
    />
  )
}
