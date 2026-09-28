import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useToast } from '@/hooks/use-toast'
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
  const { addToast } = useToast()
  return useMutation({
    mutationFn: () => verifyCloudConnection({ orgId, connectionId }),
    onSuccess: (connection) => {
      client.setQueryData(
        ['cloud-connections', orgId, connectionId],
        connection
      )
      const verified = connection.status === 'verified'
      addToast(
        <Toast
          heading={verified ? 'Connection verified' : 'Verification failed'}
          theme={verified ? 'success' : 'error'}
        >
          <Text>
            {connection.status_message ||
              (verified
                ? `${connection.name} is ready to use.`
                : `${connection.name} could not be verified.`)}
          </Text>
        </Toast>
      )
    },
    onError: (error) =>
      addToast(
        <Toast heading="Verification failed" theme="error">
          <Text>
            {error.description ||
              error.error ||
              'Unable to verify this connection.'}
          </Text>
        </Toast>
      ),
    onSettled: () =>
      client.invalidateQueries({ queryKey: ['cloud-connections', orgId] }),
  })
}
