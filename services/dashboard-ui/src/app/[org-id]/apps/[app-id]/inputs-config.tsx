import { AppInputs as Inputs } from '@/components/apps/config/AppInputs'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { PropertyGridSkeleton } from '@/components/common/PropertyGrid'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { getAppConfig } from '@/lib'

export async function AppInputs({
  appConfigId,
  appId,
  orgId,
}: {
  appConfigId?: string
  appId: string
  orgId: string
}) {
  if (!appConfigId) {
    return <AppInputsError />
  }

  const { data: config, error } = await getAppConfig({
    appConfigId,
    appId,
    orgId,
    recurse: true,
  })

  return !error && config?.input && config?.input?.input_groups?.length ? (
    <div className="flex flex-col gap-4">
      <Text variant="h3" weight="strong">
        Inputs config
      </Text>

      <Inputs appConfig={config} />
    </div>
  ) : (
    <AppInputsError />
  )
}

export const AppInputsError = () => (
  <EmptyState
    variant="diagram"
    emptyTitle="No app inputs configured"
    emptyMessage="Configure app inputs in your application configuration to see them here."
  />
)

export const AppInputsSkeleton = () => (
  <div className="flex flex-col gap-6">
    <div className="flex items-center gap-3">
      <Skeleton height="28px" width="120px" />
    </div>

    {/* Skeleton for input groups */}
    {Array.from({ length: 2 }).map((_, groupIndex) => (
      <div key={groupIndex} className="border rounded-md">
        <div className="px-4 py-3 border-b">
          <div className="flex flex-col gap-2">
            <Skeleton height="20px" width="180px" />
            <Skeleton height="16px" width="240px" />
          </div>
        </div>
        <div className="p-4 bg-code">
          <PropertyGridSkeleton count={3} columns={6} />
        </div>
      </div>
    ))}
  </div>
)
