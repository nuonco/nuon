import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { SearchInput } from '@/components/common/SearchInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TVCSConnectionReposResponse } from '@/types'
import {
  REPO_FILTER_OPTIONS,
  RepoTypeFilter,
  type TRepoFilterType,
} from './RepoTypeFilter'

interface IRepositoriesSection {
  repos?: TVCSConnectionReposResponse
  error?: any
  isLoading: boolean
}

export const RepositoriesSection = ({
  repos,
  error,
  isLoading,
}: IRepositoriesSection) => {
  const [search, setSearch] = useState('')
  const [typeFilter, setTypeFilter] =
    useState<TRepoFilterType[]>(REPO_FILTER_OPTIONS)

  const filteredRepos = repos?.repositories?.filter((repo) => {
    if (!repo.full_name?.toLowerCase().includes(search.toLowerCase()))
      return false
    const isPublic = !repo.private
    return (
      (typeFilter.includes('private') && repo.private) ||
      (typeFilter.includes('public') && isPublic) ||
      (typeFilter.includes('fork') && repo.fork)
    )
  })

  const hasRepos = repos && repos.repositories && repos.repositories.length > 0
  const isFiltering = search || typeFilter.length < REPO_FILTER_OPTIONS.length

  return (
    <div className="flex flex-col gap-4">
      <Text variant="body" weight="strong">
        Connected repositories
      </Text>

      {error ? (
        <Banner theme="error">
          {error?.error || 'Unable to load repositories.'}
        </Banner>
      ) : (
        <>
          <div className="flex justify-between items-center w-full">
            <div className="flex items-center gap-4">
              <SearchInput
                placeholder="Search repositories..."
                value={search}
                onChange={setSearch}
              />
              {isLoading ? (
                <Skeleton height="17px" width="120px" />
              ) : (
                <Text
                  variant="subtext"
                  theme="neutral"
                  className="whitespace-nowrap"
                >
                  {filteredRepos?.length} of {repos.total_count}{' '}
                  {repos.total_count === 1 ? 'repository' : 'repositories'}
                </Text>
              )}
            </div>
            <RepoTypeFilter selected={typeFilter} onChange={setTypeFilter} />
          </div>

          <div className="flex flex-col gap-2">
            {isLoading
              ? Array.from({ length: 5 }).map((_, idx) => (
                  <div
                    key={idx}
                    className="flex flex-col gap-2 py-2 px-4 border rounded-md"
                  >
                    <div className="flex items-center gap-4">
                      <Skeleton height="24px" width="110px" />
                      <Skeleton height="20px" width="60px" />
                      <Skeleton height="20px" width="40px" />
                      <Skeleton height="16px" width="110px" />
                    </div>
                    <Skeleton height="17px" width="400px" />
                  </div>
                ))
              : null}

            {filteredRepos && filteredRepos.length > 0 ? (
              filteredRepos.map((repo) => (
                <div
                  key={repo.id}
                  className="flex flex-col gap-2 py-2 px-4 border rounded-md"
                >
                  <Text
                    flex
                    className="gap-4 flex-wrap"
                    variant="base"
                    family="mono"
                    weight="strong"
                  >
                    <Link href={repo.html_url} isExternal>
                      {repo.full_name}
                    </Link>

                    {repo?.private || repo?.fork ? (
                      <Badge className="!pl-1.5" size="sm" variant="code">
                        {repo.private && (
                          <>
                            <Icon variant="Lock" size="12" /> private
                          </>
                        )}
                        {repo.fork && (
                          <>
                            <Icon variant="GitFork" size="12" /> fork
                          </>
                        )}
                      </Badge>
                    ) : null}

                    <Badge className="!pl-1.5" size="sm" variant="code">
                      <Icon variant="GitBranchIcon" size="12" />
                      {repo?.default_branch}
                    </Badge>

                    <Text
                      flex
                      className="gap-1"
                      variant="subtext"
                      theme="neutral"
                    >
                      Updated{' '}
                      <Time
                        variant="subtext"
                        time={repo?.updated_at}
                        format="relative"
                      />
                    </Text>
                  </Text>
                  {repo.description && (
                    <Text variant="subtext" theme="neutral">
                      {repo.description}
                    </Text>
                  )}
                </div>
              ))
            ) : (
              <EmptyState
                variant="search"
                size="sm"
                emptyTitle={isFiltering ? 'No results' : 'No repositories'}
                emptyMessage={isFiltering ? 'Try adjusting your search or filters.' : 'No repositories found for this connection.'}
              />
            )}
          </div>
        </>
      )}
    </div>
  )
}
