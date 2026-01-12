import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { Markdown } from '@/components/common/Markdown'
import { getAppConfig } from '@/lib'

export async function Readme({
  appConfigId,
  appId,
  orgId,
}: {
  appConfigId?: string
  appId: string
  orgId: string
}) {
  if (!appConfigId) {
    return <ReadmeError />
  }

  const { data: config, error } = await getAppConfig({
    appConfigId,
    appId,
    orgId,
  })

  return !error && config?.readme ? (
    <Markdown content={config.readme} />
  ) : (
    <ReadmeError />
  )
}

export const ReadmeError = () => (
  <EmptyState
    variant="table"
    emptyTitle="No README in app config"
    emptyMessage="You can add a README for your app in your app config TOML file."
  />
)

export const ReadmeSkeleton = () => (
  <div className="space-y-4">
    <Skeleton height="24px" />
    <Skeleton height="16px" />
    <Skeleton height="16px" />
    <Skeleton height="20px" />
    <Skeleton height="16px" />
    <Skeleton height="16px" />
    <Skeleton height="16px" />
  </div>
)
