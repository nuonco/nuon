import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { ComponentType } from '@/components/components/ComponentType'
import { BuildSandboxButton } from '@/components/sandbox/management/BuildSandbox'
import { CurrentSandboxBuild } from '@/components/sandbox/builds/CurrentSandboxBuild'
import { SandboxBuildTimeline } from '@/components/sandbox/builds/SandboxBuildTimeline'
import { SandboxConfigCard } from '@/components/sandbox/SandboxConfigCard'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { StatusWithDescription } from '@/components/common/StatusWithDescription'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import {
  HistoryPanelButton,
  HistoryRail,
} from '@/components/layout/HistoryRail'
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
import type { TSandboxConfig } from '@/types'

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
    enabled: !!org?.id && !!app?.id,
  })

  const latestBuildSummary = sandboxBuilds?.data?.find(
    (build) =>
      !branchId || !build.app_branch_id || build.app_branch_id === branchId
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

  const sandboxConfig = appConfig?.sandbox as TSandboxConfig | undefined
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
          <DetailHeader
            backLink={false}
            icon={
              <ComponentType
                type={
                  sandboxConfig?.type === 'pulumi'
                    ? 'pulumi'
                    : 'terraform_module'
                }
                displayVariant="icon-only"
                colorVariant="color"
                iconSize="24"
              />
            }
            title="Sandbox"
            status={
              latestBuild ? (
                <StatusWithDescription
                  statusProps={{
                    status: latestBuild.status_v2?.status ?? latestBuild.status,
                  }}
                  tooltipProps={{
                    tipContent:
                      latestBuild.status_v2?.status_human_description ??
                      latestBuild.status_description,
                  }}
                />
              ) : null
            }
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
            <SandboxConfigCard loading />
          ) : sandboxConfig ? (
            <div className="flex flex-col gap-4">
              {latestBuild ? (
                <CurrentSandboxBuild
                  appId={app?.id}
                  orgId={org?.id}
                  build={latestBuild}
                  buildHref={`${sandboxBasePath}/builds/${latestBuild.id}`}
                />
              ) : null}
              <SandboxConfigCard config={sandboxConfig} />
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
