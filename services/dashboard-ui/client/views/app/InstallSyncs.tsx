import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Navigate } from 'react-router'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { AppInstallSyncsTimeline } from '@/components/apps/AppInstallSyncsTimeline'
import { ManageInstallsConfigButton } from '@/components/apps/ManageInstallsConfig'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getAppInstallsConfig, triggerAppInstallSync } from '@/lib'
import type { TAPIError } from '@/types'

export const InstallSyncs = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const hasInstallSyncing = !!org?.features?.['app-install-syncing']

  const { data: installsConfig } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-installs-config', org?.id, app?.id],
    queryFn: () => getAppInstallsConfig({ appId: app!.id, orgId: org!.id }),
    enabled: hasInstallSyncing && !!org?.id && !!app?.id,
  })

  const { mutate: triggerSync, isPending } = useMutation({
    mutationFn: () =>
      triggerAppInstallSync({ appId: app!.id, orgId: org!.id }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['app-install-syncs', org?.id, app?.id],
      })
      addToast(
        <Toast heading="Sync triggered" theme="info">
          <Text>Syncing install configs for {app?.name}.</Text>
        </Toast>
      )
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Sync failed" theme="error">
          <Text>{err?.error || 'Unable to trigger sync.'}</Text>
        </Toast>
      )
    },
  })

  if (org && !hasInstallSyncing) {
    return <Navigate to={`/${org.id}/apps/${app?.id}`} replace />
  }

  return (
    <PageSection>
      <PageTitle segments={['Install syncs', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          {
            path: `/${org?.id}/apps/${app?.id}/install-syncs`,
            text: 'Install syncs',
          },
        ]}
      />
      <SectionHeader
        title="Install syncs"
        description="Sync install configurations from git."
        actions={
          <>
            <AdminDashboardLink
              path={`/queues?owner_id=${app?.id}&owner_type=apps&name=app-install-syncs`}
              label="View queue"
            />
            <ManageInstallsConfigButton />
            <Button
              variant="primary"
              onClick={() => triggerSync()}
              disabled={isPending}
            >
              {isPending ? 'Syncing installs' : 'Sync now'}
            </Button>
          </>
        }
      />

      {installsConfig ? (
        <Card>
          <Text weight="strong">Config source</Text>
          <div className="grid grid-cols-4 gap-3">
            <LabeledValue label="Source">
              <Badge variant="code" size="md">
                {installsConfig.source === 'config'
                  ? 'installs.toml'
                  : 'dashboard'}
              </Badge>
            </LabeledValue>
            <LabeledValue label="Repo">{installsConfig.repo}</LabeledValue>
            <LabeledValue label="Branch">{installsConfig.branch}</LabeledValue>
            {installsConfig.directory !== '.' ? (
              <LabeledValue label="Directory">
                /{installsConfig.directory}
              </LabeledValue>
            ) : null}
          </div>
        </Card>
      ) : null}

      <SectionHeader
        title="Sync history"
        description="Each run of the installs config sync. Select a run to see its results."
      />

      <AppInstallSyncsTimeline shouldPoll />
    </PageSection>
  )
}
