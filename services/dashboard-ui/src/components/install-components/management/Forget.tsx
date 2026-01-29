'use client'

import { usePathname, useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { forgetComponent } from '@/actions/installs/forget-component'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useAuth } from '@/hooks/use-auth'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { trackEvent } from '@/lib/segment-analytics'
import type { TComponent } from '@/types'

interface IForgetComponentModal extends IModal {
  component: TComponent
}

export const ForgetComponentModal = ({
  component,
  ...props
}: IForgetComponentModal) => {
  const router = useRouter()
  const { user } = useAuth()
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()
  const [confirmName, setConfirmName] = useState('')

  const { data, error, isLoading, execute } = useServerAction({
    action: forgetComponent,
  })

  useServerActionToast({
    data,
    error,
    errorContent: <Text>Unable to forget {component.name}.</Text>,
    errorHeading: 'Forget failed',
    onSuccess: () => {
      router.push(`/${org.id}/installs/${install.id}/components`)
      removeModal(props.modalId)
    },
    successContent: (
      <Text>Component {component.name} has been forgotten.</Text>
    ),
    successHeading: 'Component forgotten',
  })

  const isConfirmValid = confirmName === component.name
  const canForget = isConfirmValid && !isLoading

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'component_forget',
        status: 'error',
        user,
        props: {
          orgId: org.id,
          installId: install.id,
          componentId: component.id,
          err: error?.error,
        },
      })
    }

    if (data) {
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
    }
  }, [data, error, org.id, install.id, component.id, user])

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
          execute({
            orgId: org.id,
            installId: install.id,
            componentId: component.id,
          })
        },
        disabled: !canForget,
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to forget component.'}
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
