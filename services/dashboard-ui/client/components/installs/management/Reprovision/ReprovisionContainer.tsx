import { useState } from 'react'
import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { RoleSelector } from '@/components/roles/RoleSelector'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { reprovisionInstall } from '@/lib'
import { ReprovisionModal } from './Reprovision'

interface IReprovision {}

export const ReprovisionModalContainer = ({ ...props }: IReprovision & Omit<IModal, 'onSubmit'>) => {
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
      const workflowId = result.data.workflow_id
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
    <ReprovisionModal
      installName={install.name}
      isPending={isLoading}
      error={error}
      onSubmit={() => mutate()}
      roleSelector={
        <RoleSelector
          installId={install?.id}
          value={selectedRole}
          onChange={setSelectedRole}
          name="role"
        />
      }
      {...props}
    />
  )
}

export const ReprovisionButton = ({
  ...props
}: IReprovision & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ReprovisionModalContainer />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Reprovision install
      <Icon variant="ArrowURightUpIcon" />
    </Button>
  )
}
