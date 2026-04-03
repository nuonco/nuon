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
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { shutdownRunnerProcess } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'

interface IShutdownMngRunnerButton extends IButtonAsButton {
  runnerId: string
  processId: string
  showRunnerLabel?: boolean
}

interface IShutdownMngRunnerModal extends IModal {
  runnerId: string
  processId: string
  showRunnerLabel?: boolean
}

export const ShutdownMngRunnerButton = ({ runnerId, processId, showRunnerLabel, ...props }: IShutdownMngRunnerButton) => {
  const { addModal } = useSurfaces()
  const label = showRunnerLabel ? 'Shutdown runner process' : 'Shutdown process'
  const modal = <ShutdownMngRunnerModal runnerId={runnerId} processId={processId} showRunnerLabel={showRunnerLabel} />
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

export const ShutdownMngRunnerModal = ({ runnerId, processId, showRunnerLabel, ...props }: IShutdownMngRunnerModal) => {
  const label = showRunnerLabel ? 'Shutdown runner process' : 'Shutdown process'
  const { user } = useAuth()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

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
        shutdownType: 'graceful',
        orgId: org.id,
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Shutdown managed runner process started" theme="success">
          <Text>Shutdown managed runner process initiated successfully.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Shutdown managed runner process failed" theme="error">
          <Text>Unable to restart managed runner process.</Text>
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
        event: 'mng_process_shutdown',
        status: 'error',
        user,
        props: { orgId: org.id, runnerId, processId, err: error?.error },
      })
    }
    if (shutdown) {
      trackEvent({
        event: 'mng_process_shutdown',
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
            {error?.error || 'Unable to shutdown managed runner process.'}
          </Banner>
        ) : null}
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Shutdown this managed runner instance.
          </Text>
          <Text variant="body" theme="neutral" className="leading-relaxed max-w-md">
            This will destroy the managed runner instance. A new instance will be
            provisioned automatically to replace it.
          </Text>
          <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
            <li>The VM instance will be terminated</li>
            <li>A new instance will be provisioned with the latest version</li>
            <li>All local state will be lost</li>
          </ul>
        </div>
      </div>
    </Modal>
  )
}
