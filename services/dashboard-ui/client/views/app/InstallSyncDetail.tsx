import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Navigate, useParams } from 'react-router'
import { Text } from '@/components/common/Text'
import { InstallSyncDetail as InstallSyncDetailContent } from '@/components/installs/InstallSyncDetail'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppInstallSync } from '@/lib'

export const InstallSyncDetail = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const params = useParams()
  const syncId = params.syncId as string
  const hasInstallSyncing = !!org?.features?.['app-install-syncing']

  const { data: sync, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-install-sync', org?.id, app?.id, syncId],
    queryFn: () =>
      getAppInstallSync({ appId: app!.id, syncId, orgId: org!.id }),
    enabled: hasInstallSyncing && !!org?.id && !!app?.id && !!syncId,
    refetchInterval: 5000,
  })

  if (org && !hasInstallSyncing) {
    return <Navigate to={`/${org.id}/apps/${app?.id}`} replace />
  }

  if (isLoading || !sync) {
    return (
      <PageSection>
        <Text variant="body" theme="neutral">
          Loading install sync...
        </Text>
      </PageSection>
    )
  }

  return (
    <PageSection className="max-w-full">
      <PageTitle segments={['Install sync', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          {
            path: `/${org?.id}/apps/${app?.id}/install-syncs`,
            text: 'Install syncs',
          },
          {
            path: `/${org?.id}/apps/${app?.id}/install-syncs/${syncId}`,
            text: 'Sync',
          },
        ]}
      />

      <InstallSyncDetailContent
        sync={sync}
        orgId={org?.id ?? ''}
        appId={app?.id ?? ''}
        syncId={syncId}
      />
    </PageSection>
  )
}
