import { useNavigate } from 'react-router'
import { useState } from 'react'
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
import { deployComponents } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'

export const DriftScanAllComponentsButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <DriftScanAllComponentsModal />
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

export const DriftScanAllComponentsModal = ({ ...props }: IModal) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const [isKickedOff, setIsKickedOff] = useState(false)

  const { mutate: execute, isPending: isLoading, error } = useMutation({
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
      const workflowId = result?.headers?.['x-nuon-install-workflow-id']
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
    <Modal
      heading={
        <Text
          className="!inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="info"
        >
          <Icon variant="ScanIcon" size="24" />
          Drift scan all {install.name} components?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Starting drift scan
          </span>
        ) : (
          'Drift scan components'
        ),
        disabled: isKickedOff || isLoading,
        onClick: () => {
          execute({ body: { plan_only: true } })
        },
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3 mb-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to deploy components'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          Are you sure you want to drift scan components?
        </Text>
        <Text variant="base">
          This aciton will preform a drift scan against the latest build of each
          component and the component deployments on your install.
        </Text>
      </div>
    </Modal>
  )
}
