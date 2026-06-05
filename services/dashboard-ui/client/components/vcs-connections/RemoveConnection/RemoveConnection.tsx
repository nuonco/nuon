import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

interface IRemoveConnectionModal extends Omit<IModal, 'onSubmit'> {
  connectionName: string
  isPending: boolean
  error?: TAPIError | null
  deleteGithubApp: boolean
  onDeleteGithubAppChange: (val: boolean) => void
  onSubmit: () => void
}

export const RemoveConnectionModal = ({
  connectionName,
  isPending,
  error,
  deleteGithubApp,
  onDeleteGithubAppChange,
  onSubmit,
  ...props
}: IRemoveConnectionModal) => {
  const [confirmName, setConfirmName] = useState('')

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
          <Icon variant="TrashIcon" size="24" />
          Disconnect GitHub account?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Disconnecting...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="TrashIcon" />
            Disconnect GitHub
          </span>
        ),
        onClick: onSubmit,
        disabled: !canRemove,
        variant: 'danger',
      }}
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

        <div className="flex flex-col gap-1">
          <CheckboxInput
            checked={deleteGithubApp}
            onChange={(e) => onDeleteGithubAppChange(e.target.checked)}
            labelProps={{
              labelText: 'Also uninstall the Nuon GitHub App from GitHub',
            }}
          />
        </div>

        {deleteGithubApp ? (
          <Banner theme="warn">
            <Text variant="body">
              <strong>Warning:</strong> The GitHub App will be uninstalled from
              your GitHub account or organization. Any other Nuon org sharing
              this installation will lose access.
            </Text>
          </Banner>
        ) : (
          <Banner theme="warn">
            <Text variant="body">
              <strong>Note:</strong> The Nuon GitHub App will remain installed
              on your GitHub account or organization. To fully revoke access,
              remove it manually from your GitHub settings after disconnecting.
            </Text>
          </Banner>
        )}

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
