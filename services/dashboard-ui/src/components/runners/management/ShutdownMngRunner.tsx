'use client'

import { useEffect } from 'react'
import { useAuth } from '@/hooks/use-auth'
import { shutdownMngRunner } from '@/actions/runners/shutdown-mng-runner'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { trackEvent } from '@/lib/segment-analytics'

export const ShutdownMngRunnerButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ShutdownMngRunnerModal />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="ArrowClockwise" />}
      Restart process
      {props?.isMenuButton ? <Icon variant="ArrowClockwise" /> : null}
    </Button>
  )
}

export const ShutdownMngRunnerModal = ({ ...props }: IModal) => {
  const { user } = useAuth()
  const { org } = useOrg()
  const { runner } = useRunner()
  const { removeModal } = useSurfaces()
  const runnerId = runner?.id

  const {
    data: isShutdown,
    error,
    execute,
    isLoading,
  } = useServerAction({ action: shutdownMngRunner })

  useServerActionToast({
    data: isShutdown,
    error,
    errorContent: <Text>Unable to restart managed runner process.</Text>,
    errorHeading: `Restart managed runner process failed`,
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: (
      <Text>Restart managed runner process initiated successfully.</Text>
    ),
    successHeading: `Restart managed runner process started`,
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
            Restart process?
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
            <Icon variant="ArrowClockwise" />
            Restart process
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
            {error?.error || 'Unable to restart managed runner process.'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Restart this managed runner process.
          </Text>
          <Text
            variant="body"
            theme="neutral"
            className="leading-relaxed max-w-md"
          >
            The managed runner will be gracefully restarted after completing any
            queued jobs.
          </Text>

          <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
            <li>All running jobs will be completed before restart</li>
            <li>Causes all jobs to queue while the runner restarts</li>
            <li>Any new version updates will be applied</li>
          </ul>
        </div>
      </div>
    </Modal>
  )
}
