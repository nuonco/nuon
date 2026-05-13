import { useEffect } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { shutdownRunnerInstance } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import { ShutdownInstanceModal } from './ShutdownInstance'

export const ShutdownInstanceButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ShutdownInstanceModalContainer />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="CloudArrowDownIcon" />}
      Restart instance
      {props?.isMenuButton ? <Icon variant="CloudArrowDownIcon" /> : null}
    </Button>
  )
}

export const ShutdownInstanceModalContainer = ({ ...props }: IModal) => {
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
    <ShutdownInstanceModal
      error={error}
      isLoading={isLoading}
      onConfirm={() => mutate()}
      onClose={() => removeModal(props.modalId)}
      {...props}
    />
  )
}
