import { useQueries } from '@tanstack/react-query'
import { Status } from '@/components/common/Status'
import { useOrg } from '@/hooks/use-org'
import { checkVCSConnectionStatus } from '@/lib'
import type { TTheme } from '@/types'
import { getStatusTheme } from '@/utils/vcs-connection-utils'
import { VCSConnectionsStatusIndicator } from './VCSConnectionsStatusIndicator'

export const VCSConnectionsStatusIndicatorContainer = () => {
  const { org } = useOrg()
  const connections = org?.vcs_connections ?? []

  const statusQueries = useQueries({
    queries: connections.map((conn) => ({
      queryKey: ['vcs-connection-status', org?.id, conn.id],
      queryFn: () => checkVCSConnectionStatus({ orgId: org!.id, connectionId: conn.id }),
      enabled: !!org?.id && !!conn.id,
      refetchInterval: 60_000,
    })),
  })

  if (!connections.length) return null

  const items = connections.map((conn, i) => {
    const status = statusQueries[i]?.data
    const isLoading = statusQueries[i]?.isLoading
    const itemTheme = isLoading ? 'neutral' : getStatusTheme(status?.status)
    const accountName = conn.github_account_name || conn.github_account_id || 'GitHub'

    return {
      id: conn.id,
      href: `/${org.id}/connections/vcs/${conn.id}`,
      title: accountName,
      subtitle: status?.status ?? undefined,
      leftContent: (
        <Status status={itemTheme} isWithoutText variant="timeline" iconSize={16} />
      ),
    }
  })

  const statuses = statusQueries.map((q) => q.data?.status)
  const isLoading = statusQueries.some((q) => q.isLoading)
  const theme: TTheme = (() => {
    if (isLoading) return 'neutral'
    if (statuses.some((s) => s === 'suspended')) return 'error'
    if (statuses.some((s) => s === 'unknown' || s === undefined)) return 'warn'
    if (statuses.every((s) => s === 'active')) return 'success'
    return 'neutral'
  })()

  return <VCSConnectionsStatusIndicator items={items} theme={theme} />
}
