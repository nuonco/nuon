import { useState } from 'react'
import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { RoleSelector } from '@/components/roles/RoleSelector'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { reprovisionInstall } from '@/lib'

interface IReprovision {}

export const ReprovisionModal = ({ ...props }: IReprovision & IModal) => {
  const navigate = useNavigate()
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const [selectedRole, setSelectedRole] = useState<string>('')

  const { mutate, isPending: isLoading, error } = useMutation({
    mutationFn: () =>
      reprovisionInstall({
        orgId: org.id,
        installId: install.id,
        body: {
          plan_only: false,
          ...(selectedRole && { role: selectedRole }),
        },
      }),
    onSuccess: (result) => {
      addToast(
        <Toast heading={`${install.name} reprovision was started`} theme="success">
          <Text>Reprovisioning {install.name} workflow was created.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      removeModal(props.modalId)
      const workflowId = result?.headers?.['x-nuon-install-workflow-id']
      if (workflowId) {
        navigate(`/${org.id}/installs/${install.id}/workflows/${workflowId}`)
      } else {
        navigate(`/${org.id}/installs/${install.id}/workflows`)
      }
    },
    onError: (error) => {
      addToast(
        <Toast heading="Reprovision not started" theme="error">
          <Text>Unable to reprovision {install.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
          theme="info"
        >
          <Icon variant="ArrowURightUp" size="24" />
          Reprovision install?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Starting reprovision
          </span>
        ) : (
          'Reprovision install'
        ),
        onClick: () => mutate(),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-1">
        {error ? (
          <Banner theme="error">
            {error?.error ||
              'An error happened, please refresh the page and try again.'}
          </Banner>
        ) : null}
        <div className="flex flex-col gap-1">
          <Text variant="base" weight="strong">
            Are you sure you want to reprovision {}?
          </Text>
          <Text variant="base">
            Reprovisioning will recreate all resources and deploy all components
            again.
          </Text>
        </div>

        <RoleSelector
          installId={install?.id}
          value={selectedRole}
          onChange={setSelectedRole}
          name="role"
        />
      </div>
    </Modal>
  )
}

export const ReprovisionButton = ({
  ...props
}: IReprovision & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ReprovisionModal />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Reprovision install
      <Icon variant="ArrowURightUp" />
    </Button>
  )
}
