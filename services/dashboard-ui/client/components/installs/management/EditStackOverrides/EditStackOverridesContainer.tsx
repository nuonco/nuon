import { useQuery, useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import {
  getAppConfig,
  createInstallConfig,
  updateInstall,
  updateInstallConfig,
} from '@/lib'
import { ConfirmOverrideModal } from '../EnableAutoApprove/EnableAutoApprove'
import { EditStackOverridesModal } from './EditStackOverrides'

const ConfirmOverrideModalContainer = ({ onConfirm, ...props }: { onConfirm: () => void } & IModal) => {
  const { removeModal } = useSurfaces()
  const { install } = useInstall()

  const isInstallManagedByConfig = install?.metadata?.managed_by === 'nuon/cli/install-config'

  if (!isInstallManagedByConfig) {
    return null
  }

  return (
    <ConfirmOverrideModal
      onConfirm={() => {
        onConfirm()
        removeModal(props.modalId)
      }}
      {...props}
    />
  )
}

export const EditStackOverridesModalContainer = ({
  ...props
}: Omit<IModal, 'onSubmit'>) => {
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install, refresh } = useInstall()
  const { addToast } = useToast()

  const { data: appConfig } = useQuery({
    queryKey: ['app-config', org?.id, install?.app_id, install?.app_config_id, 'recurse'],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: install.app_config_id,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_config_id,
  })

  const hasInstallConfig = Boolean(install?.install_config)
  const ic = install?.install_config

  const { mutate, isPending, error } = useMutation({
    mutationFn: async (data: {
      vpc_nested_template_url?: string
      runner_nested_template_url?: string
      custom_nested_stacks?: Array<{
        name: string
        template_url: string
        index?: number
        parameters?: Record<string, string>
      }>
    }) => {
      if (install?.metadata?.managed_by === 'nuon/cli/install-config') {
        await updateInstall({
          orgId: org.id,
          installId: install.id,
          body: { metadata: { managed_by: 'nuon/dashboard' } },
        })
      }

      if (hasInstallConfig) {
        return updateInstallConfig({
          orgId: org.id,
          installId: install.id,
          installConfigId: ic!.id!,
          body: data,
        })
      } else {
        return createInstallConfig({
          orgId: org.id,
          installId: install.id,
          body: {
            approval_option: ic?.approval_option || 'prompt',
            ...data,
          },
        })
      }
    },
    onSuccess: () => {
      addToast(
        <Toast heading="Stack overrides updated" theme="success">
          <Text>Install stack overrides saved successfully.</Text>
        </Toast>
      )
      refresh()
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Save failed" theme="error">
          <Text>Unable to save stack overrides.</Text>
        </Toast>
      )
    },
  })

  return (
    <EditStackOverridesModal
      isPending={isPending}
      error={error}
      currentVpcUrl={ic?.vpc_nested_template_url || ''}
      currentRunnerUrl={ic?.runner_nested_template_url || ''}
      currentCustomStacks={
        (ic?.custom_nested_stacks || []).map((s) => ({
          name: s.name || '',
          template_url: s.template_url || '',
          index: s.index || 0,
          parameters: s.parameters,
        }))
      }
      appDefaultVpcUrl={appConfig?.stack?.vpc_nested_template_url || ''}
      appDefaultRunnerUrl={appConfig?.stack?.runner_nested_template_url || ''}
      onSubmit={(data) => mutate(data)}
      {...props}
    />
  )
}

export const EditStackOverridesButton = ({
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const { install } = useInstall()

  const isInstallManagedByConfig = install?.metadata?.managed_by === 'nuon/cli/install-config'

  const handleClick = () => {
    if (isInstallManagedByConfig) {
      const overrideModal = (
        <ConfirmOverrideModalContainer
          onConfirm={() => {
            const mainModal = <EditStackOverridesModalContainer />
            addModal(mainModal)
          }}
        />
      )
      addModal(overrideModal)
    } else {
      const modal = <EditStackOverridesModalContainer />
      addModal(modal)
    }
  }

  return (
    <Button
      onClick={handleClick}
      {...props}
    >
      <Icon variant="StackSimple" />
      Edit stack overrides
    </Button>
  )
}
