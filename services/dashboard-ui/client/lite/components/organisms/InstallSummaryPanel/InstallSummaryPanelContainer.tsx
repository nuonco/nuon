import { useQuery } from '@tanstack/react-query'
import { getAppLabels, getInstall, toLabelColorMap } from '@/lib'
import { useOrg } from '../../../providers/org-provider'
import { installHref } from '../../../utils/hrefs'
import { InstallSummaryPanel } from './InstallSummaryPanel'

export const InstallSummaryPanelContainer = ({
  installId,
}: {
  installId?: string
}) => {
  const { orgId } = useOrg()
  const installQuery = useQuery({
    queryKey: ['install', orgId, installId],
    queryFn: () => getInstall({ orgId: orgId!, installId: installId! }),
    enabled: !!orgId && !!installId,
  })
  const appId = installQuery.data?.app_id
  const labelsQuery = useQuery({
    queryKey: ['app-labels', orgId, appId],
    queryFn: () => getAppLabels({ orgId: orgId!, appId: appId! }),
    enabled: !!orgId && !!appId,
    staleTime: 60_000,
  })

  return (
    <InstallSummaryPanel
      install={installQuery.data}
      installHref={
        orgId && installId ? installHref(orgId, installId) : undefined
      }
      labelColors={toLabelColorMap(labelsQuery.data)}
      loading={installQuery.isLoading}
      error={installQuery.error}
    />
  )
}
