import { AppRunner as Runner } from '@/components/apps/config/AppRunner'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { getAppConfig } from '@/lib'

export async function AppRunner({
  appConfigId,
  appId,
  orgId,
}: {
  appConfigId?: string
  appId: string
  orgId: string
}) {
  if (!appConfigId) {
    return <AppRunnerError />
  }

  const { data: config, error } = await getAppConfig({
    appConfigId,
    appId,
    orgId,
    recurse: true,
  })

  return !error && config?.runner ? (
    <Card className="flex-initial h-fit flex flex-col gap-4">
      <Text weight="strong">Runner config</Text>
      <Runner appConfig={config} />
    </Card>
  ) : (
    <AppRunnerError />
  )
}

export const AppRunnerError = () => (
  <Card className="flex-auto">
    <EmptyState
      variant="diagram"
      emptyTitle="No runner configuration"
      emptyMessage="Configure a runner in your application configuration to see it here."
    />
  </Card>
)

export const AppRunnerSkeleton = () => (
  <Card className="flex-auto flex flex-col gap-4">
    <Skeleton height="20px" width="120px" />
    <div className="flex flex-col gap-2">
      <div className="flex flex-col gap-1">
        <Skeleton height="14px" width="100px" />
        <Skeleton height="16px" width="150px" />
      </div>
      <div className="flex flex-col gap-1">
        <Skeleton height="14px" width="80px" />
        <Skeleton height="16px" width="120px" />
      </div>
      <div className="flex flex-col gap-1">
        <Skeleton height="14px" width="90px" />
        <Skeleton height="16px" width="100px" />
      </div>
    </div>
  </Card>
)
