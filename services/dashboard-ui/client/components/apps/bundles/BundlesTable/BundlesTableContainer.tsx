import { useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppReleases } from '@/lib'
import { BundlesTable } from './BundlesTable'

const PENDING_STATUSES = ['preparing', 'queued', 'publishing']

export const BundlesTableContainer = () => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: result, isLoading } = useQuery({
    queryKey: ['app-releases', org?.id, app?.id],
    queryFn: () => getAppReleases({ orgId: org!.id, appId: app!.id }),
    enabled: !!org?.id && !!app?.id,
    refetchInterval: (query) =>
      query.state.data?.data?.some(
        (release) =>
          PENDING_STATUSES.includes(release?.status ?? '') ||
          release.packages?.some((pkg) =>
            PENDING_STATUSES.includes(pkg.status ?? '')
          )
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
