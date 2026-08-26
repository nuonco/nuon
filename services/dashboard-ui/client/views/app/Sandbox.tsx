import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { AppSandbox as SandboxConfig } from '@/components/apps/config/AppSandbox'
import { BuildSandboxButton } from '@/components/sandbox/management/BuildSandbox'
import { SandboxBuildTimeline } from '@/components/sandbox/builds/SandboxBuildTimeline'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Text } from '@/components/common/Text'
import { DetailPage } from '@/components/layout/DetailPage'
import {
  HistoryPanelButton,
  HistoryRail,
} from '@/components/layout/HistoryRail'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig, getAppConfigs } from '@/lib'

export const Sandbox = () => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: configs } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: app.id, limit: 1 }),
    enabled: !!org?.id && !!app?.id,
  })

  const appConfigId = configs?.at(0)?.id

  const { data: appConfig, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-config', org?.id, app?.id, appConfigId, 'recurse'],
    queryFn: () =>
      getAppConfig({ orgId: org.id, appId: app.id, appConfigId, recurse: true }),
    enabled: !!org?.id && !!app?.id && !!appConfigId,
  })

  const history = <SandboxBuildTimeline shouldPoll />

  return (
    <>
      <PageTitle segments={['Sandbox', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/sandbox`, text: 'Sandbox' },
        ]}
      />

      <DetailPage
        header={
          <SectionHeader
            title="Sandbox"
            description="Test builds in an isolated environment before deploying to installs."
            actions={
              <>
                <HistoryPanelButton title="Build history" history={history} />
                <BuildSandboxButton />
              </>
            }
          />
        }
      >
        <HistoryRail title="Build history" history={history}>
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
        </HistoryRail>
      </DetailPage>
    </>
  )
}