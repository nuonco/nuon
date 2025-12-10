'use client'

import { useEffect, useState } from 'react'
import { useUser } from '@auth0/nextjs-auth0'
import { shutdownRunner } from '@/actions/runners/shutdown-runner'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { trackEvent } from '@/lib/segment-analytics'

export const ShutdownRunnerButton = ({
  runnerId,
  ...props
}: IButtonAsButton & {
  runnerId: string
}) => {
  const { addModal } = useSurfaces()
  const modal = <ShutdownRunnerModal runnerId={runnerId} />
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

export const ShutdownRunnerModal = ({
  runnerId,
  ...props
}: IModal & {
  runnerId: string
}) => {
  const { user } = useUser()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()

  const [force, setForce] = useState(false)

  const {
    data: isShutdown,
    error,
    execute,
    isLoading,
  } = useServerAction({ action: shutdownRunner })

  useServerActionToast({
    data: isShutdown,
    error,
    errorContent: <Text>Unable to shutdown runner.</Text>,
    errorHeading: `Runner shutdown failed`,
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: <Text>Runner shutdown initiated successfully.</Text>,
    successHeading: `Runner shutdown started`,
  })

  const handleClose = () => {
    setForce(false)
    removeModal(props.modalId)
  }

  const handleShutdown = () => {
    execute({
      runnerId,
      orgId: org.id,
      force,
    })
  }

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'runner_shutdown',
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
        event: 'runner_shutdown',
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
            {error?.error || 'Unable to shutdown runner.'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Shutdown this runner gracefully.
          </Text>
          <Text variant="body" theme="neutral" className="leading-relaxed max-w-md">
            The runner will make a best effort to shut down after any queued jobs are complete.
          </Text>

          <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
            <li>Causes all jobs to queue while the runner restarts</li>
            <li>Any new version updates will be applied</li>
            <li>All local state will be refreshed</li>
          </ul>

          <div className="flex items-start">
            <CheckboxInput
              checked={force}
              onChange={(e) => setForce(e.target.checked)}
              className="mt-1.5"
              labelProps={{
                className: "hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !p-2 gap-4 max-w-none !items-start",
                labelText: (
                  <div className="flex flex-col gap-1">
                    <Text variant="base" weight="stronger">
                      Destroy instance
                    </Text>
                    <Text variant="subtext" theme="neutral">
                      Immediately shutdown the runner, terminating any in-flight jobs. This has the potential for loss of state.
                    </Text>
                  </div>
                ),
              }}
            />
          </div>
        </div>
      </div>
    </Modal>
  )
}
