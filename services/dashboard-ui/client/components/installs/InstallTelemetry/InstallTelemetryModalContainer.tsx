import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import {
  createInstallConfig,
  getInstallTelemetrySettings,
  updateInstallConfig,
} from '@/lib/ctl-api/installs'
import type { TInstall } from '@/types'
import { CollectorSetupContainer } from './CollectorSetupContainer'
import { InstallTelemetryModal } from './InstallTelemetryModal'

export const InstallTelemetryModalContainer = ({
  install,
  orgId,
  ...props
}: { install: TInstall; orgId: string } & Omit<IModal, 'onSubmit'>) => {
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const installId = install.id!
  const key = ['install-telemetry', orgId, installId]
  const settings = useQuery({
    queryKey: key,
    queryFn: () => getInstallTelemetrySettings({ orgId, installId }),
    retry: false,
  })
  const mutation = useMutation({
    mutationFn: (enabled: boolean | null) => {
      const cached = queryClient.getQueryData<TInstall>([
        'install',
        orgId,
        installId,
      ])
      const installConfigId =
        cached?.install_config?.id || install.install_config?.id
      const params = { orgId, installId, body: { telemetry: { enabled } } }
      return installConfigId
        ? updateInstallConfig({ ...params, installConfigId })
        : createInstallConfig(params)
    },
    onSuccess: async (data) => {
      const installKey = ['install', orgId, installId]
      await queryClient.cancelQueries({ queryKey: installKey, exact: true })
      queryClient.setQueryData<TInstall>(installKey, (old) => ({
        ...old,
        id: installId,
        install_config: data,
      }))
      await queryClient.invalidateQueries({ queryKey: key })
      queryClient.invalidateQueries({ queryKey: installKey })
      queryClient.invalidateQueries({
        queryKey: ['runner', orgId, install.runner_id],
      })
      removeModal(props.modalId)
      addToast(
        <Toast heading="Telemetry settings updated" theme="success">
          <Text>Collectors apply changes on their next settings refresh.</Text>
        </Toast>
      )
    },
  })

  if (!settings.data || settings.isError) {
    return (
      <Modal heading="Manage telemetry" {...props}>
        <FormErrorBanner
          error={settings.error}
          fallback="Telemetry settings load failed"
        />
        {settings.isError ? (
          <Button className="w-fit" onClick={() => settings.refetch()}>
            Retry settings
          </Button>
        ) : (
          <Text loading>Loading settings</Text>
        )}
      </Modal>
    )
  }

  return (
    <InstallTelemetryModal
      key={`${settings.data.override}:${settings.data.org_default}`}
      installName={install.name || installId}
      settings={settings.data}
      isManagedByConfig={
        install.metadata?.managed_by === 'nuon/cli/install-config'
      }
      isPending={mutation.isPending}
      error={mutation.error}
      onSubmit={(enabled) => mutation.mutate(enabled)}
      {...props}
    >
      <CollectorSetupContainer
        install={install}
        orgId={orgId}
        onNavigate={() => removeModal(props.modalId)}
      />
    </InstallTelemetryModal>
  )
}
