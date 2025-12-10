'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useUser } from '@auth0/nextjs-auth0'
import { deprovisionSandbox } from '@/actions/installs/deprovision-sandbox'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Input } from '@/components/common/form/Input'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { trackEvent } from '@/lib/segment-analytics'

export const DeprovisionSandboxButton = ({
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()

  const modal = <DeprovisionSandboxModal />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
      className="!text-red-600 dark:!text-red-400"
    >
      {props?.isMenuButton ? null : <Icon variant="BoxArrowDown" />}
      Deprovision sandbox
      {props?.isMenuButton ? <Icon variant="BoxArrowDown" /> : null}
    </Button>
  )
}

export const DeprovisionSandboxModal = ({
  ...props
}: IModal) => {
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()

  const [confirm, setConfirm] = useState<string>('')

  const {
    data,
    error,
    execute,
    headers,
    isLoading,
  } = useServerAction({ action: deprovisionSandbox })

  useServerActionToast({
    data,
    error,
    errorContent: <Text>Failed to start sandbox deprovision. Please try again.</Text>,
    errorHeading: `Sandbox deprovision failed`,
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: <Text>Sandbox deprovision workflow has been started successfully.</Text>,
    successHeading: `Deprovision initiated`,
  })

  const handleDeprovision = () => {
    execute({
      body: {
        plan_only: false,
        error_behavior: 'abort',
      },
      installId: install.id,
      orgId: org.id,
    })
  }

  const handleClose = () => {
    removeModal(props.modalId)
  }

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'install_sandbox_deprovision',
        user,
        status: 'error',
        props: { orgId: org.id, installId: install.id, err: error?.error },
      })
    }

    if (data) {
      trackEvent({
        event: 'install_sandbox_deprovision',
        user,
        status: 'ok',
        props: { orgId: org.id, installId: install.id },
      })

      if (headers?.['x-nuon-install-workflow-id']) {
        router.push(
          `/${org.id}/installs/${install.id}/workflows/${headers['x-nuon-install-workflow-id']}`
        )
      } else {
        router.push(`/${org.id}/installs/${install.id}/workflows`)
      }
    }
  }, [data, error, headers])

  return (
    <Modal
      className="!max-w-xl"
      heading="Deprovision install sandbox"
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Deprovisioning
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="BoxArrowDown" />
            Deprovision sandbox
          </span>
        ),
        disabled: confirm !== 'deprovision' || isLoading,
        onClick: handleDeprovision,
        variant: 'danger' as const,
      }}
      onClose={handleClose}
      {...props}
    >
      <div className="flex flex-col gap-8 mb-12">
        {error?.error ? (
          <Banner theme="error">
            {error?.error || 'Unable to kickoff sandbox deprovision'}
          </Banner>
        ) : null}

        <span className="flex flex-col gap-1">
          <Text variant="h3" weight="strong">
            Are you sure you want to deprovision {install?.name} sandbox?
          </Text>
          <Text
            className="text-cool-grey-600 dark:text-white/70"
            variant="subtext"
          >
            Deprovisioning a sandbox will remove it from the cloud account.
          </Text>
        </span>

        <div className="flex flex-col gap-2">
          <Text variant="body">
            This will create a workflow that attempts to:
          </Text>

          <ul className="flex flex-col gap-1 list-disc pl-4">
            <li className="text-sm">Teardown the install sandbox</li>
          </ul>
        </div>

        <div className="w-full">
          <label className="flex flex-col gap-1 w-full">
            <Text variant="base" weight="strong">
              To verify, type{' '}
              <span className="text-red-800 dark:text-red-500">
                deprovision
              </span>{' '}
              below.
            </Text>
            <Input
              placeholder="deprovision"
              className="w-full"
              type="text"
              value={confirm}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                setConfirm(e?.currentTarget?.value)
              }}
            />
          </label>
        </div>

      </div>
    </Modal>
  )
}