import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig, getLatestInstallRoles } from '@/lib'
import type { TNamedIAMPolicy } from '@/lib/ctl-api/installs/get-install-app-permissions-config'
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

  const { data: appConfig } = useQuery({
    queryKey: [
      'app-config',
      org?.id,
      install?.app_id,
      install?.app_config_id,
      'roles-named-policies',
    ],
    queryFn: () =>
      getAppConfig({
        appId: install.app_id,
        appConfigId: install.app_config_id,
        orgId: org.id,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_id && !!install?.app_config_id,
  })

  const permissions = appConfig?.permissions as
    | { named_policies?: TNamedIAMPolicy[] }
    | undefined

  return (
    <InstallRolesTable
      roles={result?.data ?? []}
      namedPolicies={permissions?.named_policies ?? []}
      isLoading={isLoading}
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
    />
  )
}
