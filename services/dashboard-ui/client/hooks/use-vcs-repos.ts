import { keepPreviousData, useQueries, useQuery } from '@tanstack/react-query'
import { getVCSConnectionRepos, getVCSConnections } from '@/lib'
import { useOrg } from '@/hooks/use-org'
import type { TVCSConnectionRepo } from '@/types'

export function useVCSRepos({ enabled = true }: { enabled?: boolean } = {}) {
  const { org } = useOrg()

  const { data: connections, isLoading: isLoadingConnections } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['vcs-connections', org?.id],
    queryFn: () => getVCSConnections({ orgId: org!.id }),
    enabled: enabled && !!org?.id,
  })

  const repoQueries = useQueries({
    queries: (connections ?? []).map((connection) => ({
      queryKey: ['vcs-connection-repos', org?.id, connection.id],
      queryFn: () =>
        getVCSConnectionRepos({ orgId: org!.id, connectionId: connection.id! }),
      enabled: enabled && !!org?.id && !!connection.id,
    })),
  })

  const byFullName = new Map<string, TVCSConnectionRepo>()
  for (const query of repoQueries) {
    for (const repo of query.data?.repositories ?? []) {
      if (repo.full_name && !byFullName.has(repo.full_name)) {
        byFullName.set(repo.full_name, repo)
      }
    }
  }

  const repos = [...byFullName.values()].sort((a, b) =>
    a.full_name.localeCompare(b.full_name)
  )

  return {
    repos,
    isLoading:
      isLoadingConnections || repoQueries.some((query) => query.isLoading),
    hasConnections: !!connections?.length,
  }
}
