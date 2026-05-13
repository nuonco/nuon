import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createSandboxBuild } from '@/lib'
import { BuildSandboxModal } from './BuildSandbox'

export const BuildSandboxModalContainer = ({ ...props }: Omit<IModal, 'onSubmit'>) => {
  const { org } = useOrg()
  const { app } = useApp()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { error, mutate, isPending } = useMutation({
    mutationFn: () => createSandboxBuild({ appId: app.id, orgId: org.id }),
    onSuccess: () => {
      addToast(
        <Toast heading="Sandbox build started" theme="info" />
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Sandbox build failed" theme="error" />
      )
    },
  })

  return (
    <BuildSandboxModal
      appName={app?.name}
      isPending={isPending}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const BuildSandboxButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <BuildSandboxModalContainer />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="HammerIcon" />}
      Build sandbox
      {props?.isMenuButton ? <Icon variant="HammerIcon" /> : null}
    </Button>
  )
}
