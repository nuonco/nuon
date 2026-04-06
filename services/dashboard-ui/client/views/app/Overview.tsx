import { useQuery } from '@tanstack/react-query'
import { AppInputs } from '@/components/apps/config/AppInputs'
import { AppRunner } from '@/components/apps/config/AppRunner'
import { AppSandbox } from '@/components/apps/config/AppSandbox'
import { AppStack } from '@/components/apps/config/AppStack'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { PropertyGridSkeleton } from '@/components/common/PropertyGrid'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig, getAppConfigs } from '@/lib'

export const Overview = () => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: configs, isLoading: isLoadingConfigs } = useQuery({
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: app.id, limit: 1 }),
    enabled: !!org?.id && !!app?.id,
  })

  const appConfigId = configs?.at(0)?.id

  const { data: appConfig, isLoading: isLoadingConfig } = useQuery({
    queryKey: ['app-config', org?.id, app?.id, appConfigId, 'recurse'],
    queryFn: () =>
      getAppConfig({ orgId: org.id, appId: app.id, appConfigId, recurse: true }),
    enabled: !!org?.id && !!app?.id && !!appConfigId,
  })

  const isLoading = isLoadingConfigs || isLoadingConfig

  return (
    <>
      <PageTitle title={`Configuration | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
        ]}
      />

      <PageSection>
        <div className="flex flex-col gap-4">
          <Text variant="h3" weight="strong">
            Inputs config
          </Text>
          {isLoading ? (
            <InputsSkeleton />
          ) : appConfig?.input?.input_groups?.length ? (
            <AppInputs appConfig={appConfig} />
          ) : (
            <EmptyState
              variant="diagram"
              emptyTitle="No app inputs configured"
              emptyMessage="Configure app inputs in your application configuration to see them here."
            />
          )}
        </div>

        {isLoading ? (
          <SandboxSkeleton />
        ) : appConfig?.sandbox ? (
          <Card className="h-fit flex flex-col gap-4">
            <Text weight="strong">Sandbox config</Text>
            <AppSandbox appConfig={appConfig} />
          </Card>
        ) : (
          <Card className="h-full">
            <EmptyState
              variant="diagram"
              emptyTitle="No sandbox configuration"
              emptyMessage="Configure a sandbox in your application configuration to see it here."
            />
          </Card>
        )}

        <div className="@container">
          <div className="grid grid-cols-1 @lg:grid-cols-5 gap-6">
            <div className="@lg:col-span-2">
              {isLoading ? (
                <RunnerSkeleton />
              ) : appConfig?.runner ? (
                <Card className="h-fit flex flex-col gap-4">
                  <Text weight="strong">Runner config</Text>
                  <AppRunner appConfig={appConfig} />
                </Card>
              ) : (
                <Card className="h-full">
                  <EmptyState
                    variant="diagram"
                    emptyTitle="No runner configuration"
                    emptyMessage="Configure a runner in your application configuration to see it here."
                  />
                </Card>
              )}
            </div>
            <div className="@lg:col-span-3">
              {isLoading ? (
                <StackSkeleton />
              ) : appConfig?.stack ? (
                <Card className="h-fit flex flex-col gap-4">
                  <Text weight="strong">Stack config</Text>
                  <AppStack appConfig={appConfig} />
                </Card>
              ) : (
                <Card className="h-full">
                  <EmptyState
                    variant="diagram"
                    emptyTitle="No stack configuration"
                    emptyMessage="Configure a stack in your application configuration to see it here."
                  />
                </Card>
              )}
            </div>
          </div>
        </div>

      </PageSection>
    </>
  )
}

const InputsSkeleton = () => (
  <div className="flex flex-col gap-6">
    <div className="flex items-center gap-3">
      <Skeleton height="28px" width="120px" />
    </div>
    {Array.from({ length: 2 }).map((_, i) => (
      <div key={i} className="border rounded-md">
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

const SandboxSkeleton = () => (
  <Card className="h-full flex flex-col gap-4">
    <Skeleton height="20px" width="140px" />
    <div className="grid gap-6" style={{ gridTemplateColumns: '2fr 1fr 1fr 1fr' }}>
      {Array.from({ length: 4 }).map((_, i) => (
        <div key={i} className="flex flex-col gap-1">
          <Skeleton height="14px" width="80px" />
          <Skeleton height="16px" width="100%" />
        </div>
      ))}
    </div>
  </Card>
)

const RunnerSkeleton = () => (
  <Card className="h-full flex flex-col gap-4">
    <Skeleton height="20px" width="120px" />
    <div className="flex flex-col gap-2">
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="flex flex-col gap-1">
          <Skeleton height="14px" width="100px" />
          <Skeleton height="16px" width="150px" />
        </div>
      ))}
    </div>
  </Card>
)

const StackSkeleton = () => (
  <Card className="h-full flex flex-col gap-4">
    <Skeleton height="20px" width="120px" />
    <div className="flex flex-col gap-2">
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="flex flex-col gap-1">
          <Skeleton height="14px" width="100px" />
          <Skeleton height="16px" width="150px" />
        </div>
      ))}
    </div>
  </Card>
)
