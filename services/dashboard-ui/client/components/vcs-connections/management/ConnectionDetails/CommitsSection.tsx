import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useOrg } from '@/hooks/use-org'
import { getVCSConnectionCommits } from '@/lib/ctl-api/vcs-connections'
import type { TVCSConnection, TVCSConnectionCommit } from '@/types'

interface ICommitsSection {
  vcs_connection: TVCSConnection
}

const sourceLabel = (source?: string) => {
  switch (source) {
    case 'webhook':
      return 'webhook'
    case 'manual':
      return 'manual'
    case 'poll':
      return 'poll'
    case 'build':
      return 'build'
    default:
      return undefined
  }
}

export const CommitsSection = ({ vcs_connection }: ICommitsSection) => {
  const { org } = useOrg()

  const { data: commits, isLoading } = useQuery({
    queryKey: ['vcs-connection-commits', org?.id, vcs_connection?.id],
    queryFn: () =>
      getVCSConnectionCommits({
        orgId: org!.id,
        connectionId: vcs_connection.id,
        limit: 25,
      }),
    enabled: !!org?.id && !!vcs_connection?.id,
  })

  return (
    <div className="flex flex-col gap-4">
      <Text variant="body" weight="strong">
        Recent commits
      </Text>

      <div className="flex flex-col gap-2">
        {isLoading
          ? Array.from({ length: 3 }).map((_, idx) => (
              <div
                key={idx}
                className="flex flex-col gap-1 py-2 px-4 border rounded-md"
              >
                <Skeleton height="20px" width="300px" />
                <Skeleton height="16px" width="200px" />
              </div>
            ))
          : null}

        {!isLoading && commits && commits.length > 0
          ? commits.map((commit: TVCSConnectionCommit) => (
              <div
                key={commit.id}
                className="flex flex-col gap-1 py-2 px-4 border rounded-md"
              >
                <div className="flex items-center gap-2 flex-wrap">
                  <Text
                    variant="base"
                    family="mono"
                    className="truncate max-w-[500px]"
                  >
                    {commit.message || commit.sha?.slice(0, 8)}
                  </Text>
                  <Badge size="sm" variant="code">
                    {commit.sha?.slice(0, 7)}
                  </Badge>
                  {commit.branch && (
                    <Badge className="!pl-1.5" size="sm" variant="code">
                      <Icon variant="GitBranch" size="12" />
                      {commit.branch}
                    </Badge>
                  )}
                  {sourceLabel(commit.source) && (
                    <Badge size="sm" variant="code">
                      {sourceLabel(commit.source)}
                    </Badge>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  {commit.repo_owner && commit.repo_name && (
                    <Text variant="subtext" theme="neutral" family="mono">
                      {commit.repo_owner}/{commit.repo_name}
                    </Text>
                  )}
                  {commit.author_name && (
                    <Text variant="subtext" theme="neutral">
                      {commit.author_name}
                    </Text>
                  )}
                  {commit.created_at && (
                    <Time
                      variant="subtext"
                      time={commit.created_at}
                      format="relative"
                    />
                  )}
                </div>
              </div>
            ))
          : null}

        {!isLoading && (!commits || commits.length === 0) ? (
          <Text variant="subtext" theme="neutral">
            No commits found for this connection.
          </Text>
        ) : null}
      </div>
    </div>
  )
}
