'use client'

import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { useUser } from '@auth0/nextjs-auth0'
import { teardownComponent } from '@/actions/installs/teardown-component'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TComponent } from '@/types'
import { trackEvent } from '@/lib/segment-analytics'

export const TeardownComponentButton = ({
  component,
  ...props
}: IButtonAsButton & {
  component: TComponent
}) => {
  const { addModal } = useSurfaces()
  const modal = <TeardownComponentModal component={component} />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
      className="!text-red-600 dark:!text-red-400"
    >
      {props?.isMenuButton ? null : <Icon variant="CloudArrowDownIcon" />}
      Teardown component
      {props?.isMenuButton ? <Icon variant="CloudArrowDownIcon" /> : null}
    </Button>
  )
}

export const TeardownComponentModal = ({
  component,
  ...props
}: IModal & {
  component: TComponent
}) => {
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()

  const [confirmName, setConfirmName] = useState('')

  const {
    data: teardown,
    error,
    execute,
    isLoading,
    headers,
  } = useServerAction({ action: teardownComponent })

  useServerActionToast({
    data: teardown,
    error,
    errorContent: <Text>Unable to teardown {component.name} component.</Text>,
    errorHeading: `Component teardown failed`,
    onSuccess: () => {
      removeModal(props.modalId)
      if (headers?.['x-nuon-install-workflow-id']) {
        router.push(
          `/${org.id}/installs/${install.id}/workflows/${headers['x-nuon-install-workflow-id']}`
        )
      } else {
        router.push(`/${org.id}/installs/${install.id}/workflows`)
      }
    },
    successContent: <Text>Teardown for {component.name} was started.</Text>,
    successHeading: `${component.name} teardown started`,
  })

  const handleClose = () => {
    setConfirmName('')
    removeModal(props.modalId)
  }

  const handleTeardown = () => {
    execute({
      body: {
        plan_only: false,
        error_behavior: 'continue',
      },
      componentId: component.id,
      installId: install.id,
      orgId: org.id,
    })
  }

  const isConfirmValid = confirmName === component.name
  const canTeardown = isConfirmValid && !isLoading

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'component_teardown',
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

    if (teardown) {
      trackEvent({
        event: 'component_teardown',
        status: 'ok',
        user,
        props: {
          orgId: org.id,
          installId: install.id,
          componentId: component.id,
        },
      })
    }
  }, [teardown, error, org.id, install.id, component.id, user])

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            className="inline-flex gap-4 items-center"
            variant="h3"
            weight="strong"
            theme="error"
          >
            <Icon variant="Warning" size="24" />
            Teardown {component.name} component
          </Text>
          <Text
            variant="body"
            className="text-cool-grey-600 dark:text-cool-grey-400"
          >
            This will remove all running deployments for this component
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Tearing down
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="CloudArrowDownIcon" />
            Teardown component
          </span>
        ),
        disabled: !canTeardown,
        onClick: handleTeardown,
        variant: 'danger' as const,
      }}
      onClose={handleClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error?.error ? (
          <Banner theme="error">
            {error?.error || 'Unable to teardown component'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Text variant="base" weight="strong">
              Are you sure you want to teardown {component.name}?
            </Text>
            <Text variant="body" theme="neutral">
              Tearing down a component will remove all of its running deployments from the cloud account.
            </Text>
          </div>

          <div className="flex flex-col gap-3">
            <Text variant="body">
              This will create a workflow that attempts to:
            </Text>
            <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
              <li>Remove all infrastructure provisioned by this component</li>
              <li>Clean up all associated resources and dependencies</li>
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
              errorMessage={confirmName.length > 0 && !isConfirmValid ? "Component name doesn't match" : undefined}
            />
          </div>

          <Banner theme="warn">
            <Text variant="body">
              <strong>Important:</strong> This action cannot be undone. All infrastructure provisioned by this component will be permanently destroyed.
            </Text>
          </Banner>
        </div>
      </div>
    </Modal>
  )
}
