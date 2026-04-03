import { useEffect, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { shutdownRunnerProcess } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'

interface IShutdownRunnerButton extends IButtonAsButton {
  runnerId: string
  processId: string
  showRunnerLabel?: boolean
}

interface IShutdownRunnerModal extends IModal {
  runnerId: string
  processId: string
  showRunnerLabel?: boolean
}

export const ShutdownRunnerButton = ({ runnerId, processId, showRunnerLabel, ...props }: IShutdownRunnerButton) => {
  const { addModal } = useSurfaces()
  const label = showRunnerLabel ? 'Shutdown runner process' : 'Shutdown process'
  const modal = <ShutdownRunnerModal runnerId={runnerId} processId={processId} showRunnerLabel={showRunnerLabel} />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="Power" />}
      {label}
      {props?.isMenuButton ? <Icon variant="Power" /> : null}
    </Button>
  )
}

export const ShutdownRunnerModal = ({ runnerId, processId, showRunnerLabel, ...props }: IShutdownRunnerModal) => {
  const label = showRunnerLabel ? 'Shutdown runner process' : 'Shutdown process'
  const { user } = useAuth()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const [force, setForce] = useState(false)

  const {
    data: shutdown,
    error,
    mutate,
    isPending: isLoading,
  } = useMutation({
    mutationFn: () =>
      shutdownRunnerProcess({
        runnerId,
        processId,
        shutdownType: force ? 'force' : 'graceful',
        orgId: org.id,
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Shutdown runner process started" theme="success">
          <Text>Shutdown runner process initiated successfully.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Shutdown runner process failed" theme="error">
          <Text>Unable to restart runner process.</Text>
        </Toast>
      )
    },
  })

  const handleClose = () => {
    setForce(false)
    removeModal(props.modalId)
  }

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'runner_process_shutdown',
        status: 'error',
        user,
        props: { orgId: org.id, runnerId, processId, err: error?.error },
      })
    }
    if (shutdown) {
      trackEvent({
        event: 'runner_process_shutdown',
        status: 'ok',
        user,
        props: { orgId: org.id, runnerId, processId },
      })
    }
  }, [shutdown, error, org.id, runnerId, processId, user])

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
            <Icon variant="Power" size="24" />
            {label}?
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
            <Icon variant="Power" />
            {label}
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
            {error?.error || 'Unable to shutdown runner process.'}
          </Banner>
        ) : null}
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Shutdown this runner process.
          </Text>
          <Text variant="body" theme="neutral" className="leading-relaxed max-w-md">
            This will terminate the container and restart the process. The runner
            will make a best effort to complete any queued jobs before shutting down.
          </Text>
          <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
            <li>Causes all jobs to queue while the process restarts</li>
            <li>Any new version updates will be applied on restart</li>
            <li>All local state will be refreshed</li>
          </ul>
          <div className="flex items-start">
            <CheckboxInput
              checked={force}
              onChange={(e) => setForce(e.target.checked)}
              className="mt-1.5"
              labelProps={{
                className:
                  'hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !p-2 gap-4 max-w-none !items-start',
                labelText: (
                  <div className="flex flex-col gap-1">
                    <Text variant="base" weight="stronger">
                      Force shutdown
                    </Text>
                    <Text variant="subtext" theme="neutral">
                      Immediately shutdown the runner, terminating any in-flight
                      jobs. This has the potential for loss of state.
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
