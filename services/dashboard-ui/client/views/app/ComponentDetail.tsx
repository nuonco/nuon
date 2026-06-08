import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { BackLink } from '@/components/common/BackLink'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { BuildTimeline } from '@/components/builds/BuildTimeline'
import {
  ComponentConfigCard,
  ComponentConfigCardSkeleton,
} from '@/components/components/ComponentConfigCard'
import { ComponentDependencies } from '@/components/components/ComponentDependencies'
import { ComponentDependencyGraphButton } from '@/components/components/ComponentDependencyGraph'
import { ComponentType } from '@/components/components/ComponentType'
import { BuildComponentButton } from '@/components/components/management/BuildComponent'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { Panel } from '@/components/surfaces/Panel'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import {
  getAppConfig,
  getAppConfigs,
  getComponent,
  getComponentBuilds,
} from '@/lib'

export const ComponentDetail = () => {
  const { componentId } = useParams()
  const { org } = useOrg()
  const { app } = useApp()
  const { addPanel } = useSurfaces()

  const { data: component, isLoading: isLoadingComponent } = useQuery({
    queryKey: ['component', org?.id, app?.id, componentId],
    queryFn: () => getComponent({ orgId: org.id, componentId: componentId! }),
    enabled: !!org?.id && !!app?.id && !!componentId,
  })

  const { data: configs } = useQuery({
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: app.id, limit: 1 }),
    enabled: !!org?.id && !!app?.id,
  })

  const appConfigId = configs?.at(0)?.id

  const { data: appConfig, isLoading: isLoadingConfig } = useQuery({
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

  const config = appConfig?.component_config_connections?.find(
    (c) => c.component_id === componentId
  )

  const dependentIds = appConfig?.component_config_connections
    ?.filter((c) => c.component_dependency_ids?.includes(componentId!))
    .map((c) => c.component_id!)
    .filter(Boolean) ?? []

  const { data: latestBuilds } = useQuery({
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
  const latestResolvedBuild = latestBuilds?.data?.find(
    (b) => !!b.source_digest
  )

  return (
    <PageSection>
      <PageTitle title={`${component?.name ?? 'Component'} | ${app?.name}`} />
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

      <div className="@container flex flex-col flex-auto gap-6">
        <div className="flex items-start justify-between">
          <HeadingGroup>
            <BackLink className="mb-6" />
            <span className="flex items-center gap-2">
              <ComponentType
                type={component?.type}
                displayVariant="icon-only"
                colorVariant="color"
                iconSize="24"
              />
              <Text variant="base" weight="strong">
                {component?.name}
              </Text>
            </span>
            {component?.id ? <ID>{component.id}</ID> : null}
            {component?.labels && Object.keys(component.labels).length > 0 ? (
              <span className="flex flex-wrap gap-1 mt-1">
                {Object.keys(component.labels)
                  .sort()
                  .map((k) => (
                    <Badge key={k} variant="code" size="sm" theme="neutral">
                      {k}: {component.labels[k]}
                    </Badge>
                  ))}
              </span>
            ) : null}
          </HeadingGroup>

          <div className="flex items-center gap-2">
            <div className="@5xl:hidden">
              <Button
                variant="secondary"
                onClick={() =>
                  addPanel(
                    <Panel heading="Build history">
                      <BuildTimeline
                        componentId={componentId!}
                        componentName={component?.name ?? ''}
                        shouldPoll
                      />
                    </Panel>
                  )
                }
              >
                <Icon variant="ClockCounterClockwiseIcon" size={16} />
                Build history
              </Button>
            </div>
            {component ? (
              <BuildComponentButton component={component} variant="primary" />
            ) : null}
          </div>
        </div>

        <div className="grid grid-cols-1 @5xl:grid-cols-12 gap-6">
          <div className="@5xl:col-span-8 flex flex-col gap-6">
            {isLoadingConfig ? (
              <ComponentConfigCardSkeleton />
            ) : config ? (
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
                  (config.component_dependency_ids?.length || dependentIds.length > 0) ? (
                    <>
                      {config.component_dependency_ids?.length ? (
                        <div className="flex flex-col gap-2">
                          <Text variant="body" weight="strong" level={5}>Dependencies</Text>
                          <ComponentDependencies
                            deps={config.component_dependency_ids}
                            variant="inline"
                          />
                        </div>
                      ) : null}
                      {dependentIds.length > 0 ? (
                        <div className="flex flex-col gap-2">
                          <Text variant="body" weight="strong" level={5}>Dependents</Text>
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
          </div>

          <div className="hidden @5xl:flex flex-col @5xl:col-span-4 gap-4">
            <Text variant="base" weight="strong">
              Build history
            </Text>
            <BuildTimeline
              componentId={componentId!}
              componentName={component?.name ?? ''}
              shouldPoll
            />
          </div>
        </div>
      </div>

    </PageSection>
  )
}
