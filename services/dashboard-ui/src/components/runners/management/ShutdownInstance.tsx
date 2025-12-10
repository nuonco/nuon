'use client'

import { useEffect } from 'react'
import { useUser } from '@auth0/nextjs-auth0'
import { shutdownInstance } from '@/actions/runners/shutdown-instance'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { trackEvent } from '@/lib/segment-analytics'

export const ShutdownInstanceButton = ({
  runnerId,
  ...props
}: IButtonAsButton & {
  runnerId: string
}) => {
  const { addModal } = useSurfaces()
  const modal = <ShutdownInstanceModal runnerId={runnerId} />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="CloudArrowDown" />}
      Shutdown instance
      {props?.isMenuButton ? <Icon variant="CloudArrowDown" /> : null}
    </Button>
  )
}

export const ShutdownInstanceModal = ({
  runnerId,
  ...props
}: IModal & {
  runnerId: string
}) => {
  const { user } = useUser()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()

  const {
    data: isShutdown,
    error,
    execute,
    isLoading,
  } = useServerAction({ action: shutdownInstance })

  useServerActionToast({
    data: isShutdown,
    error,
    errorContent: <Text>Unable to shutdown runner instance.</Text>,
    errorHeading: `Runner instance shutdown failed`,
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: <Text>Runner instance shutdown initiated successfully.</Text>,
    successHeading: `Runner instance shutdown started`,
  })

  const handleClose = () => {
    removeModal(props.modalId)
  }

  const handleShutdown = () => {
    execute({
      runnerId,
      orgId: org.id,
    })
  }

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'runner_shutdown_instance',
        status: 'error',
        user,
        props: {
          orgId: org.id,
          runnerId,
          err: error?.error,
        },
      })
    }

    if (isShutdown) {
      trackEvent({
        event: 'runner_shutdown_instance',
        status: 'ok',
        user,
        props: {
          orgId: org.id,
          runnerId,
        },
      })
    }
  }, [isShutdown, error, org.id, runnerId, user])

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
            Shutdown runner instance?
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Shutting down
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="CloudArrowDown" />
            Shutdown instance
          </span>
        ),
        disabled: isLoading,
        onClick: handleShutdown,
        variant: 'primary' as const,
      }}
      onClose={handleClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error?.error ? (
          <Banner theme="error">
            {error?.error || 'Unable to shutdown runner instance.'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Shutdown this runner instance.
          </Text>
          <Text variant="body" theme="neutral" className="leading-relaxed max-w-md">
            The runner VM will be shutdown and restarted.
          </Text>
        </div>
      </div>
    </Modal>
  )
}
