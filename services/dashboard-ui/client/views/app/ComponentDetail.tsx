import { useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { LabelBadge } from '@/components/common/LabelBadge'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Text } from '@/components/common/Text'
import { BuildTimeline } from '@/components/builds/BuildTimeline'
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
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import {
  getAppConfig,
  getAppConfigs,
  getBranchWorkflowRuns,
  getComponent,
  getComponentBuilds,
} from '@/lib'

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

  const latestBuildWithCommit = latestBuilds?.data?.find(
    (b) =>
      !!b.vcs_connection_commit && (!branchId || b.app_branch_id === branchId)
  )
  const buildCommit = latestBuildWithCommit?.vcs_connection_commit
  const componentBasePath = branchId
    ? `/${org?.id}/apps/${app?.id}/branches/${branchId}/components/${componentId}`
    : `/${org?.id}/apps/${app?.id}/components/${componentId}`
  const latestCommit = buildCommit
    ? {
        status: latestBuildWithCommit?.status_v2?.status,
        href: `${componentBasePath}/builds/${latestBuildWithCommit?.id}`,
        message: buildCommit.message?.split('\n')[0],
        author: buildCommit.author_name,
        avatarUrl: buildCommit.author_avatar_url,
        sha: buildCommit.sha,
        createdAt: buildCommit.created_at,
      }
    : undefined

  const labelKeys = Object.keys(component?.labels ?? {}).sort()
  const history = (
    <BuildTimeline
      componentId={componentId!}
      componentName={component?.name ?? ''}
      shouldPoll
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
            path: `/${org?.id}/apps/${app?.id}/components`,
            text: 'Components',
          },
          {
            path: `/${org?.id}/apps/${app?.id}/components/${componentId}`,
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
                <HistoryPanelButton title="Build history" history={history} />
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
        <HistoryRail title="Build history" history={history}>
          {isLoadingConfig ? (
            <ComponentConfigCard loading />
          ) : config ? (
            <ComponentConfigCard
              config={config}
              latestBuild={latestResolvedBuild}
              latestCommit={latestCommit}
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
