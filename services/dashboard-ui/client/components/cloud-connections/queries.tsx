import { useEffect, useState } from 'react'
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { getCloudConnection, verifyCloudConnection } from '@/lib'

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
