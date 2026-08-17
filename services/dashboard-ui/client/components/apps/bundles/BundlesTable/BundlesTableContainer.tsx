import { useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAirgapBundles } from '@/lib'
import { BundlesTable } from './BundlesTable'

const PENDING_STATUSES = ['queued', 'publishing']

export const BundlesTableContainer = () => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: result, isLoading } = useQuery({
    queryKey: ['airgap-bundles', org?.id, app?.id],
    queryFn: () => getAirgapBundles({ orgId: org!.id, appId: app!.id }),
    enabled: !!org?.id && !!app?.id,
    refetchInterval: (query) =>
      query.state.data?.data?.some((bundle) =>
        PENDING_STATUSES.includes(bundle?.status ?? '')
      )
        ? 5000
        : 30000,
  })

  return (
    <BundlesTable
      data={result?.data ?? []}
      isLoading={isLoading}
      orgId={org?.id}
      appId={app?.id}
    />
  )
}
