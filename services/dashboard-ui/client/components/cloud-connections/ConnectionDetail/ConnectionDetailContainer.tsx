import { useEffect, useRef } from 'react'
import { Text } from '@/components/common/Text'
import { InstallsTable } from '@/components/installs/InstallsTable'
import { Toast } from '@/components/surfaces/Toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
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
  isVerifying,
  verificationTimedOut,
}: {
  connection?: TCloudConnection
  orgId: string
  connectionId: string
  tab: TConnectionTab
  error?: TAPIError | null
  isVerifying: boolean
  verificationTimedOut: boolean
}) => {
  const { addModal } = useSurfaces()
  const { addToast } = useToast()
  const verify = useVerifyCloudConnection(orgId, connectionId)
  const notifiedRequest = useRef<string>()
  useEffect(() => {
    const requestedAt = verify.data?.verification_requested_at
    if (
      !requestedAt ||
      notifiedRequest.current === requestedAt ||
      connection?.verification_requested_at !== requestedAt ||
      connection.verification_in_progress
    )
      return
    notifiedRequest.current = requestedAt
    const verified = connection.status === 'verified'
    addToast(
      <Toast
        heading={verified ? 'Connection verified' : 'Verification failed'}
        theme={verified ? 'success' : 'error'}
      >
        <Text>{connection.status_message}</Text>
      </Toast>
    )
  }, [connection, verify.data, addToast])
  return (
    <ConnectionDetail
      connection={connection}
      tab={tab}
      basePath={`/${orgId}/settings/cloud-connections/${connectionId}`}
      error={error || verify.error}
      isVerifying={verify.isPending || isVerifying}
      verificationTimedOut={verificationTimedOut}
      onVerify={() =>
        verify.mutate(undefined, {
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
