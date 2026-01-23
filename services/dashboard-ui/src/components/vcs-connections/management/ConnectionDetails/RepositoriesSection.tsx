import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TVCSConnectionReposResponse } from '@/types'

interface IRepositoriesSection {
  repos?: TVCSConnectionReposResponse
  error?: any
  isLoading: boolean
}

export const RepositoriesSection = ({
  repos,
  error,
  isLoading,
}: IRepositoriesSection) => (
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
        {isLoading ? (
          <Skeleton height="16px" width="80px" />
        ) : repos && repos?.repositories?.length > 0 ? (
          <Text variant="subtext" theme="neutral">
            {repos.total_count}{' '}
            {repos.total_count === 1 ? 'repository' : 'repositories'}
          </Text>
        ) : null}
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

          {repos && repos?.repositories?.length > 0 ? (
            repos.repositories.map((repo) => (
              <div
                key={repo.id}
                className="flex flex-col gap-2 py-2 px-4 border rounded-md"
              >
                <Text
                  className="!flex items-center gap-4 flex-wrap"
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
                    className="!flex items-center gap-1"
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
            <Text variant="subtext" theme="neutral">
              No repositories accessible
            </Text>
          )}
        </div>
      </>
    )}
  </div>
)
