import { useEffect } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { type IButtonAsButton } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { shutdownRunnerProcess } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import {
  ShutdownRunnerModal as ShutdownRunnerModalComponent,
  ShutdownRunnerButton as ShutdownRunnerButtonComponent,
} from './ShutdownRunner'

interface IShutdownRunnerContainer {
  runnerId: string
  processId: string
  showRunnerLabel?: boolean
}

export const ShutdownRunnerButton = ({ runnerId, processId, showRunnerLabel, ...props }: IShutdownRunnerContainer & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ShutdownRunnerModal runnerId={runnerId} processId={processId} showRunnerLabel={showRunnerLabel} />
  return (
    <ShutdownRunnerButtonComponent
      showRunnerLabel={showRunnerLabel}
      onOpenModal={() => addModal(modal)}
      {...props}
    />
  )
}

export const ShutdownRunnerModal = ({ runnerId, processId, showRunnerLabel, ...props }: IShutdownRunnerContainer & Omit<IModal, 'onSubmit'>) => {
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
    mutationFn: (force: boolean) =>
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
    <ShutdownRunnerModalComponent
      showRunnerLabel={showRunnerLabel}
      isPending={isLoading}
      error={error}
      onSubmit={(force) => mutate(force)}
      onClose={() => removeModal(props.modalId)}
      {...props}
    />
  )
}
