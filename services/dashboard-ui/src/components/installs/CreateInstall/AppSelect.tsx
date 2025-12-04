'use client'

import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Icon } from '@/components/common/Icon'
import { SearchInput } from '@/components/common/SearchInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { RadioInput } from '@/components/common/form/RadioInput'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import type { TApp } from '@/types'

interface AppSelectProps {
  onSelectApp: (app: TApp) => void
  onClose: () => void
}

export const AppSelect = ({ onSelectApp, onClose }: AppSelectProps) => {
  const { org } = useOrg()
  const [currentPage, setCurrentPage] = useState(0)
  const [searchQuery, setSearchQuery] = useState('')
  const limit = 10

  // Reset page when search changes
  const handleSearchChange = (query: string) => {
    setSearchQuery(query)
    setCurrentPage(0)
  }

  const searchParam = searchQuery ? `&q=${encodeURIComponent(searchQuery)}` : ''
  const {
    data: apps,
    isLoading,
    error,
  } = useQuery<TApp[]>({
    path: `/api/orgs/${org?.id}/apps?offset=${currentPage * limit}&limit=${limit}${searchParam}`,
  })

  const hasNext = apps && apps.length === limit
  const hasPrev = currentPage > 0

  const handlePrev = () => {
    setCurrentPage(prev => Math.max(0, prev - 1))
  }

  const handleNext = () => {
    setCurrentPage(prev => prev + 1)
  }

  const renderContent = () => {
    if (isLoading) {
      return (
        <div className="space-y-1">
          {[1, 2, 3, 4, 5].map((i) => (
            <div key={i} className="flex items-start gap-3 p-3 border rounded">
              <Skeleton width="16px" height="16px" className="rounded-full mt-0.5" />
              <div className="flex-1 flex flex-col gap-1">
                <Skeleton width="120px" height="18px" />
                <div className="flex items-center gap-2">
                  <Skeleton width="180px" height="12px" />
                  <span className="text-cool-grey-400 dark:text-cool-grey-500">•</span>
                  <Skeleton width="80px" height="12px" />
                </div>
              </div>
            </div>
          ))}
        </div>
      )
    }

    if (error) {
      return <Banner theme="error">{error.error || 'Unable to load apps'}</Banner>
    }

    const validApps =
      apps?.filter((app) => app?.runner_config?.app_runner_type) || []

    if (validApps.length === 0) {
      if (searchQuery) {
        return (
          <EmptyState
            variant="search"
            size="sm"
            emptyTitle="No apps found"
            emptyMessage={`No apps found matching "${searchQuery}". Try a different search term.`}
          />
        )
      } else if (currentPage === 0) {
        return (
          <EmptyState
            variant="search"
            size="sm"
            emptyTitle="No apps available"
            emptyMessage="No apps with runner configurations found. Create an app first."
          />
        )
      } else {
        return (
          <EmptyState
            variant="search"
            size="sm"
            emptyTitle="No more apps"
            emptyMessage="No more apps found."
          />
        )
      }
    }

    return (
      <div className="flex flex-col gap-4">
        <div className="space-y-1">
          {validApps.map((app) => (
            <RadioInput
              key={app.id}
              name="app-selection"
              value={app.id}
              onChange={() => onSelectApp(app)}
              labelProps={{
                labelText: (
                  <div className="flex flex-col">
                    <Text className="!leading-[1]" variant="base" weight="strong">
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
                ),
                className:
                  'flex items-start gap-3 p-3 border rounded hover:bg-cool-grey-50 dark:hover:bg-dark-grey-800 cursor-pointer',
              }}
            />
          ))}
        </div>
        
        {(hasNext || hasPrev) && (
          <div className="flex items-center justify-center gap-3 pt-4">
            <Button
              disabled={!hasPrev}
              onClick={handlePrev}
              title="previous"
            >
              <Icon variant="ArrowLeft" />
            </Button>
            
            <Button
              disabled={!hasNext}
              onClick={handleNext}
              title="next"
            >
              <Icon variant="ArrowRight" />
            </Button>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <SearchInput
        value={searchQuery}
        onChange={handleSearchChange}
        placeholder="Search apps..."
        className="w-full"
        labelClassName="w-full"
      />
      {renderContent()}
    </div>
  )
}