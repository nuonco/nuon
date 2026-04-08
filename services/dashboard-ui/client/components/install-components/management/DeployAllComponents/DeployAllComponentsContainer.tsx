import { useNavigate } from 'react-router'
import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
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
        <Toast heading={`${install.name} component deployments started`} theme="success">
          <Text>Deploy all components workflow was created.</Text>
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
        <Toast heading="Deployments failed" theme="error">
          <Text>Unable to deploy all components to {install.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <DeployAllComponentsModal
      installName={install.name}
      isPending={isPending}
      isKickedOff={isKickedOff}
      error={error as any}
      onSubmit={() => execute({ body: { plan_only: false } })}
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
