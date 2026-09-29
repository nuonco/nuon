import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallUpdates } from '@/lib'
import type { TAppBranchRun } from '@/types'

export const useCurrentAppBranchRun = (): {
  run: TAppBranchRun | undefined
  isLoading: boolean
} => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data, isLoading } = useQuery({
    queryKey: ['install-current-app-branch-run', org?.id, install?.id],
    queryFn: () =>
      getInstallUpdates({
        orgId: org!.id,
        installId: install!.id,
        page: 0,
        limit: 1,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  return { run: data?.current_app_branch_run, isLoading }
}
