import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getLatestInstallRoles } from '@/lib'
import { InstallRolesTable } from './InstallRolesTable'

export const InstallRolesTableContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: roles, isLoading } = useQuery({
    queryKey: ['install-roles-latest', org?.id, install?.id],
    queryFn: () =>
      getLatestInstallRoles({
        installId: install.id,
        orgId: org.id,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  return <InstallRolesTable roles={roles ?? []} isLoading={isLoading} />
}
