import { useQuery } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { InstallsTable } from '@/components/installs/InstallsTable'
import { Toast } from '@/components/surfaces/Toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { getCloudConnectionSetup } from '@/lib'
import type { TAPIError, TCloudConnection } from '@/types'
import { DeleteConnection } from '../DeleteConnection'
import { useVerifyCloudConnection } from '../queries'
import { ConnectionDetail, type TConnectionTab } from './ConnectionDetail'

export const ConnectionDetailContainer = ({
  connection,
  orgId,
  connectionId,
  tab,
  error,
}: {
  connection?: TCloudConnection
  orgId: string
  connectionId: string
  tab: TConnectionTab
  error?: TAPIError | null
}) => {
  const { addModal } = useSurfaces()
  const { addToast } = useToast()
  const verify = useVerifyCloudConnection(orgId, connectionId)
  const setup = useQuery({
    queryKey: ['cloud-connections', orgId, connectionId, 'setup'],
    queryFn: () => getCloudConnectionSetup({ orgId, connectionId }),
    enabled: !!connection && tab === 'overview',
  })
  return (
    <ConnectionDetail
      connection={connection}
      tab={tab}
      basePath={`/${orgId}/cloud-connections/${connectionId}`}
      error={error || verify.error}
      setup={setup.data}
      setupError={setup.error}
      isVerifying={verify.isPending}
      onVerify={() =>
        verify.mutate(undefined, {
          onSuccess: (connection) => {
            const verified = connection.status === 'verified'
            addToast(
              <Toast
                heading={
                  verified ? 'Connection verified' : 'Verification failed'
                }
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
        })
      }
      onDelete={() =>
        connection && addModal(<DeleteConnection connection={connection} />)
      }
      installs={
        tab === 'installs' ? (
          <InstallsTable
            cloudConnectionId={connectionId}
            emptyTitle="No installs use this connection yet"
            emptyMessage="Select this connection when creating an install."
          />
        ) : null
      }
    />
  )
}
