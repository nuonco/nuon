import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useStoredRecord } from '@/hooks/use-stored-record'
import { getInstallWorkflows } from '@/lib'
import { DeprovisionBanner } from './DeprovisionBanner'

const BANNER_STATUSES = ['provisioning', 'deprovisioning', 'deprovisioned']
const DISMISSED_STORAGE_KEY = 'nuon:dismissed-lifecycle-banners'

export const DeprovisionBannerContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [dismissed, setDismissed] = useStoredRecord<boolean>(
    DISMISSED_STORAGE_KEY
  )

  const lifecycleStatus = install?.lifecycle_phase?.phase
  const dismissKey = `${install?.id}:${lifecycleStatus}`
  const showBanner =
    !!lifecycleStatus &&
    BANNER_STATUSES.includes(lifecycleStatus) &&
    !dismissed[dismissKey]

  const { data: workflows } = useQuery({
    placeholderData: keepPreviousData,
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
      onDismiss={() => setDismissed(dismissKey, true)}
    />
  )
}
