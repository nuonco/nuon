import { useParams } from 'react-router'
import { useApp } from '@/hooks/use-app'
import { useBranchScopedAppConfig } from '@/hooks/use-branch-scoped-app-config'
import { useOrg } from '@/hooks/use-org'
import type { TNamedIAMPolicy } from '@/lib/ctl-api/installs/get-install-app-permissions-config'
import { AppRolesTable } from './AppRolesTable'

export const AppRolesTableContainer = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branchId } = useParams()
  const { appConfig, isLoading } = useBranchScopedAppConfig({
    orgId: org?.id,
    appId: app?.id,
    branchId,
  })

  const permissions = appConfig?.permissions as
    | { named_policies?: TNamedIAMPolicy[] }
    | undefined

  return (
    <AppRolesTable
      roles={appConfig?.permissions?.aws_iam_roles ?? []}
      namedPolicies={permissions?.named_policies ?? []}
      isLoading={isLoading}
    />
  )
}
