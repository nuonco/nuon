import { useEffect } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { shutdownRunnerInstance } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'

export const ShutdownInstanceButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ShutdownInstanceModal />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="CloudArrowDown" />}
      Restart instance
      {props?.isMenuButton ? <Icon variant="CloudArrowDown" /> : null}
    </Button>
  )
}

export const ShutdownInstanceModal = ({ ...props }: IModal) => {
  const { user } = useAuth()
  const { org } = useOrg()
  const { runner } = useRunner()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const {
    data: isShutdown,
    error,
    mutate,
    isPending: isLoading,
  } = useMutation({
    mutationFn: () => shutdownRunnerInstance({ runnerId: runner.id, orgId: org.id }),
    onSuccess: () => {
      addToast(
        <Toast heading="Restart runner instance started" theme="success">
          <Text>Restart runner instance initiated successfully.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Restart runner instance failed" theme="error">
          <Text>Unable to restart runner instance.</Text>
        </Toast>
      )
    },
  })

  const handleClose = () => {
    removeModal(props.modalId)
  }

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'runner_shutdown_instance',
        status: 'error',
        user,
        props: { orgId: org.id, runnerId: runner.id, err: error?.error },
      })
    }
    if (isShutdown) {
      trackEvent({
        event: 'runner_shutdown_instance',
        status: 'ok',
        user,
        props: { orgId: org.id, runnerId: runner.id },
      })
    }
  }, [isShutdown, error, org.id, runner.id, user])

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            className="inline-flex gap-4 items-center"
            variant="h3"
            weight="strong"
            theme="warn"
          >
            <Icon variant="CloudArrowDown" size="24" />
            Restart runner instance?
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Restarting
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="CloudArrowDown" />
            Restart instance
          </span>
        ),
        disabled: isLoading,
        onClick: () => mutate(),
        variant: 'primary' as const,
      }}
      onClose={handleClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to restart runner instance.'}
          </Banner>
        ) : null}
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Restart this runner instance.
          </Text>
          <Text variant="body" theme="neutral" className="leading-relaxed max-w-md">
            The runner VM will be restarted.
          </Text>
        </div>
      </div>
    </Modal>
  )
}
