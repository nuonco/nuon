import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { InstallConfigsTimelineComponent } from '@/components/install-configs/InstallConfigsTimeline'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import {
  generateCLIInstallConfig,
  getInstallConfigVersions,
  syncInstallConfig,
} from '@/lib'
import { ConfigSyncButton } from './ConfigSyncButton'
import { InstallConfigFile } from './InstallConfigFile'

export const InstallConfigFileContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const isManagedByConfig =
    install?.metadata?.managed_by === 'nuon/cli/install-config'

  const { data: generatedConfig, isLoading: configLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-generate-cli-config', org?.id, install?.id],
    queryFn: () =>
      generateCLIInstallConfig({
        orgId: org!.id,
        installId: install!.id,
      }),
    enabled: !!org?.id && !!install?.id && isManagedByConfig,
  })

  const { data: versions, isLoading: versionsLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-config-versions', org?.id, install?.id],
    queryFn: () =>
      getInstallConfigVersions({
        orgId: org!.id,
        installId: install!.id,
      }),
    enabled: !!org?.id && !!install?.id && isManagedByConfig,
  })

  const { mutate: syncNow, isPending } = useMutation({
    mutationFn: () =>
      syncInstallConfig({
        orgId: org!.id,
        installId: install!.id,
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
    onError: (error) => {
      addToast(
        <Toast heading="Config sync failed" theme="error">
          <Text>{error?.error || 'Unable to trigger config sync.'}</Text>
        </Toast>
      )
    },
  })

  const latestVersion = versions?.at(0)

  return (
    <InstallConfigFile
      action={
        <ConfigSyncButton isPending={isPending} onSync={() => syncNow()} />
      }
      content={generatedConfig?.content}
      filename={latestVersion?.file_path ?? generatedConfig?.filename}
      history={
        <InstallConfigsTimelineComponent
          versions={versions ?? []}
          isLoading={versionsLoading}
          orgId={org?.id}
          installId={install?.id}
        />
      }
      isLoading={configLoading}
      isManagedByConfig={isManagedByConfig}
      latestVersionId={latestVersion?.id}
      syncedAt={latestVersion?.created_at}
    />
  )
}
