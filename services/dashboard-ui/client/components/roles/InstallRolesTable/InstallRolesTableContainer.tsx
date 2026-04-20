import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getLatestInstallRoles } from '@/lib'
import { InstallRolesTable } from './InstallRolesTable'

const LIMIT = 10

export const InstallRolesTableContainer = () => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const offset = Number(searchParams.get('offset') ?? 0)
  const q = searchParams.get('q') || undefined

  const { data: result, isLoading } = useQuery({
    queryKey: ['install-roles-latest', org?.id, install?.id, offset, q],
    queryFn: () =>
      getLatestInstallRoles({
        installId: install.id,
        orgId: org.id,
        offset,
        limit: LIMIT,
        q,
      }),
    enabled: !!org?.id && !!install?.id,
    placeholderData: keepPreviousData,
  })

  return (
    <InstallRolesTable
      roles={result?.data ?? []}
      isLoading={isLoading}
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
    />
  )
}
