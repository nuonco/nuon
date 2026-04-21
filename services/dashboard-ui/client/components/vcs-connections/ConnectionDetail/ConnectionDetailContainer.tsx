import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { useOrg } from '@/hooks/use-org'
import { getVCSConnectionRepos } from '@/lib'
import type { TVCSConnection } from '@/types'
import { ConnectionDetail } from './ConnectionDetail'

interface IConnectionDetailContainer {
  vcs_connection: TVCSConnection
}

export const ConnectionDetailContainer = ({ vcs_connection }: IConnectionDetailContainer) => {
  const { connectionId } = useParams<{ connectionId: string }>()
  const { org } = useOrg()

  const {
    data: repos,
    error: reposError,
    isLoading: isLoadingRepos,
  } = useQuery({
    queryKey: ['vcs-connection-repos', org?.id, connectionId],
    queryFn: () => getVCSConnectionRepos({ orgId: org!.id, connectionId: connectionId! }),
    enabled: !!org?.id && !!connectionId,
  })

  return (
    <ConnectionDetail
      vcs_connection={vcs_connection}
      repos={repos}
      reposError={reposError}
      isLoadingRepos={isLoadingRepos}
    />
  )
}
