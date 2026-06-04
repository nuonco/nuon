import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallWorkflows } from '@/lib'
import { DeprovisionBanner } from './DeprovisionBanner'

const BANNER_STATUSES = ['provisioning', 'deprovisioning', 'deprovisioned']

export const DeprovisionBannerContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const lifecycleStatus = install?.lifecycle_status?.status
  const showBanner = !!lifecycleStatus && BANNER_STATUSES.includes(lifecycleStatus)

  const { data: workflows } = useQuery({
    queryKey: ['install-workflows', org?.id, install?.id, 'lifecycle-banner'],
    queryFn: () =>
      getInstallWorkflows({ installId: install!.id, orgId: org!.id }),
    enabled:
      !!org?.id &&
      !!install?.id &&
      (lifecycleStatus === 'deprovisioning' ||
        lifecycleStatus === 'provisioning'),
  })

  if (!showBanner) return null

  const activeWorkflow = workflows?.data?.find((w) => !w.finished_at)

  return (
    <DeprovisionBanner
      install={install}
      orgId={org?.id ?? ''}
      workflowId={activeWorkflow?.id}
    />
  )
}
