import { useQuery } from '@tanstack/react-query'
import { BackToTop } from '@/components/common/BackToTop'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Markdown } from '@/components/common/Markdown'
import { Skeleton } from '@/components/common/Skeleton'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig, getAppConfigs } from '@/lib'

const CONTAINER_ID = 'app-readme-page'

export const Readme = () => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: configs, isLoading: isLoadingConfigs } = useQuery({
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: app.id, limit: 1 }),
    enabled: !!org?.id && !!app?.id,
  })

  const appConfigId = configs?.at(0)?.id

  const { data: appConfig, isLoading: isLoadingConfig } = useQuery({
    queryKey: ['app-config', org?.id, app?.id, appConfigId],
    queryFn: () =>
      getAppConfig({ orgId: org.id, appId: app.id, appConfigId }),
    enabled: !!org?.id && !!app?.id && !!appConfigId,
  })

  const isLoading = isLoadingConfigs || isLoadingConfig

  return (
    <PageSection id={CONTAINER_ID} className="!pb-6" isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/readme`, text: 'README' },
        ]}
      />

      {isLoading ? (
        <ReadmeSkeleton />
      ) : appConfig?.readme ? (
        <Markdown content={appConfig.readme} />
      ) : (
        <EmptyState
          variant="table"
          emptyTitle="No README in app config"
          emptyMessage="You can add a README for your app in your app config TOML file."
        />
      )}

      <BackToTop containerId={CONTAINER_ID} />
    </PageSection>
  )
}

const ReadmeSkeleton = () => (
  <div className="space-y-4">
    <Skeleton height="24px" />
    <Skeleton height="16px" />
    <Skeleton height="16px" />
    <Skeleton height="20px" />
    <Skeleton height="16px" />
    <Skeleton height="16px" />
    <Skeleton height="16px" />
  </div>
)
