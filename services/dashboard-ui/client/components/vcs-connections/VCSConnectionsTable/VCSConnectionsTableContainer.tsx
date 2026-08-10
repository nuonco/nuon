import { useQueries } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { checkVCSConnectionStatus } from '@/lib'
import {
  VCSConnectionsTable,
  type TVCSConnectionRow,
} from './VCSConnectionsTable'

export const VCSConnectionsTableContainer = () => {
  const { org } = useOrg()
  const connections = org?.vcs_connections ?? []

  const statusQueries = useQueries({
    queries: connections.map((conn) => ({
      queryKey: ['vcs-connection-status', org?.id, conn.id],
      queryFn: () =>
        checkVCSConnectionStatus({ orgId: org!.id, connectionId: conn.id! }),
      enabled: !!org?.id && !!conn.id,
      refetchInterval: 60_000,
    })),
  })

  const rows: TVCSConnectionRow[] = connections.map((connection, i) => ({
    connection,
    href: `/${org?.id}/settings/vcs/${connection.id}`,
    status: statusQueries[i]?.data?.status,
    checkedAt: statusQueries[i]?.data?.checked_at,
    isLoadingStatus: statusQueries[i]?.isLoading,
  }))

  return <VCSConnectionsTable data={rows} />
}
