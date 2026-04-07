import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { removeVCSConnection } from '@/lib/ctl-api/vcs-connections'
import type { TVCSConnection } from '@/types'
import { RemoveConnectionModal } from './RemoveConnection'

interface IRemoveConnection {
  vcs_connection: TVCSConnection
}

export const RemoveConnectionModalContainer = ({
  vcs_connection,
  ...props
}: IRemoveConnection & Omit<IModal, 'onSubmit'>) => {
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const { org, refresh: refreshOrg } = useOrg()

  const connectionName =
    vcs_connection?.github_account_name ||
    vcs_connection?.github_install_id ||
    ''

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
    onError: () => {
      addToast(
        <Toast theme="error" heading="Removal failed">
          <Text>Unable to remove connection for {connectionName}.</Text>
        </Toast>
      )
    },
  })

  return (
    <RemoveConnectionModal
      connectionName={connectionName}
      isPending={isPending}
      error={error}
      onSubmit={() => mutate({ connectionId: vcs_connection.id, orgId: org.id })}
      {...props}
    />
  )
}

export const RemoveConnectionButton = ({
  vcs_connection,
  ...props
}: IRemoveConnection & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <RemoveConnectionModalContainer vcs_connection={vcs_connection} />

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
