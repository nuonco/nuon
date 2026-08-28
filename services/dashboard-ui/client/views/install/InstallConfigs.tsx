import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { InstallConfigsTimeline } from '@/components/install-configs/InstallConfigsTimeline'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { syncInstallConfig } from '@/lib'

export const InstallConfigs = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate: syncNow, isPending: isSyncing } = useMutation({
    mutationFn: () =>
      syncInstallConfig({
        installId: install!.id,
        orgId: org!.id,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['install-config-versions', org?.id, install?.id],
      })
      addToast(
        <Toast heading="Syncing config" theme="info">
          <Text>Syncing configs for {install?.name}.</Text>
        </Toast>
      )
    },
    onError: (err) => {
      addToast(
        <Toast heading="Config sync failed" theme="error">
          <Text>{err?.error || 'Unable to trigger config sync.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <PageSection>
      <PageTitle segments={['Configs', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/configs`,
            text: 'Configs',
          },
        ]}
      />

      <SectionHeader
        title="Config history"
        description="Install config versions synced from git."
        actions={
          install?.id ? (
            <Button
              variant="secondary"
              disabled={isSyncing}
              onClick={() => syncNow()}
            >
              {isSyncing ? 'Syncing config' : 'Sync now'}
            </Button>
          ) : null
        }
      />

      <InstallConfigsTimeline />
    </PageSection>
  )
}
