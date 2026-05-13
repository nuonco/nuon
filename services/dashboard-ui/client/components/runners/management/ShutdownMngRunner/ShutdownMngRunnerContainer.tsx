import { useEffect } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { shutdownRunnerProcess } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import { ShutdownMngRunnerModal } from './ShutdownMngRunner'

interface IShutdownMngRunnerButton extends IButtonAsButton {
  runnerId: string
  processId: string
  showRunnerLabel?: boolean
}

interface IShutdownMngRunnerModalContainer extends IModal {
  runnerId: string
  processId: string
  showRunnerLabel?: boolean
}

export const ShutdownMngRunnerButton = ({ runnerId, processId, showRunnerLabel, ...props }: IShutdownMngRunnerButton) => {
  const { addModal } = useSurfaces()
  const label = showRunnerLabel ? 'Shutdown runner process' : 'Shutdown process'
  const modal = <ShutdownMngRunnerModalContainer runnerId={runnerId} processId={processId} showRunnerLabel={showRunnerLabel} />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="PowerIcon" />}
      {label}
      {props?.isMenuButton ? <Icon variant="PowerIcon" /> : null}
    </Button>
  )
}

export const ShutdownMngRunnerModalContainer = ({ runnerId, processId, showRunnerLabel, ...props }: IShutdownMngRunnerModalContainer) => {
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
    <ShutdownMngRunnerModal
      label={label}
      error={error}
      isLoading={isLoading}
      onConfirm={() => mutate()}
      onClose={() => removeModal(props.modalId)}
      {...props}
    />
  )
}
