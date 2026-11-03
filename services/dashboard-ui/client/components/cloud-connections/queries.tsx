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

export const useCloudConnections = (
  orgId: string,
  options?: { enabled?: boolean }
) =>
  useQuery({
    queryKey: ['cloud-connections', orgId, 'all'],
    enabled: !!orgId && (options?.enabled ?? true),
    placeholderData: keepPreviousData,
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

type TPrincipalMatch = Pick<
  TCloudConnectionSummary,
  'platform' | 'principal' | 'azure_tenant_id'
>

export const samePrincipal = (a: TPrincipalMatch, b: TPrincipalMatch) =>
  a.platform === b.platform &&
  (a.platform === 'azure'
    ? a.principal?.toLowerCase() === b.principal?.toLowerCase() &&
      a.azure_tenant_id?.toLowerCase() === b.azure_tenant_id?.toLowerCase()
    : a.principal === b.principal)

export const sharedPrincipalConnections = (
  connections: TCloudConnectionSummary[],
  connection: TPrincipalMatch & { id?: string }
) =>
  connections.filter(
    (other) => other.id !== connection.id && samePrincipal(other, connection)
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
