import { useNavigate } from 'react-router'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { reprovisionSandbox } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'

export const DriftScanSandboxButton = ({
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()

  const modal = <DriftScanSandboxModal />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="Scan" />}
      Drift scan sandbox
      {props?.isMenuButton ? <Icon variant="Scan" /> : null}
    </Button>
  )
}

export const DriftScanSandboxModal = ({
  ...props
}: IModal) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate: execute, isPending: isLoading, error } = useMutation({
    mutationFn: () =>
      reprovisionSandbox({
        body: { plan_only: true },
        installId: install.id,
        orgId: org.id,
      }),
    onSuccess: (result) => {
      trackEvent({
        event: 'install_sandbox_drift_scan',
        user,
        status: 'ok',
        props: { orgId: org.id, installId: install.id },
      })
      addToast(
        <Toast heading="Drift scan initiated" theme="success">
          <Text>Sandbox drift scan workflow has been started successfully.</Text>
        </Toast>
      )
      removeModal(props.modalId)
      const workflowId = result?.headers?.['x-nuon-install-workflow-id']
      if (workflowId) {
        navigate(`/${org.id}/installs/${install.id}/workflows/${workflowId}`)
      } else {
        navigate(`/${org.id}/installs/${install.id}/workflows`)
      }
    },
    onError: (err: any) => {
      trackEvent({
        event: 'install_sandbox_drift_scan',
        user,
        status: 'error',
        props: { orgId: org.id, installId: install.id, err: err?.error },
      })
      addToast(
        <Toast heading="Sandbox drift scan failed" theme="error">
          <Text>Failed to initiate sandbox drift scan. Please try again.</Text>
        </Toast>
      )
    },
  })

  const handleDriftScan = () => {
    execute()
  }

  const handleClose = () => {
    removeModal(props.modalId)
  }

  return (
    <Modal
      heading="Drift scan sandbox?"
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Scanning
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="Scan" />
            Drift scan sandbox
          </span>
        ),
        disabled: isLoading,
        onClick: handleDriftScan,
        variant: 'primary' as const,
      }}
      onClose={handleClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to drift scan sandbox.'}
          </Banner>
        ) : null}

        <Text variant="body" className="leading-relaxed">
          Are you sure you want to drift scan this sandbox?
        </Text>
      </div>
    </Modal>
  )
}
