import { Banner } from '../../atoms/Banner'
import { Input } from '../../atoms/Input'
import { Text } from '../../atoms/Text'
import { Field } from '../../molecules/Field'
import { actionErrorMessage } from '../TeamTable/action-error'
import { Modal } from '../surfaces/Modal'

export interface IRemoveMemberModal {
  email: string
  confirmation: string
  onConfirmationChange: (value: string) => void
  onSubmit: () => void
  pending?: boolean
  error?: unknown
}

export const RemoveMemberModal = ({
  email,
  confirmation,
  onConfirmationChange,
  onSubmit,
  pending = false,
  error,
}: IRemoveMemberModal) => {
  const confirmed = confirmation === email
  const errorMessage = actionErrorMessage(error, 'Unable to remove team member')

  return (
    <Modal
      heading="Remove team member?"
      size="default"
      primaryAction={{
        children: pending ? 'Removing team member' : 'Remove team member',
        variant: 'danger',
        loading: pending,
        disabled: pending || !confirmed,
        onClick: onSubmit,
      }}
    >
      {errorMessage ? (
        <Banner theme="error" heading="Team member removal failed">
          {errorMessage}
        </Banner>
      ) : null}
      <Text color="secondary">
        Removing {email} will revoke their access to your org.
      </Text>
      <Text>
        <strong>Warning:</strong> Their access will be revoked immediately.
      </Text>
      <Text>
        To verify, type <strong>{email}</strong> below.
      </Text>
      <Field
        error={confirmation && !confirmed ? "Email doesn't match" : undefined}
      >
        <Input
          value={confirmation}
          onChange={(event) => onConfirmationChange(event.target.value)}
          placeholder={email}
          aria-label="Confirm member email"
          autoComplete="off"
        />
      </Field>
    </Modal>
  )
}
