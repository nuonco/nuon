import { useEffect, useState } from 'react'
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import {
  getCloudConnection,
  getCloudConnections,
  verifyCloudConnection,
} from '@/lib'
import type { TCloudConnectionSummary } from '@/types'

export const useCloudConnections = (orgId: string) =>
  useQuery({
    queryKey: ['cloud-connections', orgId, 'all'],
    enabled: !!orgId,
    queryFn: async () => {
      const connections: TCloudConnectionSummary[] = []
      let offset = 0
      let hasNext = true
      while (hasNext) {
        const page = await getCloudConnections({ orgId, limit: 100, offset })
        connections.push(...page.data)
        hasNext = page.pagination.hasNext
        offset += page.pagination.limit
      }
      return connections
    },
  })

export const sharedPrincipalConnections = (
  connections: TCloudConnectionSummary[],
  connection: Pick<
    TCloudConnectionSummary,
    'platform' | 'principal' | 'azure_tenant_id'
  > & {
    id?: string
  }
) =>
  connections.filter(
    (other) =>
      other.id !== connection.id &&
      other.platform === connection.platform &&
      (connection.platform === 'azure'
        ? other.principal?.toLowerCase() ===
            connection.principal?.toLowerCase() &&
          other.azure_tenant_id?.toLowerCase() ===
            connection.azure_tenant_id?.toLowerCase()
        : other.principal === connection.principal)
  )

export const useCloudConnection = (orgId: string, connectionId: string) => {
  const [expiredRequest, setExpiredRequest] = useState<string>()
  const query = useQuery({
    queryKey: ['cloud-connections', orgId, connectionId],
    queryFn: () => getCloudConnection({ orgId, connectionId }),
    enabled: !!orgId && !!connectionId,
    placeholderData: keepPreviousData,
    refetchInterval: ({ state }) =>
      state.data?.verification_in_progress &&
      state.data.verification_requested_at !== expiredRequest
        ? 2000
        : false,
  })
  const requestedAt = query.data?.verification_requested_at
  const inProgress = !!query.data?.verification_in_progress
  useEffect(() => {
    if (!inProgress || !requestedAt) return
    const timer = setTimeout(() => setExpiredRequest(requestedAt), 120000)
    return () => clearTimeout(timer)
  }, [requestedAt, inProgress])
  const verificationTimedOut = inProgress && requestedAt === expiredRequest
  return {
    ...query,
    isVerifying: inProgress && !verificationTimedOut,
    verificationTimedOut,
  }
}

export const useVerifyCloudConnection = (
  orgId: string,
  connectionId: string
) => {
  const client = useQueryClient()
  return useMutation({
    mutationFn: () => verifyCloudConnection({ orgId, connectionId }),
    onSuccess: (connection) => {
      client.setQueryData(
        ['cloud-connections', orgId, connectionId],
        connection
      )
    },
    onSettled: () =>
      client.invalidateQueries({ queryKey: ['cloud-connections', orgId] }),
  })
}
