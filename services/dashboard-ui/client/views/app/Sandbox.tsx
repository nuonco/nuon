import { useQuery } from '@tanstack/react-query'
import { AppSandbox as SandboxConfig } from '@/components/apps/config/AppSandbox'
import { BuildSandboxButton } from '@/components/sandbox/management/BuildSandbox'
import { SandboxBuildTimeline } from '@/components/sandbox/builds/SandboxBuildTimeline'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig, getAppConfigs } from '@/lib'

export const Sandbox = () => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: configs } = useQuery({
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: app.id, limit: 1 }),
    enabled: !!org?.id && !!app?.id,
  })

  const appConfigId = configs?.at(0)?.id

  const { data: appConfig, isLoading } = useQuery({
    queryKey: ['app-config', org?.id, app?.id, appConfigId, 'recurse'],
    queryFn: () =>
      getAppConfig({ orgId: org.id, appId: app.id, appConfigId, recurse: true }),
    enabled: !!org?.id && !!app?.id && !!appConfigId,
  })

  return (
    <PageSection>
      <PageTitle title={`Sandbox | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/sandbox`, text: 'Sandbox' },
        ]}
      />

      <div className="flex items-start justify-between">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Sandbox
          </Text>
          <Text variant="subtext" theme="neutral">
            Test builds in an isolated environment before deploying to installs.
          </Text>
        </HeadingGroup>
        <BuildSandboxButton />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto gap-6">
        <div className="md:col-span-8 flex flex-col gap-6">
          {isLoading ? (
            <Card>
              <Text>Loading...</Text>
            </Card>
          ) : appConfig?.sandbox ? (
            <Card className="flex flex-col gap-4">
              <Text weight="strong">Sandbox config</Text>
              <SandboxConfig appConfig={appConfig} />
            </Card>
          ) : (
            <EmptyState
              variant="diagram"
              emptyTitle="No sandbox configured"
              emptyMessage="Configure a sandbox in your application configuration to see it here."
            />
          )}
        </div>

        <div className="md:col-span-4 flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Build history
          </Text>
          <SandboxBuildTimeline shouldPoll />
        </div>
      </div>
    </PageSection>
  )
}