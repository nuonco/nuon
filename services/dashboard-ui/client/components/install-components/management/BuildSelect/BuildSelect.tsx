import { useState, useEffect, useCallback } from 'react'
import { Badge } from '@/components/common/Badge'
import { Status } from '@/components/common/Status'
import { Banner } from '@/components/common/Banner'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { RadioInput } from '@/components/common/form/RadioInput'
import type { TBuild } from '@/types'

interface IBuildSelect {
  componentType?: string
  selectedBuildId?: string
  currentBuildId?: string
  currentDeployStatus?: string
  onSelectBuild: (buildId: string) => void
  builds: TBuild[]
  isLoading: boolean
  isLoadingMore: boolean
  hasMorePages: boolean
  error?: { error?: string } | null
  onScroll: (e: React.UIEvent<HTMLDivElement>) => void
}

export const BuildSelect = ({
  componentType,
  selectedBuildId,
  currentBuildId,
  currentDeployStatus,
  onSelectBuild,
  builds,
  isLoading,
  isLoadingMore,
  hasMorePages,
  error,
  onScroll,
}: IBuildSelect) => {
  useEffect(() => {
    if (!selectedBuildId && builds.length > 0) {
      const mostRecentActiveBuild = builds.find(
        (build) => build?.status_v2?.status === 'active'
      )
      if (mostRecentActiveBuild) {
        onSelectBuild(mostRecentActiveBuild.id)
      }
    }
  }, [selectedBuildId, builds, onSelectBuild])

  const renderContent = () => {
    if (isLoading && builds.length === 0) {
      return (
        <div className="flex flex-col gap-1">
          {[1, 2, 3, 4, 5].map((i) => (
            <div key={i} className="flex items-start gap-3 p-3 border rounded">
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
                  <Skeleton
                    width="60px"
                    height="20px"
                    className="rounded-full"
                  />
                  <Skeleton
                    width="50px"
                    height="20px"
                    className="rounded-full"
                  />
                </div>
              </div>
            </div>
          ))}
        </div>
      )
    }

    if (error) {
      return (
        <Banner theme="error">{error?.error || 'Unable to load builds'}</Banner>
      )
    }

    if (builds.length === 0 && !isLoading) {
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
          {builds.map((build) => {
            const isActive = build?.status_v2?.status === 'active'
            const isCurrentDeployment =
              currentBuildId && build.id === currentBuildId
            const externalImageTag =
              componentType === 'external_image'
                ? build?.component_config_connection?.external_image?.tag
                : undefined
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
                        <div className="flex items-center gap-2">
                          <Text
                            className="!leading-[1] font-mono"
                            variant="base"
                            weight="strong"
                          >
                            {build.id}
                          </Text>
                          {externalImageTag && (
                            <Badge size="sm" theme="neutral">
                              {externalImageTag}
                            </Badge>
                          )}
                        </div>
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
                        {isCurrentDeployment ? (
                          currentDeployStatus === 'active' ? (
                            <Badge size="sm" theme="info">
                              Current deployment
                            </Badge>
                          ) : currentDeployStatus === 'inactive' ? (
                            <Badge size="sm" theme="neutral">
                              Previously deployed
                            </Badge>
                          ) : null
                        ) : null}
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
                    <Skeleton
                      width="60px"
                      height="20px"
                      className="rounded-full"
                    />
                    <Skeleton
                      width="50px"
                      height="20px"
                      className="rounded-full"
                    />
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
      onScroll={onScroll}
    >
      <div className="px-6 pt-6">{renderContent()}</div>
    </div>
  )
}
