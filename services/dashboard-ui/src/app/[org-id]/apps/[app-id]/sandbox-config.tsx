import { AppSandbox as Sandbox } from '@/components/apps/config/AppSandbox'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { getAppConfig } from '@/lib'

export async function AppSandbox({
  appConfigId,
  appId,
  orgId,
}: {
  appConfigId?: string
  appId: string
  orgId: string
}) {
  if (!appConfigId) {
    return <AppSandboxError />
  }

  const { data: config, error } = await getAppConfig({
    appConfigId,
    appId,
    orgId,
    recurse: true,
  })

  return !error && config?.sandbox ? (
    <Card className="h-fit flex flex-col gap-4">
      <Text weight="strong">Sandbox config</Text>
      <Sandbox appConfig={config} />
    </Card>
  ) : (
    <AppSandboxError />
  )
}

export const AppSandboxError = () => (
  <Card className="h-full">
    <EmptyState
      variant="diagram"
      emptyTitle="No sandbox configuration"
      emptyMessage="Configure a sandbox in your application configuration to see it here."
    />
  </Card>
)

export const AppSandboxSkeleton = () => (
  <Card className="h-full flex flex-col gap-4">
    <Skeleton height="20px" width="140px" />
    <div className="grid gap-6" style={{ gridTemplateColumns: '2fr 1fr 1fr 1fr' }}>
      <div className="flex flex-col gap-1">
        <Skeleton height="14px" width="80px" />
        <Skeleton height="16px" width="100%" />
      </div>
      <div className="flex flex-col gap-1">
        <Skeleton height="14px" width="60px" />
        <Skeleton height="16px" width="80px" />
      </div>
      <div className="flex flex-col gap-1">
        <Skeleton height="14px" width="70px" />
        <Skeleton height="16px" width="90px" />
      </div>
      <div className="flex flex-col gap-1">
        <Skeleton height="14px" width="75px" />
        <Skeleton height="16px" width="60px" />
      </div>
    </div>
  </Card>
)
