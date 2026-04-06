import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { removeVCSConnection } from '@/lib/ctl-api/vcs-connections'
import type { TVCSConnection } from '@/types'

interface IRemoveConnection {
  vcs_connection: TVCSConnection
}

export const RemoveConnectionModal = ({
  vcs_connection,
  ...props
}: IRemoveConnection & IModal) => {
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const { org, refresh: refreshOrg } = useOrg()

  const connectionName =
    vcs_connection?.github_account_name ||
    vcs_connection?.github_install_id ||
    ''

  const [confirmName, setConfirmName] = useState('')

  const { mutate, isPending, error } = useMutation({
    mutationFn: removeVCSConnection,
    onSuccess: () => {
      refreshOrg()
      addToast(
        <Toast theme="info" heading="Connection removed">
          <Text>GitHub connection {connectionName} has been removed.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err) => {
      addToast(
        <Toast theme="error" heading="Removal failed">
          <Text>Unable to remove connection for {connectionName}.</Text>
        </Toast>
      )
    },
  })

  const isConfirmValid = confirmName === connectionName
  const canRemove = isConfirmValid && !isPending

  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
          theme="error"
        >
          <Icon variant="Trash" size="24" />
          Are you sure you want to disconnect this GitHub account?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
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
          mutate({ connectionId: vcs_connection.id, orgId: org.id })
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

        <Banner theme="warn">
          <Text variant="body">
            <strong>Warning:</strong> This action cannot be undone, but you can
            reconnect later.
          </Text>
        </Banner>

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
