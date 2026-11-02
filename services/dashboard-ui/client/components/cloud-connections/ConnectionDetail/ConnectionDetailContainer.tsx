import { useEffect, useRef } from 'react'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { InstallsTable } from '@/components/installs/InstallsTable'
import { Toast } from '@/components/surfaces/Toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import type { TAPIError, TCloudConnection } from '@/types'
import { DeleteConnection } from '../DeleteConnection'
import {
  sharedPrincipalConnections,
  useCloudConnections,
  useVerifyCloudConnection,
} from '../queries'
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
  const identities = useCloudConnections(orgId)
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
      sharedConnections={
        connection
          ? sharedPrincipalConnections(identities.data || [], connection)
          : []
      }
      tab={tab}
      basePath={`/${orgId}/settings/cloud-connections/${connectionId}`}
      error={error || verify.error || identities.error}
      isVerifying={verify.isPending || isVerifying}
      verificationTimedOut={verificationTimedOut}
      repositoryAccess={
        connection && ['ecr', 'gar', 'acr'].includes(connection.preset) ? (
          <section
            className="flex flex-col gap-4 border-t pt-6"
            aria-label="Image repositories"
          >
            <SectionHeader
              title="Image repositories"
              description="Save each repository and explicitly choose its pull identity in Image repositories."
            />
            <Button
              className="w-fit"
              href={`/${orgId}/settings/image-repositories`}
            >
              View image repositories
            </Button>
          </section>
        ) : undefined
      }
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
