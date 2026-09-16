import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { AppSandbox as SandboxConfig } from '@/components/apps/config/AppSandbox'
import { BuildSandboxButton } from '@/components/sandbox/management/BuildSandbox'
import { CurrentSandboxBuild } from '@/components/sandbox/builds/CurrentSandboxBuild'
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
import {
  getAppConfig,
  getAppConfigs,
  getBranchWorkflowRuns,
  getSandboxBuild,
  getSandboxBuilds,
} from '@/lib'
import { isTerminalStatusV2 } from '@/lib/sse/use-sse-resource-query'

export const Sandbox = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branchId } = useParams()

  const { data: branchRuns } = useQuery({
    queryKey: ['branch-runs', org?.id, app?.id, branchId, 'sandbox-overview'],
    queryFn: () =>
      getBranchWorkflowRuns({
        orgId: org!.id,
        appId: app!.id,
        branchId: branchId!,
        limit: 10,
      }),
    enabled: !!org?.id && !!app?.id && !!branchId,
  })
  const latestRun = branchRuns?.data?.[0]?.app_branch_runs?.at(0)

  const { data: configs } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: app.id, limit: 1 }),
    enabled: !!org?.id && !!app?.id && !branchId,
  })

  const appConfigId = branchId ? latestRun?.app_config_id : configs?.at(0)?.id

  const { data: appConfig, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-config', org?.id, app?.id, appConfigId, 'recurse'],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: app.id,
        appConfigId,
        recurse: true,
      }),
    enabled: !!org?.id && !!app?.id && !!appConfigId,
  })

  const { data: sandboxBuilds } = useQuery({
    queryKey: ['sandbox-builds', org?.id, app?.id, branchId, 'overview'],
    queryFn: () =>
      getSandboxBuilds({
        orgId: org!.id,
        appId: app!.id,
        limit: 50,
      }),
    enabled: !!org?.id && !!app?.id && !!branchId,
  })

  const latestBuildSummary = sandboxBuilds?.data?.find(
    (build) =>
      build.app_branch_id === branchId &&
      (!appConfigId || build.app_config_id === appConfigId)
  )

  const { data: latestBuild } = useQuery({
    placeholderData: latestBuildSummary,
    queryKey: [
      'sandbox-build',
      'overview',
      org?.id,
      app?.id,
      latestBuildSummary?.id,
    ],
    queryFn: () =>
      getSandboxBuild({
        orgId: org!.id,
        appId: app!.id,
        buildId: latestBuildSummary!.id,
      }),
    enabled: !!org?.id && !!app?.id && !!latestBuildSummary?.id,
    refetchInterval: (query) => {
      if (isTerminalStatusV2(query.state.data)) return false
      return 5000
    },
  })

  const sourceRepo =
    appConfig?.sandbox?.connected_github_vcs_config?.repo ??
    appConfig?.sandbox?.public_git_vcs_config?.repo
  const sandboxBasePath = branchId
    ? `/${org?.id}/apps/${app?.id}/branches/${branchId}/sandbox`
    : `/${org?.id}/apps/${app?.id}/sandbox`

  const history = (
    <SandboxBuildTimeline
      shouldPoll
      branchId={branchId}
      excludeBuildId={latestBuild?.id}
    />
  )

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
                <HistoryPanelButton title="Previous builds" history={history} />
                <BuildSandboxButton />
              </>
            }
          />
        }
      >
        <HistoryRail title="Previous builds" history={history}>
          {isLoading ? (
            <Card>
              <Text>Loading...</Text>
            </Card>
          ) : appConfig?.sandbox ? (
            <div className="flex flex-col gap-4">
              {branchId && latestBuild ? (
                <CurrentSandboxBuild
                  appId={app?.id}
                  orgId={org?.id}
                  build={latestBuild}
                  buildHref={`${sandboxBasePath}/builds/${latestBuild.id}`}
                  sourceRepo={sourceRepo}
                />
              ) : null}
              <Card className="flex flex-col gap-4">
                <Text weight="strong">Sandbox config</Text>
                <SandboxConfig appConfig={appConfig} />
              </Card>
            </div>
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
