import { useNavigate } from 'react-router'
import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Badge } from '@/components/common/Badge'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { RoleSelector } from '@/components/roles/RoleSelector'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { deployComponents } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import { DeployAllComponentsModal } from './DeployAllComponents'

export const DeployAllComponentsModalContainer = ({
  ...props
}: Omit<IModal, 'onSubmit'>) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
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
        event: 'components_deploy',
        status: 'ok',
        user,
        props: {
          installId: install.id,
          orgId: org.id,
        },
      })
      addToast(
        <Toast heading={<span className="inline-flex items-center gap-1.5"><Badge variant="code" size="md">{install.name}</Badge> deploy started</span>} theme="info" />
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
    onError: (err: any) => {
      trackEvent({
        event: 'components_deploy',
        status: 'error',
        user,
        props: {
          installId: install.id,
          orgId: org.id,
          err: err?.error,
        },
      })
      addToast(
        <Toast heading={<span className="inline-flex items-center gap-1.5"><Badge variant="code" size="md">{install.name}</Badge> deploy failed</span>} theme="error" />
      )
    },
  })

  return (
    <DeployAllComponentsModal
      installName={install.name}
      isPending={isPending}
      isKickedOff={isKickedOff}
      error={error as any}
      onSubmit={({ role }) =>
        execute({ body: { plan_only: false, ...(role && { role }) } })
      }
      roleSelector={({ value, onChange }) => (
        <RoleSelector
          installId={install?.id}
          operationType="deploy"
          value={value}
          onChange={onChange}
          name="role"
        />
      )}
      {...props}
    />
  )
}

export const DeployAllComponentsButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <DeployAllComponentsModalContainer />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Deploy components <Icon variant="CloudArrowUpIcon" />
    </Button>
  )
}
