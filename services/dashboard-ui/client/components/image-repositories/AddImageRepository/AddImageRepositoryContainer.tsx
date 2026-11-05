import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { Toast } from '@/components/surfaces/Toast'
import { useCloudConnections } from '@/components/cloud-connections/queries'
import { createCloudConnection, createImageRepository } from '@/lib'
import type {
  TCloudConnectionSummary,
  TCreateCloudConnectionRequest,
  TCreateImageRepositoryRequest,
} from '@/types'
import { AddImageRepository } from './AddImageRepository'

export const AddImageRepositoryContainer = () => {
  const { org } = useOrg()
  const client = useQueryClient()
  const navigate = useNavigate()
  const { addToast } = useToast()
  const identities = useCloudConnections(org?.id)
  const createIdentity = useMutation({
    mutationFn: (body: TCreateCloudConnectionRequest) =>
      createCloudConnection({ orgId: org.id, body }),
    onSuccess: (connection) => {
      client.setQueryData<TCloudConnectionSummary[]>(
        ['cloud-connections', org.id, 'all'],
        (connections) =>
          connections && !connections.some((c) => c.id === connection.id)
            ? [...connections, connection]
            : connections
      )
      client.invalidateQueries({
        queryKey: ['cloud-connections', org?.id],
      })
    },
  })
  const save = useMutation({
    mutationFn: (body: TCreateImageRepositoryRequest) =>
      createImageRepository({ orgId: org.id, body }),
    onSuccess: (repository) => {
      client.setQueryData(
        ['image-repositories', org.id, repository.id],
        repository
      )
      client.invalidateQueries({
        queryKey: ['image-repositories', org.id],
        predicate: (query) => query.queryKey[2] !== repository.id,
      })
      addToast(
        <Toast heading="Repository added" theme="success">
          <Text>
            Added <Text family="mono">{repository.image_url_prefix}</Text>.
            Review its setup commands next.
          </Text>
        </Toast>
      )
      navigate(`/${org.id}/settings/image-repositories/${repository.id}`)
    },
  })
  return (
    <AddImageRepository
      orgId={org?.id}
      connections={identities.data ?? []}
      isLoading={identities.isLoading}
      isPending={save.isPending}
      error={save.error || identities.error}
      isCreatingIdentity={createIdentity.isPending}
      identityError={createIdentity.error}
      onCreateIdentity={createIdentity.mutateAsync}
      onSave={save.mutate}
    />
  )
}
