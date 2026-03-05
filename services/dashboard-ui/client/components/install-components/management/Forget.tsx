import { useNavigate } from 'react-router'
import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useAuth } from '@/hooks/use-auth'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { forgetComponent } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import type { TComponent } from '@/types'

interface IForgetComponentModal extends IModal {
  component: TComponent
}

export const ForgetComponentModal = ({
  component,
  ...props
}: IForgetComponentModal) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()
  const [confirmName, setConfirmName] = useState('')

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

  const isConfirmValid = confirmName === component.name
  const canForget = isConfirmValid && !isLoading

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="error"
        >
          <Icon variant="Trash" size="24" />
          Forget {component.name}
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Forgetting...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="Trash" />
            Forget component
          </span>
        ),
        onClick: () => {
          execute()
        },
        disabled: !canForget,
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to forget component.'}
          </Banner>
        ) : null}

        <Banner theme="warn">
          <Text variant="body">
            <strong>Warning:</strong> This should only be used in cases where a
            component was broken in an unordinary way and needs to be manually
            removed.
          </Text>
        </Banner>

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Text variant="base" weight="strong">
              Are you sure you want to forget {component.name}?
            </Text>
            <Text variant="body" theme="neutral">
              This action will remove the component and cannot be undone.
            </Text>
          </div>

          <div className="flex flex-col gap-3">
            <Text variant="body">You should only do this after you have:</Text>
            <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
              <li>Successfully tore down the component</li>
              <li>Verified no infrastructure remains in the cloud account</li>
              <li>Confirmed all dependencies are handled</li>
            </ul>
          </div>

          <div className="flex flex-col gap-2">
            <Text variant="body">
              To verify, type{' '}
              <span className="font-mono font-medium text-red-800 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-1 py-0.5 rounded">
                {component.name}
              </span>{' '}
              below.
            </Text>
            <Input
              id="confirm-component-name"
              placeholder="component name"
              type="text"
              value={confirmName}
              onChange={(e) => setConfirmName(e.target.value)}
              error={confirmName.length > 0 && !isConfirmValid}
              errorMessage={
                confirmName.length > 0 && !isConfirmValid
                  ? "Component name doesn't match"
                  : undefined
              }
            />
          </div>
        </div>
      </div>
    </Modal>
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
  const modal = <ForgetComponentModal component={component} />

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
