'use client'

import { useState } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { forgetInstall } from '@/actions/installs/forget-install'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'

interface IForget {}

export const ForgetModal = ({ ...props }: IForget & IModal) => {
  const path = usePathname()
  const router = useRouter()
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()
  
  const [confirmName, setConfirmName] = useState('')

  const { data, error, isLoading, execute } = useServerAction({
    action: forgetInstall,
  })

  useServerActionToast({
    data,
    error,
    errorContent: (
      <Text>Unable to forget {install.name}.</Text>
    ),
    errorHeading: `Forget failed`,
    onSuccess: () => {
      // Navigate to installs list after successful forget
      router.push(`/${org.id}/installs`)
      removeModal(props.modalId)
    },
    successContent: (
      <Text>Install {install.name} has been forgotten.</Text>
    ),
    successHeading: `Install forgotten`,
  })

  const isConfirmValid = confirmName === install.name
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
          Forget {install.name}
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
            Forget install
          </span>
        ),
        onClick: () => {
          execute({
            orgId: org.id,
            path,
            installId: install.id,
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
            {error?.error || 'Unable to forget install.'}
          </Banner>
        ) : null}

        <Banner theme="warn">
          <Text variant="body">
            <strong>Warning:</strong> This should only be used in cases where an install was broken in an unordinary way and needs to be manually removed.
          </Text>
        </Banner>

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Text variant="base" weight="strong">
              Are you sure you want to forget {install.name}?
            </Text>
            <Text variant="body" theme="neutral">
              This action will remove the install and can not be undone.
            </Text>
          </div>

          <div className="flex flex-col gap-3">
            <Text variant="body">
              You should only do this after you have:
            </Text>
            <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
              <li>Successfully deprovisioned the install</li>
              <li>Deprovisioned the CloudFormation stack for this install</li>
            </ul>
          </div>

          <div className="flex flex-col gap-2">
            <Text variant="body">
              To verify, type{' '}
              <span className="font-mono font-medium text-red-800 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-1 py-0.5 rounded">
                {install.name}
              </span>{' '}
              below.
            </Text>
            <Input
              id="confirm-install-name"
              placeholder="install name"
              type="text"
              value={confirmName}
              onChange={(e) => setConfirmName(e.target.value)}
              error={confirmName.length > 0 && !isConfirmValid}
              errorMessage={confirmName.length > 0 && !isConfirmValid ? "Install name doesn't match" : undefined}
            />
          </div>
        </div>
      </div>
    </Modal>
  )
}

export const ForgetButton = ({
  ...props
}: IForget & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ForgetModal />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      variant="ghost"
      className="!text-red-800 dark:!text-red-500 !p-2 w-full justify-between"
      {...props}
    >
      Forget install
      <Icon variant="Trash" />
    </Button>
  )
}
