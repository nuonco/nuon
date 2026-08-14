import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useAuth } from '@/hooks/use-auth'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { recoverHelmRelease } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import type { TComponent } from '@/types'
import { RecoverHelmReleaseModal } from './RecoverHelmRelease'

interface IRecoverHelmRelease {
  component: TComponent
}

export const RecoverHelmReleaseButton = ({
  component,
  ...props
}: IRecoverHelmRelease & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <RecoverHelmReleaseModalContainer component={component} />

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      {props?.isMenuButton ? null : <Icon variant="ArrowCounterClockwiseIcon" />}
      Recover Helm release
      {props?.isMenuButton ? <Icon variant="ArrowCounterClockwiseIcon" /> : null}
    </Button>
  )
}

export const RecoverHelmReleaseModalContainer = ({
  component,
  ...props
}: IRecoverHelmRelease & Omit<IModal, 'onSubmit'>) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const {
    mutate: execute,
    isPending,
    error,
  } = useMutation({
    mutationFn: () =>
      recoverHelmRelease({
        componentId: component.id,
        installId: install.id,
        orgId: org.id,
      }),
    onSuccess: (result) => {
      trackEvent({
        event: 'component_helm_release_recover',
        status: 'ok',
        user,
        props: { orgId: org.id, installId: install.id, componentId: component.id },
      })
      addToast(
        <Toast heading="Recovering Helm release" theme="info">
          <Text>
            Recovering the Helm release for {component.name} on {install.name}. This may take a
            few minutes.
          </Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['install-component'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      removeModal(props.modalId)

      const workflowId = result.data.workflow_id
      navigate(
        workflowId
          ? `/${org.id}/installs/${install.id}/workflows/${workflowId}`
          : `/${org.id}/installs/${install.id}/workflows`
      )
    },
    onError: (err: any) => {
      trackEvent({
        event: 'component_helm_release_recover',
        status: 'error',
        user,
        props: {
          orgId: org.id,
          installId: install.id,
          componentId: component.id,
          err: err?.error,
        },
      })
    },
  })

  return (
    <RecoverHelmReleaseModal
      componentName={component.name}
      isPending={isPending}
      error={error as any}
      onSubmit={() => execute()}
      onClose={() => removeModal(props.modalId)}
      {...props}
    />
  )
}
