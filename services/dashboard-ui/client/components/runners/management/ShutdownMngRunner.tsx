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
import { shutdownMngRunner } from '@/lib'
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
  const { addToast } = useToast()

  const {
    data: isShutdown,
    error,
    mutate,
    isPending: isLoading,
  } = useMutation({
    mutationFn: () => shutdownMngRunner({ runnerId: runner.id, orgId: org.id }),
    onSuccess: () => {
      addToast(
        <Toast heading="Restart managed runner process started" theme="success">
          <Text>Restart managed runner process initiated successfully.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Restart managed runner process failed" theme="error">
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
        event: 'managed_runner_shutdown',
        status: 'error',
        user,
        props: { orgId: org.id, runnerId: runner.id, err: error?.error },
      })
    }
    if (isShutdown) {
      trackEvent({
        event: 'managed_runner_shutdown',
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
        onClick: () => mutate(),
        variant: 'primary' as const,
      }}
      onClose={handleClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to restart managed runner process.'}
          </Banner>
        ) : null}
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Restart this managed runner process.
          </Text>
          <Text variant="body" theme="neutral" className="leading-relaxed max-w-md">
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
