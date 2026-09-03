import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { createInstallConfig, updateInstallConfig } from '@/lib'
import { EditStackOverridesModal } from './EditStackOverrides'

const MANAGED_BY_CONFIG_TIP = 'Managed by config. Disable config sync to edit.'

export const EditStackOverridesModalContainer = ({
  ...props
}: Omit<IModal, 'onSubmit'>) => {
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install, refresh } = useInstall()
  const { addToast } = useToast()

  const { appConfig } = useInstallAppConfig()

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
          <Text>Install stack overrides saved.</Text>
        </Toast>
      )
      refresh()
      removeModal(props.modalId)
    },
  })

  return (
    <EditStackOverridesModal
      isPending={isPending}
      error={error}
      currentVpcUrl={ic?.vpc_nested_template_url || ''}
      currentRunnerUrl={ic?.runner_nested_template_url || ''}
      currentCustomStacks={(ic?.custom_nested_stacks || []).map((s) => ({
        name: s.name || '',
        template_url: s.template_url || '',
        index: s.index || 0,
        parameters: s.parameters,
      }))}
      appDefaultVpcUrl={appConfig?.stack?.vpc_nested_template_url || ''}
      appDefaultRunnerUrl={appConfig?.stack?.runner_nested_template_url || ''}
      onSubmit={(data) => mutate(data)}
      {...props}
    />
  )
}

export const EditStackOverridesButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const { install } = useInstall()

  const isManagedByConfig =
    install?.metadata?.managed_by === 'nuon/cli/install-config'

  const handleClick = () => {
    const modal = <EditStackOverridesModalContainer />
    addModal(modal)
  }

  if (isManagedByConfig) {
    return (
      <Button
        disabled
        tooltipProps={{
          tipContent: MANAGED_BY_CONFIG_TIP,
          position: 'left',
          tipContentClassName:
            '!whitespace-normal !w-auto max-w-[200px] text-xs',
          className: 'w-full',
        }}
        {...props}
      >
        <Icon variant="StackSimpleIcon" />
        Edit stack overrides
      </Button>
    )
  }

  return (
    <Button onClick={handleClick} {...props}>
      <Icon variant="StackSimpleIcon" />
      Edit stack overrides
    </Button>
  )
}
