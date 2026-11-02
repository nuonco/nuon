import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router'
import { useCloudConnections } from '@/components/cloud-connections/queries'
import { useOrg } from '@/hooks/use-org'
import { createCloudConnection, createImageRepository } from '@/lib'
import type {
  TCreateCloudConnectionRequest,
  TCreateImageRepositoryRequest,
} from '@/types'
import { AddImageRepository } from './AddImageRepository'

export const AddImageRepositoryContainer = () => {
  const { org } = useOrg()
  const client = useQueryClient()
  const navigate = useNavigate()
  const identities = useCloudConnections(org?.id)
  const createIdentity = useMutation({
    mutationFn: (body: TCreateCloudConnectionRequest) =>
      createCloudConnection({ orgId: org.id, body }),
    onSuccess: () =>
      client.invalidateQueries({ queryKey: ['cloud-connections', org?.id] }),
  })
  const save = useMutation({
    mutationFn: (body: TCreateImageRepositoryRequest) =>
      createImageRepository({ orgId: org.id, body }),
    onSuccess: (repository) => {
      client.setQueryData(
        ['image-repositories', org.id, repository.id],
        repository
      )
      client.invalidateQueries({ queryKey: ['image-repositories', org.id] })
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
