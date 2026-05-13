import { useNavigate } from 'react-router'
import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { deprovisionInstall } from '@/lib'
import { DeprovisionModal } from './Deprovision'

interface IDeprovision {}

export const DeprovisionModalContainer = ({ ...props }: IDeprovision & Omit<IModal, 'onSubmit'>) => {
  const navigate = useNavigate()
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()

  const { mutate, isPending: isLoading, error } = useMutation({
    mutationFn: () =>
      deprovisionInstall({
        orgId: org.id,
        installId: install.id,
        body: {
          plan_only: false,
          error_behavior: 'abort',
        },
      }),
    onSuccess: (result) => {
      addToast(
        <Toast heading="Deprovision started" theme="success">
          <Text>Deprovision workflow started for {install.name}.</Text>
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
    onError: (error) => {
      addToast(
        <Toast heading="Deprovision failed" theme="error">
          <Text>Unable to start deprovision workflow for {install.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <DeprovisionModal
      installName={install.name}
      isPending={isLoading}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const DeprovisionButton = ({
  ...props
}: IDeprovision & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <DeprovisionModalContainer />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Deprovision install
      <Icon variant="ArrowDownIcon" />
    </Button>
  )
}
