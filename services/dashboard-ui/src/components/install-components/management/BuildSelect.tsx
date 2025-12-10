'use client'

import { useState, useEffect, useCallback } from 'react'
import { Badge } from '@/components/common/Badge'
import { Status } from '@/components/common/Status'
import { Banner } from '@/components/common/Banner'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { RadioInput } from '@/components/common/form/RadioInput'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import type { TBuild } from '@/types'

interface BuildSelectProps {
  componentId: string
  selectedBuildId?: string
  currentBuildId?: string
  onSelectBuild: (buildId: string) => void
  onClose: () => void
}

export const BuildSelect = ({
  componentId,
  selectedBuildId,
  currentBuildId,
  onSelectBuild,
  onClose,
}: BuildSelectProps) => {
  const { org } = useOrg()
  const [allBuilds, setAllBuilds] = useState<TBuild[]>([])
  const [currentPage, setCurrentPage] = useState(0)
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const [hasMorePages, setHasMorePages] = useState(true)
  const limit = 5

  const {
    data: builds,
    isLoading,
    error,
  } = useQuery<TBuild[]>({
    path: `/api/orgs/${org?.id}/components/${componentId}/builds?offset=${currentPage * limit}&limit=${limit}`,
  })

  // Update accumulated builds when new data comes in
  useEffect(() => {
    if (builds) {
      if (currentPage === 0) {
        // First page or search reset
        setAllBuilds(builds)
      } else {
        // Append to existing builds, but deduplicate by ID
        setAllBuilds((prev) => {
          const existingIds = new Set(prev.map((build) => build.id))
          const newBuilds = builds.filter((build) => !existingIds.has(build.id))
          return [...prev, ...newBuilds]
        })
      }

      // Check if we have more pages
      setHasMorePages(builds.length === limit)

      // Add a delay to ensure loading state is visible
      if (isLoadingMore) {
        setTimeout(() => setIsLoadingMore(false), 800)
      }
    }
  }, [builds, currentPage, isLoadingMore])

  // Auto-select most recent active build on first load
  useEffect(() => {
    if (!selectedBuildId && allBuilds.length > 0 && currentPage === 0) {
      // Find the most recent active build (API returns in descending order by created_at)
      const mostRecentActiveBuild = allBuilds.find(build => build?.status_v2?.status === 'active')
      if (mostRecentActiveBuild) {
        onSelectBuild(mostRecentActiveBuild.id)
      }
    }
  }, [selectedBuildId, allBuilds, currentPage, onSelectBuild])

  // Load more when scrolling near bottom
  const handleScroll = useCallback(
    (e: React.UIEvent<HTMLDivElement>) => {
      const { scrollTop, scrollHeight, clientHeight } = e.currentTarget
      const isNearBottom = scrollTop + clientHeight >= scrollHeight - 100 // 100px from bottom

      if (isNearBottom && !isLoading && !isLoadingMore && hasMorePages) {
        setIsLoadingMore(true)
        setCurrentPage((prev) => prev + 1)
      }
    },
    [isLoading, isLoadingMore, hasMorePages]
  )

  const renderContent = () => {
    if (isLoading && currentPage === 0) {
      // Initial loading skeleton
      return (
        <div className="flex flex-col gap-1">
          {[1, 2, 3, 4, 5].map((i) => (
            <div
              key={i}
              className="flex items-start gap-3 p-3 border rounded"
            >
              <Skeleton
                width="16px"
                height="16px"
                className="rounded-full shrink-0 mt-0.5"
              />
              <div className="flex items-start justify-between w-full">
                <div className="flex flex-col gap-1">
                  <Skeleton width="240px" height="20px" />
                  <Skeleton width="320px" height="14px" />
                  <div className="flex items-center gap-2">
                    <Skeleton width="120px" height="12px" />
                    <span className="text-cool-grey-400 dark:text-cool-grey-500">
                      •
                    </span>
                    <Skeleton width="140px" height="12px" />
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Skeleton width="60px" height="20px" className="rounded-full" />
                  <Skeleton width="50px" height="20px" className="rounded-full" />
                </div>
              </div>
            </div>
          ))}
        </div>
      )
    }

    if (error) {
      return (
        <Banner theme="error">{error.error || 'Unable to load builds'}</Banner>
      )
    }

    if (allBuilds.length === 0 && !isLoading) {
      return (
        <EmptyState
          variant="search"
          size="sm"
          emptyTitle="No builds available"
          emptyMessage="No builds found. Create a build first."
        />
      )
    }

    return (
      <>
        <div className="flex flex-col gap-1">
          {allBuilds.map((build) => {
            const isActive = build?.status_v2?.status === 'active'
            const isCurrentDeployment = currentBuildId && build.id === currentBuildId
            return (
              <RadioInput
                key={build.id}
                name="build-selection"
                value={build.id}
                checked={selectedBuildId === build.id}
                disabled={!isActive}
                onChange={() => {
                  if (isActive) {
                    onSelectBuild(build.id)
                  }
                }}
                labelProps={{
                  labelText: (
                    <div className="flex items-start justify-between w-full">
                      <div className="flex flex-col">
                        <Text
                          className="!leading-[1] font-mono"
                          variant="base"
                          weight="strong"
                        >
                          {build.id}
                        </Text>
                        <div className="flex items-center gap-2">
                          {build?.vcs_connection_commit?.message && (
                            <>
                              <Text
                                className="!leading-relaxed max-w-[280px] truncate"
                                variant="subtext"
                                theme="neutral"
                                title={build.vcs_connection_commit.message}
                              >
                                {build.vcs_connection_commit.message}
                              </Text>
                              <Text theme="neutral">•</Text>
                            </>
                          )}
                          <Text variant="subtext" theme="neutral">
                            {build.created_by?.email || 'Unknown'}
                          </Text>
                          {build.created_at && (
                            <>
                              <Text theme="neutral">•</Text>
                              <Time
                                time={build.created_at}
                                variant="subtext"
                                theme="neutral"
                              />
                            </>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        {isCurrentDeployment && (
                          <Badge size="sm" theme="info">
                            Current deployment
                          </Badge>
                        )}
                        {build?.status_v2?.status && (
                          <Status
                            status={build.status_v2.status}
                            variant="badge"
                          />
                        )}
                      </div>
                    </div>
                  ),
                  className: `flex items-start gap-3 p-3 border rounded ${
                    !isActive
                      ? 'opacity-50 bg-cool-grey-100 dark:bg-dark-grey-800'
                      : ''
                  }`,
                }}
              />
            )
          })}
        </div>

        {isLoadingMore && (
          <div className="flex flex-col gap-1 mt-1">
            {[1, 2, 3].map((i) => (
              <div
                key={`loading-${i}`}
                className="flex items-start gap-3 p-3 border rounded"
              >
                <Skeleton
                  width="16px"
                  height="16px"
                  className="rounded-full shrink-0 mt-0.5"
                />
                <div className="flex items-start justify-between w-full">
                  <div className="flex flex-col gap-1">
                    <Skeleton width="240px" height="20px" />
                    <Skeleton width="320px" height="14px" />
                    <div className="flex items-center gap-2">
                      <Skeleton width="120px" height="12px" />
                      <span className="text-cool-grey-400 dark:text-cool-grey-500">
                        •
                      </span>
                      <Skeleton width="140px" height="12px" />
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Skeleton width="60px" height="20px" className="rounded-full" />
                    <Skeleton width="50px" height="20px" className="rounded-full" />
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </>
    )
  }

  return (
    <div
      className="relative flex flex-col max-h-80 overflow-y-auto -mx-6 -my-6 pb-6"
      onScroll={handleScroll}
    >
      <div className="px-6 pt-6">{renderContent()}</div>
    </div>
  )
}
