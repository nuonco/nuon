import { useNavigate } from 'react-router'
import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
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
import { teardownComponents } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'

export const TeardownAllComponentsButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <TeardownAllComponentsModal />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
      className="!text-red-600 dark:!text-red-400"
    >
      Teardown components <Icon variant="CloudArrowDownIcon" />
    </Button>
  )
}

export const TeardownAllComponentsModal = ({ ...props }: IModal) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const [isKickedOff, setIsKickedOff] = useState(false)

  const { mutate: execute, isPending: isLoading, error } = useMutation({
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
      const workflowId = result?.headers?.['x-nuon-install-workflow-id']
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
    <Modal
      heading={
        <Text
          className="!inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="error"
        >
          <Icon variant="CloudArrowDownIcon" size="24" />
          Teardown all {install.name} components?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Starting teardown
          </span>
        ) : (
          'Teardown components'
        ),
        disabled: isKickedOff || isLoading,
        onClick: () => {
          execute({ body: { plan_only: false } })
        },
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3 mb-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to teardown components'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          Are you sure you want to teardown all components?
        </Text>
        <Text variant="base">
          This aciton will remove all running component deployments from your
          install.
        </Text>
      </div>
    </Modal>
  )
}
