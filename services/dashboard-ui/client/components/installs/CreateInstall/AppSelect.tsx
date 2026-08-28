import { useState, useEffect, useCallback } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Icon } from '@/components/common/Icon'
import { SearchInput } from '@/components/common/SearchInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { RadioInput } from '@/components/common/form/RadioInput'
import type { TApp } from '@/types'

interface AppSelectProps {
  apps: TApp[]
  isLoading: boolean
  isLoadingMore: boolean
  hasMorePages: boolean
  error: any
  searchQuery: string
  onSearchChange: (query: string) => void
  onLoadMore: () => void
  onSelectApp: (app: TApp) => void
  onClose: () => void
}

export const AppSelect = ({
  apps,
  isLoading,
  isLoadingMore,
  hasMorePages,
  error,
  searchQuery,
  onSearchChange,
  onLoadMore,
  onSelectApp,
  onClose,
}: AppSelectProps) => {
  const handleScroll = useCallback(
    (e: React.UIEvent<HTMLDivElement>) => {
      const { scrollTop, scrollHeight, clientHeight } = e.currentTarget
      const isNearBottom = scrollTop + clientHeight >= scrollHeight - 100

      if (isNearBottom && !isLoading && !isLoadingMore && hasMorePages) {
        onLoadMore()
      }
    },
    [isLoading, isLoadingMore, hasMorePages, onLoadMore]
  )

  const renderContent = () => {
    if (isLoading && apps.length === 0) {
      return (
        <div className="flex flex-col gap-1">
          {[1, 2, 3, 4, 5].map((i) => (
            <div
              key={i}
              className="flex items-start gap-3 hover:bg-black/5 dark:hover:bg-white/5 rounded-md py-2.5 px-3 border h-[66px]"
            >
              <Skeleton
                width="16px"
                height="16px"
                className="rounded-full shrink-0 mt-0.5"
              />
              <div className="flex items-start justify-between w-full">
                <div className="flex flex-col gap-0">
                  <Skeleton width="150px" height="20px" />
                  <div className="flex items-center gap-2">
                    <Skeleton width="168px" height="14px" />
                    <span className="text-cool-grey-400 dark:text-cool-grey-500">
                      •
                    </span>
                    <Skeleton width="122px" height="14px" />
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )
    }

    if (error) {
      return (
        <Banner theme="error">{error.error || 'Unable to load apps'}</Banner>
      )
    }

    if (apps.length === 0 && !isLoading) {
      if (searchQuery) {
        return (
          <EmptyState
            variant="search"
            size="sm"
            emptyTitle="No apps found"
            emptyMessage={`No apps found matching "${searchQuery}". Try a different search term.`}
          />
        )
      } else {
        return (
          <EmptyState
            variant="search"
            size="sm"
            emptyTitle="No apps available"
            emptyMessage="No apps found. Create an app first."
          />
        )
      }
    }

    return (
      <>
        <div className="flex flex-col gap-1">
          {apps.map((app) => {
            const isProvisionable = app?.runner_config?.app_runner_type
            return (
              <RadioInput
                key={app.id}
                name="app-selection"
                value={app.id}
                disabled={!isProvisionable}
                onChange={() => isProvisionable && onSelectApp(app)}
                labelProps={{
                  labelText: (
                    <div className="flex items-start justify-between w-full">
                      <div className="flex flex-col">
                        <Text
                          className="!leading-[1]"
                          variant="base"
                          weight="strong"
                        >
                          {app.name}
                        </Text>
                        <div className="flex items-center gap-2">
                          <Text variant="subtext" theme="neutral">
                            {app.id}
                          </Text>
                          {app.updated_at && (
                            <>
                              <Text theme="neutral">•</Text>
                              <Time
                                time={app.updated_at}
                                variant="subtext"
                                theme="neutral"
                              />
                            </>
                          )}
                        </div>
                      </div>
                      {!isProvisionable && (
                        <Badge size="sm" theme="neutral">
                          Not provisionable
                        </Badge>
                      )}
                    </div>
                  ),
                  className: `flex items-start gap-3 p-3 border rounded ${
                    !isProvisionable
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
                className="flex items-start gap-3 hover:bg-black/5 dark:hover:bg-white/5 rounded-md py-2.5 px-3 border h-[66px]"
              >
                <Skeleton
                  width="16px"
                  height="16px"
                  className="rounded-full shrink-0 mt-0.5"
                />
                <div className="flex items-start justify-between w-full">
                  <div className="flex flex-col gap-0">
                    <Skeleton width="150px" height="20px" />
                    <div className="flex items-center gap-2">
                      <Skeleton width="168px" height="14px" />
                      <span className="text-cool-grey-400 dark:text-cool-grey-500">
                        •
                      </span>
                      <Skeleton width="122px" height="14px" />
                    </div>
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
      className="relative flex flex-col gap-4 max-h-80 overflow-y-auto -mx-6 -my-6 pb-6"
      onScroll={handleScroll}
    >
      <div className="sticky border-b top-0 bg-background z-10 px-6 py-4 shadow-sm">
        <SearchInput
          value={searchQuery}
          onChange={onSearchChange}
          placeholder="Search apps..."
          className="w-full"
          labelClassName="w-full"
        />
      </div>

      <div className="px-6">{renderContent()}</div>
    </div>
  )
}
