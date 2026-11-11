import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  sharedPrincipalConnections,
  useCloudConnection,
  useCloudConnections,
  useVerifyCloudConnection,
} from '@/components/cloud-connections/queries'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getCloudConnectionSetup } from '@/lib'
import { useImageRepository } from '../queries'
import { ImageRepository } from './ImageRepository'
import { RemoveImageRepositoryContainer } from './RemoveImageRepositoryContainer'

export const ImageRepositoryContainer = ({
  repositoryId,
}: {
  repositoryId: string
}) => {
  const { org } = useOrg()
  const { addModal } = useSurfaces()
  const query = useImageRepository(org?.id, repositoryId)
  const connectionId = query.data?.cloud_connection_id || ''
  const connection = useCloudConnection(org?.id, connectionId)
  const identities = useCloudConnections(org?.id)
  const shared = query.data
    ? sharedPrincipalConnections(
        identities.data || [],
        query.data.cloud_connection
      )
    : []
  const [setupMode, setSetupMode] = useState<boolean>()
  // Azure hides the setup-mode selector, so it never uses an existing principal.
  const existingPrincipal =
    query.data?.platform === 'azure'
      ? false
      : (setupMode ??
        (shared.length > 0 || connection.data?.status === 'verified'))
  const verify = useVerifyCloudConnection(org?.id, connectionId)
  const setup = useQuery({
    queryKey: [
      'image-repository-setup',
      org?.id,
      repositoryId,
      connectionId,
      query.data?.repository,
      existingPrincipal,
    ],
    queryFn: () =>
      getCloudConnectionSetup({
        orgId: org.id,
        connectionId,
        repositories: [query.data!.repository],
        existingPrincipal,
      }),
    enabled: !!org?.id && !!connectionId && !!query.data,
  })
  return (
    <ImageRepository
      orgId={org?.id}
      repository={query.data}
      connection={connection.data}
      setup={setup.data}
      sharedConnections={shared}
      existingPrincipal={existingPrincipal}
      onSetupModeChange={setSetupMode}
      error={query.error || connection.error}
      verifyError={verify.error}
      setupError={setup.error || identities.error}
      isVerifying={verify.isPending || connection.isVerifying}
      verificationTimedOut={connection.verificationTimedOut}
      onVerify={() => verify.mutate()}
      onReloadSetup={() => setup.refetch()}
      onRemove={() =>
        query.data &&
        addModal(<RemoveImageRepositoryContainer repository={query.data} />)
      }
    />
  )
}
