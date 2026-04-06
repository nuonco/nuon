import { useMutation } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createSandboxBuild } from '@/lib'

export const BuildSandboxButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <BuildSandboxModal />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="Hammer" />}
      Build sandbox
      {props?.isMenuButton ? <Icon variant="Hammer" /> : null}
    </Button>
  )
}

export const BuildSandboxModal = ({ ...props }: IModal) => {
  const { org } = useOrg()
  const { app } = useApp()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { error, mutate, isPending: isLoading } = useMutation({
    mutationFn: () => createSandboxBuild({ appId: app.id, orgId: org.id }),
    onSuccess: () => {
      addToast(
        <Toast heading="Sandbox build started" theme="success">
          <Text>Sandbox build for {app.name} was started.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Sandbox build failed" theme="error">
          <Text>Unable to start sandbox build for {app.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
          theme="info"
        >
          <Icon variant="Hammer" size="24" />
          Build sandbox?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Building sandbox
          </span>
        ) : (
          'Build sandbox'
        ),
        disabled: isLoading,
        onClick: () => mutate(),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3 mb-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to start sandbox build'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          Are you sure you want to build the sandbox for {app?.name}?
        </Text>
        <Text variant="base">
          This will start a sandbox build. The build process may take several
          minutes to complete.
        </Text>
      </div>
    </Modal>
  )
}
