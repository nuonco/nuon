import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallAppConfigVersions } from '@/lib'
import { branchRunForConfig, branchRunHref } from './branch-run-for-config'

export const useInstallConfigRun = (appConfigId?: string) => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data, isLoading } = useQuery({
    queryKey: ['install-app-config-versions', org?.id, install?.id],
    queryFn: () =>
      getInstallAppConfigVersions({
        orgId: org!.id,
        installId: install!.id,
      }),
    enabled: !!org?.id && !!install?.id && !!appConfigId,
  })

  const run = branchRunForConfig(data, appConfigId)

  return {
    isLoading,
    run,
    runHref: branchRunHref({
      orgId: org?.id,
      appId: install?.app_id,
      run,
    }),
  }
}
