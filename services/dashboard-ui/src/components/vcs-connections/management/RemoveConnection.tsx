'use client'

import { useState } from 'react'
import { usePathname } from 'next/navigation'
import { removeVCSConnection } from '@/actions/vcs-connection/remove-vcs-connection'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import type { TVCSConnection } from '@/types'

interface IRemoveConnection {
  vcs_connection: TVCSConnection
}

export const RemoveConnectionModal = ({
  vcs_connection,
  ...props
}: IRemoveConnection & IModal) => {
  const path = usePathname()
  const { removeModal } = useSurfaces()
  const { org } = useOrg()

  // Connection name for validation
  const connectionName =
    vcs_connection?.github_account_name ||
    vcs_connection?.github_install_id ||
    ''

  // State for confirmation input
  const [confirmName, setConfirmName] = useState('')

  // Server action execution
  const { data, error, isLoading, execute } = useServerAction({
    action: removeVCSConnection,
  })

  // Toast notifications
  useServerActionToast({
    data,
    error,
    errorContent: (
      <Text>Unable to remove connection for {connectionName}.</Text>
    ),
    errorHeading: 'Removal failed',
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: (
      <Text>GitHub connection {connectionName} has been removed.</Text>
    ),
    successHeading: 'Connection removed',
  })

  // Validation
  const isConfirmValid = confirmName === connectionName
  const canRemove = isConfirmValid && !isLoading

  return (
    <Modal
      heading={
        <Text
          className="!inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="error"
        >
          <Icon variant="Trash" size="24" />
          Are you sure you want to disconnect this GitHub account?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Disconnecting...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="Trash" />
            Disconnect GitHub
          </span>
        ),
        onClick: () => {
          execute({
            connectionId: vcs_connection.id,
            orgId: org.id,
            path,
          })
        },
        disabled: !canRemove,
        variant: 'danger',
      }}
      size="half"
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to remove VCS connection'}
          </Banner>
        ) : null}

        {/* Connection name display */}
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-2">
            <Icon
              variant="GitHub"
              size="20"
              className="text-red-800 dark:text-red-400"
            />
            <Text
              variant="body"
              weight="strong"
              className="text-red-800 dark:text-red-400"
            >
              {connectionName}
            </Text>
          </div>
        </div>

        {/* Consequences list */}
        <div className="flex flex-col gap-3">
          <Text variant="body">This will:</Text>
          <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
            <li>
              Remove your Nuon organization&apos;s access to private repositories
            </li>
            <li>Potentially affect any workflows using private repos</li>
            <li>Allow you to reconnect this account at any time</li>
          </ul>
        </div>

        {/* Warning banner */}
        <Banner theme="warn">
          <Text variant="body">
            <strong>Warning:</strong> This action cannot be undone, but you can
            reconnect later.
          </Text>
        </Banner>

        {/* Verification input */}
        <div className="flex flex-col gap-2">
          <Text variant="body">
            To verify, type{' '}
            <span className="font-mono font-medium text-red-800 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-1 py-0.5 rounded">
              {connectionName}
            </span>{' '}
            below.
          </Text>
          <Input
            id="confirm-connection-name"
            placeholder="GitHub connection name"
            type="text"
            value={confirmName}
            onChange={(e) => setConfirmName(e.target.value)}
            error={confirmName.length > 0 && !isConfirmValid}
            errorMessage={
              confirmName.length > 0 && !isConfirmValid
                ? "Connection name doesn't match"
                : undefined
            }
          />
        </div>
      </div>
    </Modal>
  )
}

export const RemoveConnectionButton = ({
  vcs_connection,
  ...props
}: IRemoveConnection & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <RemoveConnectionModal vcs_connection={vcs_connection} />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
      className="!text-red-800 dark:!text-red-500 !p-2 w-full justify-between"
    >
      Remove connection
      <Icon variant="Trash" />
    </Button>
  )
}

