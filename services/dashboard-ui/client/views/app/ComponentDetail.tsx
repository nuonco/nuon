import { useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { LabelBadge } from '@/components/common/LabelBadge'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { BuildTimeline } from '@/components/builds/BuildTimeline'
import { CurrentComponentBuild } from '@/components/builds/CurrentComponentBuild'
import { ComponentConfigCard } from '@/components/components/ComponentConfigCard'
import { ComponentDependencies } from '@/components/components/ComponentDependencies'
import { ComponentDependencyGraphButton } from '@/components/components/ComponentDependencyGraph'
import { ComponentType } from '@/components/components/ComponentType'
import { BuildComponentButton } from '@/components/components/management/BuildComponent'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import {
  HistoryPanelButton,
  HistoryRail,
} from '@/components/layout/HistoryRail'
import { Text } from '@/components/common/Text'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import {
  getAppConfig,
  getAppConfigs,
  getBranchWorkflowRuns,
  getComponent,
  getComponentBuild,
  getComponentBuilds,
} from '@/lib'
import { isTerminalStatusV2 } from '@/lib/sse/use-sse-resource-query'
export const ComponentDetail = () => {
  const { componentId, branchId } = useParams()
  const { org } = useOrg()
  const { app, labelColors } = useApp()

  const { data: component, isLoading: isLoadingComponent } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['component', org?.id, app?.id, componentId],
    queryFn: () => getComponent({ orgId: org.id, componentId: componentId! }),
    enabled: !!org?.id && !!app?.id && !!componentId,
  })

  const { data: branchRuns, isLoading: isLoadingBranchRuns } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-runs-latest', org?.id, app?.id, branchId],
    queryFn: () =>
      getBranchWorkflowRuns({
        orgId: org.id,
        appId: app.id,
        branchId: branchId!,
        limit: 1,
      }),
    enabled: !!org?.id && !!app?.id && !!branchId,
  })
  const branchAppConfigId = branchId
    ? branchRuns?.data?.at(0)?.app_branch_runs?.at(0)?.app_config_id
    : undefined

  const { data: configs } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: app.id, limit: 1 }),
    enabled: !!org?.id && !!app?.id && !branchId,
  })

  const appConfigId = branchId ? branchAppConfigId : configs?.at(0)?.id

  const { data: appConfig, isLoading: isLoadingAppConfig } = useQuery({
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
  const isLoadingConfig =
    isLoadingAppConfig || (!!branchId && isLoadingBranchRuns)

  const config = appConfig?.component_config_connections?.find(
    (c) => c.component_id === componentId
  )

  const dependentIds =
    appConfig?.component_config_connections
      ?.filter((c) => c.component_dependency_ids?.includes(componentId!))
      .map((c) => c.component_id!)
      .filter(Boolean) ?? []

  const { data: latestBuilds } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['component-builds', org?.id, componentId, 0],
    queryFn: () =>
      getComponentBuilds({
        orgId: org.id,
        componentId: componentId!,
        limit: 10,
        offset: 0,
      }),
    enabled: !!org?.id && !!componentId,
  })
  const latestResolvedBuild = latestBuilds?.data?.find((b) => !!b.source_digest)

  const latestBuildSummary = latestBuilds?.data?.find(
    (b) => !branchId || b.app_branch_id === branchId
  )

  const { data: latestBuild } = useQuery({
    placeholderData: latestBuildSummary,
    queryKey: ['component-build', org?.id, componentId, latestBuildSummary?.id],
    queryFn: () =>
      getComponentBuild({
        orgId: org!.id,
        componentId: componentId!,
        buildId: latestBuildSummary!.id,
      }),
    enabled:
      !!org?.id && !!componentId && !!branchId && !!latestBuildSummary?.id,
    refetchInterval: (query) => {
      if (isTerminalStatusV2(query.state.data)) return false
      return 5000
    },
  })

  const appBase = branchId
    ? `/${org?.id}/apps/${app?.id}/branches/${branchId}`
    : `/${org?.id}/apps/${app?.id}`
  const componentBasePath = `${appBase}/components/${componentId}`
  const labelKeys = Object.keys(component?.labels ?? {}).sort()
  const history = (
    <BuildTimeline
      componentId={componentId!}
      componentName={component?.name ?? ''}
      shouldPoll
      branchId={branchId}
      excludeBuildId={latestBuild?.id}
    />
  )

  return (
    <>
      <PageTitle segments={[component?.name ?? 'Component', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          {
            path: `${appBase}/components`,
            text: 'Components',
          },
          {
            path: componentBasePath,
            text: component?.name,
          },
        ]}
      />

      <DetailPage
        header={
          <DetailHeader
            icon={
              <ComponentType
                type={component?.type}
                displayVariant="icon-only"
                colorVariant="color"
                iconSize="24"
              />
            }
            title={component?.name}
            loading={isLoadingComponent}
            loadingWidth={20}
            status={
              config?.toggleable ? (
                <>
                  <Badge size="sm" theme="info">
                    Toggleable
                  </Badge>
                  <Badge
                    size="sm"
                    theme={config?.default_enabled ? 'success' : 'neutral'}
                  >
                    {config?.default_enabled ? 'Default: on' : 'Default: off'}
                  </Badge>
                </>
              ) : null
            }
            id={component?.id}
            identity={
              labelKeys.length ? (
                <span className="flex flex-wrap gap-1">
                  {labelKeys.map((k) => (
                    <LabelBadge
                      key={k}
                      labelKey={k}
                      labelValue={component?.labels?.[k]}
                      size="sm"
                      customColor={labelColors?.[k]}
                    />
                  ))}
                </span>
              ) : null
            }
            actions={
              <>
                <HistoryPanelButton title="Previous builds" history={history} />
                {component ? (
                  <BuildComponentButton
                    component={component}
                    variant="primary"
                  />
                ) : null}
              </>
            }
          />
        }
      >
        <HistoryRail title="Previous builds" history={history}>
          {isLoadingConfig ? (
            <ComponentConfigCard loading />
          ) : config ? (
            <div className="flex flex-col gap-4">
              {branchId && latestBuild ? (
                <CurrentComponentBuild
                  appId={app?.id}
                  orgId={org?.id}
                  build={latestBuild}
                  buildHref={`${componentBasePath}/builds/${latestBuild.id}`}
                />
              ) : null}
              <ComponentConfigCard
                config={config}
                latestBuild={latestResolvedBuild}
                headerActions={
                  appConfig && componentId && component?.name ? (
                    <ComponentDependencyGraphButton
                      componentId={componentId}
                      componentName={component.name}
                      componentType={component.type}
                      appConfig={appConfig}
                      basePath={`/${org?.id}/apps/${app?.id}/components`}
                      size="sm"
                    />
                  ) : null
                }
                footer={
                  config.component_dependency_ids?.length ||
                  dependentIds.length > 0 ? (
                    <>
                      {config.component_dependency_ids?.length ? (
                        <div className="flex flex-col gap-2">
                          <Text variant="body" weight="strong" level={5}>
                            Dependencies
                          </Text>
                          <ComponentDependencies
                            deps={config.component_dependency_ids}
                            variant="inline"
                          />
                        </div>
                      ) : null}
                      {dependentIds.length > 0 ? (
                        <div className="flex flex-col gap-2">
                          <Text variant="body" weight="strong" level={5}>
                            Dependents
                          </Text>
                          <ComponentDependencies
                            deps={dependentIds}
                            variant="inline"
                            tooltipTitle="More dependents"
                          />
                        </div>
                      ) : null}
                    </>
                  ) : undefined
                }
              />
            </div>
          ) : (
            <EmptyState
              variant="table"
              emptyTitle="No configuration"
              emptyMessage="This component has no configuration yet."
            />
          )}
        </HistoryRail>
      </DetailPage>
    </>
  )
}
