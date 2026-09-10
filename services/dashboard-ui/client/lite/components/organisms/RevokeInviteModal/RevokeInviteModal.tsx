import { Banner } from '../../atoms/Banner'
import { Text } from '../../atoms/Text'
import { actionErrorMessage } from '../TeamTable/action-error'
import { Modal } from '../surfaces/Modal'

export interface IRevokeInviteModal {
  email: string
  onSubmit: () => void
  pending?: boolean
  error?: unknown
}

export const RevokeInviteModal = ({
  email,
  onSubmit,
  pending = false,
  error,
}: IRevokeInviteModal) => {
  const errorMessage = actionErrorMessage(error, 'Unable to revoke invite')

  return (
    <Modal
      heading="Revoke invite?"
      size="sm"
      primaryAction={{
        children: pending ? 'Revoking invite' : 'Revoke invite',
        variant: 'danger',
        loading: pending,
        disabled: pending,
        onClick: onSubmit,
      }}
    >
      {errorMessage ? (
        <Banner theme="error" heading="Invite revocation failed">
          {errorMessage}
        </Banner>
      ) : null}
      <Text color="secondary">
        Revoking the invitation for {email} will prevent it from being accepted.
        You can send a new invite later.
      </Text>
    </Modal>
  )
}
