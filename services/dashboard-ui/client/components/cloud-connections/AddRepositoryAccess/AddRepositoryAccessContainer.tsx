import { useMutation } from '@tanstack/react-query'
import { getCloudConnectionSetup } from '@/lib'
import type { TCloudConnection } from '@/types'
import { canonicalRepositories } from '../repository-scope'
import { AddRepositoryAccess } from './AddRepositoryAccess'

export const AddRepositoryAccessContainer = ({
  connection,
  orgId,
}: {
  connection: TCloudConnection
  orgId: string
}) => {
  const generate = useMutation({
    mutationFn: async (names: string[]) => {
      const setup = await getCloudConnectionSetup({
        orgId,
        connectionId: connection.id,
        repositories: canonicalRepositories(connection, names),
      })
      if (!setup.permissions_cli?.trim() || !setup.permissions_policy) {
        throw {
          error: 'Permission command generation failed',
          description:
            'No repository permission commands were returned. Try generating them again.',
          user_error: false,
        }
      }
      return {
        names,
        permissionsCli: setup.permissions_cli,
        permissionsPolicy: setup.permissions_policy,
      }
    },
  })
  return (
    <AddRepositoryAccess
      connection={connection}
      onGenerate={(names) => generate.mutate(names)}
      isPending={generate.isPending}
      error={generate.error}
      result={generate.isSuccess ? generate.data : undefined}
    />
  )
}
