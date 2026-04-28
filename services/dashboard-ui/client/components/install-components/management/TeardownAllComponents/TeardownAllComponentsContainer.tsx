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
import { teardownComponents } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import { TeardownAllComponentsModal } from './TeardownAllComponents'

export const TeardownAllComponentsModalContainer = ({
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
    mutationFn: (params: { body: Parameters<typeof teardownComponents>[0]['body'] }) =>
      teardownComponents({
        body: params.body,
        installId: install.id,
        orgId: org.id,
      }),
    onSuccess: (result) => {
      trackEvent({
        event: 'components_teardown',
        status: 'ok',
        user,
        props: {
          installId: install.id,
          orgId: org.id,
        },
      })
      addToast(
        <Toast heading={`${install.name} component teardowns started`} theme="success">
          <Text>Teardown all components workflow was created.</Text>
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
    onError: (err: any) => {
      trackEvent({
        event: 'components_teardown',
        status: 'error',
        user,
        props: {
          installId: install.id,
          orgId: org.id,
          err: err?.error,
        },
      })
      addToast(
        <Toast heading="Teardowns failed" theme="error">
          <Text>Unable to teardown all components to {install.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <TeardownAllComponentsModal
      installName={install.name}
      isPending={isPending}
      isKickedOff={isKickedOff}
      error={error as any}
      onSubmit={() => execute({ body: { plan_only: false } })}
      {...props}
    />
  )
}

export const TeardownAllComponentsButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <TeardownAllComponentsModalContainer />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
      variant="danger"
    >
      Teardown components <Icon variant="CloudArrowDownIcon" />
    </Button>
  )
}
