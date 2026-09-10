import { Banner } from '../../atoms/Banner'
import { Text } from '../../atoms/Text'
import { actionErrorMessage } from '../TeamTable/action-error'
import { Modal } from '../surfaces/Modal'

export interface IResendInviteModal {
  email: string
  onSubmit: () => void
  pending?: boolean
  error?: unknown
}

export const ResendInviteModal = ({
  email,
  onSubmit,
  pending = false,
  error,
}: IResendInviteModal) => {
  const errorMessage = actionErrorMessage(error, 'Unable to resend invite')

  return (
    <Modal
      heading="Resend invite?"
      size="sm"
      primaryAction={{
        children: pending ? 'Sending invite' : 'Resend invite',
        variant: 'primary',
        loading: pending,
        disabled: pending,
        onClick: onSubmit,
      }}
    >
      {errorMessage ? (
        <Banner theme="error" heading="Invite resend failed">
          {errorMessage}
        </Banner>
      ) : null}
      <Text color="secondary">Resend the invitation email to {email}.</Text>
    </Modal>
  )
}
