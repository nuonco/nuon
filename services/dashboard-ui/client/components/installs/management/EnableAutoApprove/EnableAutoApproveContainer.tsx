import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createInstallConfig, updateInstallConfig } from '@/lib'
import { EnableAutoApproveModal } from './EnableAutoApprove'

const MANAGED_BY_CONFIG_TIP = 'Managed by config. Disable config sync to edit.'

export const EnableAutoApproveModalContainer = ({ ...props }: Omit<IModal, 'onSubmit'>) => {
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()

  const hasInstallConfig = Boolean(install?.install_config)
  const isApproveAll = hasInstallConfig && install?.install_config?.approval_option === 'approve-all'

  const { mutate, isPending: isLoading, data, error } = useMutation({
    mutationFn: async () => {
      if (hasInstallConfig) {
        return updateInstallConfig({
          orgId: org.id,
          installId: install.id,
          installConfigId: install.install_config.id,
          body: { approval_option: isApproveAll ? 'prompt' : 'approve-all' },
        })
      } else {
        return createInstallConfig({
          orgId: org.id,
          installId: install.id,
          body: { approval_option: 'approve-all' },
        })
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['install', org.id, install.id] })
      addToast(
        <Toast heading={`Auto approve ${isApproveAll ? 'disabled' : 'enabled'}`} theme="success">
          <Text>Auto approve {isApproveAll ? 'disabled' : 'enabled'} for {install.name}.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (error) => {
      addToast(
        <Toast heading={`Auto approve ${isApproveAll ? 'disable' : 'enable'} failed`} theme="error">
          <Text>Unable to {isApproveAll ? 'disable' : 'enable'} auto approve for {install.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <EnableAutoApproveModal
      isPending={isLoading}
      error={error}
      isApproveAll={isApproveAll}
      isSuccess={Boolean(data)}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const EnableAutoApproveButton = ({
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const { install } = useInstall()

  const hasInstallConfig = Boolean(install?.install_config)
  const isApproveAll = hasInstallConfig && install?.install_config?.approval_option === 'approve-all'
  const isManagedByConfig = install?.metadata?.managed_by === 'nuon/cli/install-config'

  const buttonText = isApproveAll ? 'Disable auto approval' : 'Enable auto approval'
  const buttonIcon = isApproveAll ? 'ToggleRightIcon' : 'ToggleLeftIcon'

  const handleClick = () => {
    const modal = <EnableAutoApproveModalContainer />
    addModal(modal)
  }

  if (isManagedByConfig) {
    return (
      <Tooltip tipContent={MANAGED_BY_CONFIG_TIP} position="left" tipContentClassName="!whitespace-normal !w-auto max-w-[200px] text-xs" className="w-full">
        <Button disabled className="pointer-events-none" {...props}>
          {buttonText}
          <Icon variant={buttonIcon} />
        </Button>
      </Tooltip>
    )
  }

  return (
    <Button onClick={handleClick} {...props}>
      {buttonText}
      <Icon variant={buttonIcon} />
    </Button>
  )
}
