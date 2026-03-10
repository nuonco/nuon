import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { BuildTimeline } from '@/components/builds/BuildTimeline'
import {
  ComponentConfigCard,
  ComponentConfigCardSkeleton,
} from '@/components/components/ComponentConfigCard'
import { ComponentDependencies } from '@/components/components/ComponentDependencies'
import { ComponentType } from '@/components/components/ComponentType'
import { BuildComponentButton } from '@/components/components/management/BuildComponent'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig, getAppConfigs, getComponent } from '@/lib'

const CONTAINER_ID = 'component-detail-page'

export const ComponentDetail = () => {
  const { componentId } = useParams()
  const { org } = useOrg()
  const { app } = useApp()

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

  return (
    <PageSection id={CONTAINER_ID} isScrollable>
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

      <div className="flex items-start justify-between">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <span className="flex items-center gap-2">
            <ComponentType type={component?.type} displayVariant="icon-only" />
            <Text variant="base" weight="strong">
              {component?.name}
            </Text>
          </span>
          {component?.id ? <ID>{component.id}</ID> : null}
        </HeadingGroup>

        {component ? (
          <BuildComponentButton component={component} variant="primary" />
        ) : null}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto gap-6">
        <div className="md:col-span-8 flex flex-col gap-6">
          {config?.component_dependency_ids?.length ? (
            <Card>
              <Text weight="strong">Dependencies</Text>
              <ComponentDependencies
                deps={config.component_dependency_ids}
                variant="inline"
              />
            </Card>
          ) : null}

          {isLoadingConfig ? (
            <ComponentConfigCardSkeleton />
          ) : config ? (
            <ComponentConfigCard config={config} />
          ) : (
            <EmptyState
              variant="table"
              emptyTitle="No configuration"
              emptyMessage="This component has no configuration yet."
            />
          )}
        </div>

        <div className="md:col-span-4 flex flex-col gap-4">
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

      <BackToTop containerId={CONTAINER_ID} />
    </PageSection>
  )
}
