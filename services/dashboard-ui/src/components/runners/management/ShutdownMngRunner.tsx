'use client'

import { useEffect } from 'react'
import { useUser } from '@auth0/nextjs-auth0'
import { shutdownMngRunner } from '@/actions/runners/shutdown-mng-runner'
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

export const ShutdownMngRunnerButton = ({
  runnerId,
  ...props
}: IButtonAsButton & {
  runnerId: string
}) => {
  const { addModal } = useSurfaces()
  const modal = <ShutdownMngRunnerModal runnerId={runnerId} />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="ArrowClockwise" />}
      Shutdown runner
      {props?.isMenuButton ? <Icon variant="ArrowClockwise" /> : null}
    </Button>
  )
}

export const ShutdownMngRunnerModal = ({
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
  } = useServerAction({ action: shutdownMngRunner })

  useServerActionToast({
    data: isShutdown,
    error,
    errorContent: <Text>Unable to shutdown managed runner.</Text>,
    errorHeading: `Managed runner shutdown failed`,
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: (
      <Text>Managed runner shutdown initiated successfully.</Text>
    ),
    successHeading: `Managed runner shutdown started`,
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
        event: 'managed_runner_shutdown',
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
        event: 'managed_runner_shutdown',
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
            <Icon variant="ArrowClockwise" size="24" />
            Shutdown runner?
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
            <Icon variant="ArrowClockwise" />
            Shutdown runner
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
            {error?.error || 'Unable to shutdown managed runner.'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Shutdown this managed runner.
          </Text>
          <Text
            variant="body"
            theme="neutral"
            className="leading-relaxed max-w-md"
          >
            The managed runner will be gracefully shut down and all resources
            will be cleaned up automatically.
          </Text>

          <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
            <li>All running jobs will be completed before shutdown</li>
            <li>The runner will be removed from the managed pool</li>
            <li>Resources will be automatically cleaned up</li>
          </ul>
        </div>
      </div>
    </Modal>
  )
}
