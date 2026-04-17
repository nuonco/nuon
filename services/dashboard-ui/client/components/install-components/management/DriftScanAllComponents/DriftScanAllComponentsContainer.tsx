import { useNavigate } from 'react-router'
import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { deployComponents } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import { DriftScanAllComponentsModal } from './DriftScanAllComponents'

export const DriftScanAllComponentsModalContainer = ({
  ...props
}: Omit<IModal, 'onSubmit'>) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const [isKickedOff, setIsKickedOff] = useState(false)

  const { mutate: execute, isPending, error } = useMutation({
    mutationFn: (params: { body: Parameters<typeof deployComponents>[0]['body'] }) =>
      deployComponents({
        body: params.body,
        installId: install.id,
        orgId: org.id,
      }),
    onSuccess: (result) => {
      trackEvent({
        event: 'components_drift_scan',
        status: 'ok',
        user,
        props: {
          installId: install.id,
          orgId: org.id,
        },
      })
      addToast(
        <Toast heading={`${install.name} component drift scan started`} theme="success">
          <Text>Components drift scan workflow was created.</Text>
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
        event: 'components_drift_scan',
        status: 'error',
        user,
        props: {
          installId: install.id,
          orgId: org.id,
          err: err?.error,
        },
      })
      addToast(
        <Toast heading="Drift scan failed" theme="error">
          <Text>Unable to deploy all components to {install.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <DriftScanAllComponentsModal
      installName={install.name}
      isPending={isPending}
      isKickedOff={isKickedOff}
      error={error as any}
      onSubmit={() => execute({ body: { plan_only: true } })}
      {...props}
    />
  )
}

export const DriftScanAllComponentsButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <DriftScanAllComponentsModalContainer />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Drift scan components <Icon variant="ScanIcon" />
    </Button>
  )
}
