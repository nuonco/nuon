import { useForm, useStore } from '@tanstack/react-form'
import { z } from 'zod'
import type { TAPIError } from '@/types'
import type { TRoleInfo } from '@/types/ctl-api.types'
import { FormErrorBanner } from '../../molecules/FormErrorBanner'
import { FormInput } from '../../molecules/FormInput'
import { FormSelect } from '../../molecules/FormSelect'
import { Modal } from '../surfaces/Modal'

export const inviteMemberSchema = z.object({
  email: z
    .string()
    .trim()
    .min(1, 'Email is required')
    .email('Enter a valid email'),
  roleType: z.string().min(1, 'Role is required'),
})

export type TInviteMemberValues = z.infer<typeof inviteMemberSchema>

export const inviteMemberValuesValid = (values: TInviteMemberValues): boolean =>
  inviteMemberSchema.safeParse(values).success

export interface IInviteMember {
  roles: TRoleInfo[]
  onSubmit: (values: TInviteMemberValues) => void
  pending?: boolean
  rolesLoading?: boolean
  error?: TAPIError | Error | null
}

export const InviteMember = ({
  roles,
  onSubmit,
  pending = false,
  rolesLoading = false,
  error,
}: IInviteMember) => {
  const form = useForm({
    defaultValues: {
      email: '',
      roleType: '',
    },
    validators: {
      onMount: inviteMemberSchema,
      onChange: inviteMemberSchema,
    },
    onSubmit: ({ value }) => onSubmit(value),
  })
  const canSubmit = useStore(form.store, (state) => state.canSubmit)

  return (
    <Modal
      heading="Invite team member"
      primaryAction={{
        children: pending ? 'Inviting team member' : 'Invite team member',
        variant: 'primary',
        loading: pending,
        disabled: pending || !canSubmit,
        onClick: () => form.handleSubmit(),
      }}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(event) => event.preventDefault()}
        className="flex flex-col gap-4"
      >
        <FormErrorBanner
          error={error}
          fallback="Unable to invite team member"
        />
        <form.Field name="email">
          {(field) => (
            <FormInput
              field={field}
              label="Email"
              type="email"
              placeholder="member@example.com"
              autoComplete="off"
            />
          )}
        </form.Field>
        <form.Field name="roleType">
          {(field) => (
            <FormSelect
              field={field}
              label="Role"
              placeholder="Select role"
              loading={rolesLoading}
              options={roles.map((role) => ({
                value: role.role_type,
                label: role.title || role.role_type,
                description: role.description,
              }))}
            />
          )}
        </form.Field>
      </form>
    </Modal>
  )
}
