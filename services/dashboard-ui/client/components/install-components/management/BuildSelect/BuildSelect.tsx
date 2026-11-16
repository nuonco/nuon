import { useState, useEffect, useMemo } from 'react'
import { Badge } from '@/components/common/Badge'
import { Status } from '@/components/common/Status'
import { Banner } from '@/components/common/Banner'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { SearchInput } from '@/components/common/SearchInput'
import { Select } from '@/components/common/form/Select'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { RadioInput } from '@/components/common/form/RadioInput'
import type { TBuild } from '@/types'

type TSortOrder = 'newest' | 'oldest'

const SORT_OPTIONS = [
  { value: 'newest', label: 'Newest first' },
  { value: 'oldest', label: 'Oldest first' },
]

const buildTimestamp = (build: TBuild) =>
  build?.created_at ? new Date(build.created_at).getTime() : 0

export const excludePreviewBuilds = (builds: TBuild[]) =>
  builds.filter((build) => !build?.is_preview)

export const buildBranchName = (build: TBuild) =>
  build.app_branch_run?.app_branch?.name

export const buildMatchesInstallConfig = (
  build: TBuild,
  installAppConfigId?: string
) => {
  const buildConfigId = build.component_config_connection?.app_config_id
  if (!installAppConfigId || !buildConfigId) return true
  return buildConfigId === installAppConfigId
}

export const buildAppBranchId = (build: TBuild) =>
  build.app_branch_id || build.app_branch_run?.app_branch?.id

export const buildMatchesInstallBranch = (
  build: TBuild,
  installAppBranchId?: string
) => {
  if (!installAppBranchId) return true
  const branchId = buildAppBranchId(build)
  if (!branchId) return false
  return branchId === installAppBranchId
}

const matchesQuery = (build: TBuild, query: string) => {
  const haystack = [
    build?.id,
    build?.vcs_connection_commit?.message,
    build?.created_by?.email,
    build?.component_config_connection?.external_image?.tag,
    buildBranchName(build),
  ]
    .filter(Boolean)
    .join(' ')
    .toLowerCase()
  return haystack.includes(query.toLowerCase())
}

