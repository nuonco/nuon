import { useNavigate } from 'react-router'
import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useAuth } from '@/hooks/use-auth'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { forgetComponent } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import type { TComponent } from '@/types'
import { ForgetComponentModal } from './Forget'

interface IForgetComponentModalContainer extends IModal {
  component: TComponent
}

export const ForgetComponentModalContainer = ({
  component,
  ...props
}: IForgetComponentModalContainer) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()

  const { mutate: execute, isPending: isLoading, error } = useMutation({
    mutationFn: () =>
      forgetComponent({
        orgId: org.id,
        installId: install.id,
        componentId: component.id,
      }),
    onSuccess: () => {
      trackEvent({
        event: 'component_forget',
        status: 'ok',
        user,
        props: {
          orgId: org.id,
          installId: install.id,
          componentId: component.id,
        },
      })
      addToast(
        <Toast heading="Component forgotten" theme="success">
          <Text>Component {component.name} has been forgotten.</Text>
        </Toast>
      )
      removeModal(props.modalId)
      navigate(`/${org.id}/installs/${install.id}/components`)
    },
    onError: (err: any) => {
      trackEvent({
        event: 'component_forget',
        status: 'error',
        user,
        props: {
          orgId: org.id,
          installId: install.id,
          componentId: component.id,
          err: err?.error,
        },
      })
      addToast(
        <Toast heading="Forget failed" theme="error">
          <Text>Unable to forget {component.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <ForgetComponentModal
      componentName={component.name}
      isLoading={isLoading}
      error={error}
      onConfirm={() => execute()}
      {...props}
    />
  )
}

interface IForgetComponentButton extends IButtonAsButton {
  component: TComponent
}

export const ForgetComponentButton = ({
  component,
  ...props
}: IForgetComponentButton) => {
  const { addModal } = useSurfaces()
  const modal = <ForgetComponentModalContainer component={component} />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      variant="ghost"
      {...props}
      className="!text-red-800 dark:!text-red-500"
    >
      {props?.isMenuButton ? null : <Icon variant="Trash" />}
      Forget component
      {props?.isMenuButton ? <Icon variant="Trash" /> : null}
    </Button>
  )
}
