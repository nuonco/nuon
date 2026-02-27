import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { BuildTimeline } from '@/components/builds/BuildTimeline'
import { ComponentConfigCard, ComponentConfigCardSkeleton } from '@/components/components/ComponentConfigCard'
import { ComponentDependencies } from '@/components/components/ComponentDependencies'
import { ComponentType } from '@/components/components/ComponentType'
import { BuildComponentButton } from '@/components/components/management/BuildComponent'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { api } from '@/lib/api'
import { getComponent } from '@/lib'
import type { TComponentConfig } from '@/types'

const CONTAINER_ID = 'component-detail-page'

export const ComponentDetail = () => {
  const { componentId } = useParams()
  const { org } = useOrg()
  const { app } = useApp()

  const { data: component, isLoading: isLoadingComponent } = useQuery({
    queryKey: ['component', org?.id, app?.id, componentId],
    queryFn: () => getComponent({ orgId: org.id, appId: app.id, componentId: componentId! }),
    enabled: !!org?.id && !!app?.id && !!componentId,
  })

  const { data: config, isLoading: isLoadingConfig } = useQuery({
    queryKey: ['component-config-latest', org?.id, componentId],
    queryFn: () => api<TComponentConfig>({ orgId: org.id, path: `components/${componentId}/configs/latest` }),
    enabled: !!org?.id && !!componentId,
  })


  return (
    <PageSection id={CONTAINER_ID} isScrollable className="!p-0 !gap-0">
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/components`, text: 'Components' },
          { path: `/${org?.id}/apps/${app?.id}/components/${componentId}`, text: component?.name },
        ]}
      />

      <div className="p-6 border-b flex justify-between">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <span className="flex items-center gap-2">
            <ComponentType type={component?.type} displayVariant="icon-only" />
            <Text variant="base" weight="strong">
              {component?.name}
            </Text>
          </span>
          {component?.id ? <ID>{component.id}</ID> : null}
          {component?.dependencies?.length ? (
            <ComponentDependencies deps={component.dependencies} />
          ) : null}
        </HeadingGroup>

        {component ? (
          <BuildComponentButton component={component} variant="primary" />
        ) : null}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="divide-y flex flex-col md:col-span-8">
          <PageSection>
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
          </PageSection>
        </div>

        <div className="divide-y flex flex-col md:col-span-4">
          <PageSection>
            <Text variant="base" weight="strong">
              Build history
            </Text>
            <BuildTimeline
              componentId={componentId!}
              componentName={component?.name ?? ''}
              pagination={{ hasNext: false, offset: 0, limit: 10 }}
              shouldPoll
            />
          </PageSection>
        </div>
      </div>

      <BackToTop containerId={CONTAINER_ID} />
    </PageSection>
  )
}