const BuildRowLoading = () => (
  <div className="flex items-start gap-3 p-3 border rounded">
    <div className="flex items-start justify-between w-full">
      <div className="flex flex-col gap-1">
        <Text loading loadingWidth={24} />
        <Text variant="subtext" loading loadingWidth={32} />
        <div className="flex items-center gap-2">
          <Text variant="label" loading loadingWidth={12} />
          <Text variant="label" loading loadingWidth={14} />
        </div>
      </div>
      <div className="flex items-center gap-2">
        <Text variant="label" loading loadingWidth={6} />
      </div>
    </div>
  </div>
)

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
  installAppConfigId?: string
  installAppBranchId?: string
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
  installAppConfigId,
  installAppBranchId,
  onScroll,
}: IBuildSelect) => {
  const [searchQuery, setSearchQuery] = useState('')
  const [sortOrder, setSortOrder] = useState<TSortOrder>('newest')
  const deployableBuilds = useMemo(() => excludePreviewBuilds(builds), [builds])
  const branchBuilds = useMemo(
    () =>
      deployableBuilds.filter((build) =>
        buildMatchesInstallBranch(build, installAppBranchId)
      ),
    [deployableBuilds, installAppBranchId]
  )

  useEffect(() => {
    if (!selectedBuildId && branchBuilds.length > 0) {
      const mostRecentActiveBuild = branchBuilds.find(
        (build) => build?.status_v2?.status === 'active'
      )
      if (mostRecentActiveBuild) {
        onSelectBuild(mostRecentActiveBuild.id)
      }
    }
  }, [selectedBuildId, branchBuilds, onSelectBuild])

  const visibleBuilds = useMemo(() => {
    const filtered = searchQuery
      ? branchBuilds.filter((build) => matchesQuery(build, searchQuery))
      : branchBuilds
    return [...filtered].sort((a, b) =>
      sortOrder === 'newest'
        ? buildTimestamp(b) - buildTimestamp(a)
        : buildTimestamp(a) - buildTimestamp(b)
    )
  }, [branchBuilds, searchQuery, sortOrder])

  const showControls = branchBuilds.length > 0 && !error
  const selectedBuild = branchBuilds.find((build) => build.id === selectedBuildId)
  const selectedConfigMismatch =
    !!selectedBuild &&
    !installAppBranchId &&
    !buildMatchesInstallConfig(selectedBuild, installAppConfigId)
  const selectedBranchName = selectedBuild
    ? buildBranchName(selectedBuild)
    : undefined

  const renderContent = () => {
    if (isLoading && builds.length === 0) {
      return (
        <div className="flex flex-col gap-1">
          {[1, 2, 3, 4, 5].map((i) => (
            <BuildRowLoading key={i} />
          ))}
        </div>
      )
    }

    if (error) {
      return (
        <Banner theme="error">{error?.error || 'Unable to load builds'}</Banner>
      )
    }

    if (branchBuilds.length === 0 && !isLoading) {
      return (
        <EmptyState
          variant="search"
          size="sm"
          emptyTitle={
            installAppBranchId ? 'No builds on this branch' : 'No builds available'
          }
          emptyMessage={
            installAppBranchId
              ? 'Builds from other app branches are hidden. Run this branch to create a build.'
              : 'No builds found. Create a build first.'
          }
        />
      )
    }

    if (visibleBuilds.length === 0) {
      return (
        <EmptyState
          variant="search"
          size="sm"
          emptyTitle="No matching builds"
          emptyMessage="No builds match your search. Try a different term."
        />
      )
    }

    return (
      <>
        <div className="flex flex-col gap-1">
          {visibleBuilds.map((build) => {
            const isActive = build?.status_v2?.status === 'active'
            const isCurrentDeployment =
              currentBuildId && build.id === currentBuildId
            const externalImageTag =
              componentType === 'external_image'
                ? build?.component_config_connection?.external_image?.tag
                : undefined
            const branchName = buildBranchName(build)
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
                                nowrap
                                title={build.vcs_connection_commit.message}
                              >
                                {
                                  build.vcs_connection_commit.message.split(
                                    '\n'
                                  )[0]
                                }
                              </Text>
                              <Text theme="neutral">•</Text>
                            </>
                          )}
                          <Text variant="subtext" theme="neutral">
                            {build.created_by?.email || 'Unknown'}
                          </Text>
                          {branchName && (
                            <>
                              <Text theme="neutral">•</Text>
                              <Text variant="subtext" theme="neutral">
                                Branch {branchName}
                              </Text>
                            </>
                          )}
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
              <BuildRowLoading key={`loading-${i}`} />
            ))}
          </div>
        )}
      </>
    )
  }

  return (
    <div
      className="relative flex flex-col max-h-80 overflow-y-auto -mx-6 -mt-6 border-b border-cool-grey-200 dark:border-dark-grey-600"
      onScroll={onScroll}
    >
      {showControls && (
        <div className="sticky top-0 z-10 flex flex-col gap-3 bg-white dark:bg-dark-grey-900 px-6 py-3 border-b border-cool-grey-200 dark:border-dark-grey-600">
          {selectedConfigMismatch && (
            <Banner theme="warn">
              {selectedBranchName
                ? `This build is from the ${selectedBranchName} branch and uses a different app config than this install.`
                : 'This build uses a different app config than this install.'}
            </Banner>
          )}
          <div className="flex items-center gap-2">
            <SearchInput
              value={searchQuery}
              onChange={setSearchQuery}
              placeholder="Search builds…"
              labelClassName="flex-1"
              className="!min-w-0 w-full"
            />
            <div className="w-44 shrink-0">
              <Select
                value={sortOrder}
                onChange={(value) => setSortOrder(value as TSortOrder)}
                options={SORT_OPTIONS}
              />
            </div>
          </div>
        </div>
      )}
      <div className="px-6 py-6">{renderContent()}</div>
    </div>
  )
}
