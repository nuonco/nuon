import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { getCloudConnection, verifyCloudConnection } from '@/lib'

export const useCloudConnection = (orgId: string, connectionId: string) =>
  useQuery({
    queryKey: ['cloud-connections', orgId, connectionId],
    queryFn: () => getCloudConnection({ orgId, connectionId }),
    enabled: !!orgId && !!connectionId,
    placeholderData: keepPreviousData,
  })

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
