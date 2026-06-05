import { useNavigate } from 'react-router'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Badge } from '@/components/common/Badge'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { reprovisionSandbox } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import { DriftScanSandboxModal } from './DriftScanSandbox'

export const DriftScanSandboxModalContainer = ({
  ...props
}: Omit<IModal, 'onSubmit'>) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate: execute, isPending, error } = useMutation({
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
        <Toast heading="Sandbox drift scan started" theme="info">
          <Text>Scanning <Badge variant="code" size="md">{install.name}</Badge> sandbox for drift.</Text>
        </Toast>
      )
      removeModal(props.modalId)
      const workflowId = result.data.workflow_id
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
          <Text>Unable to start drift scan for <Badge variant="code" size="md">{install.name}</Badge> sandbox.</Text>
        </Toast>
      )
    },
  })

  return (
    <DriftScanSandboxModal
      isPending={isPending}
      error={error}
      onSubmit={() => execute()}
      onClose={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const DriftScanSandboxButton = ({
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()

  const modal = <DriftScanSandboxModalContainer />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="ScanIcon" />}
      Drift scan sandbox
      {props?.isMenuButton ? <Icon variant="ScanIcon" /> : null}
    </Button>
  )
}
