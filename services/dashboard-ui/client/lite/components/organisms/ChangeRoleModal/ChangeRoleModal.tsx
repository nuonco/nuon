import { useForm, useStore } from '@tanstack/react-form'
import type { TRoleInfo } from '@/types/ctl-api.types'
import { Banner } from '../../atoms/Banner'
import { Text } from '../../atoms/Text'
import { FormSelect } from '../../molecules/FormSelect'
import { actionErrorMessage } from '../TeamTable/action-error'
import { Modal } from '../surfaces/Modal'

export interface IChangeRoleModal {
  email: string
  roles: TRoleInfo[]
  currentRole?: string
  onSubmit: (roleType: string) => void
  pending?: boolean
  error?: unknown
}

export const ChangeRoleModal = ({
  email,
  roles,
  currentRole,
  onSubmit,
  pending = false,
  error,
}: IChangeRoleModal) => {
  const form = useForm({
    defaultValues: { roleType: currentRole ?? '' },
  })
  const selectedRole = useStore(form.store, (state) => state.values.roleType)
  const errorMessage = actionErrorMessage(error, 'Unable to change role')

  return (
    <Modal
      heading="Change role"
      size="sm"
      primaryAction={{
        children: pending ? 'Saving role' : 'Save role',
        variant: 'primary',
        loading: pending,
        disabled: pending || !selectedRole || selectedRole === currentRole,
        onClick: () => onSubmit(selectedRole),
      }}
    >
      {errorMessage ? (
        <Banner theme="error" heading="Role update failed">
          {errorMessage}
        </Banner>
      ) : null}
      <Text color="secondary">Change the role for {email}.</Text>
      <form.Field name="roleType">
        {(field) => (
          <FormSelect
            field={field}
            label="Role"
            options={roles.map((role) => ({
              value: role.role_type,
              label: role.title || role.role_type,
              description: role.description,
            }))}
          />
        )}
      </form.Field>
    </Modal>
  )
}
