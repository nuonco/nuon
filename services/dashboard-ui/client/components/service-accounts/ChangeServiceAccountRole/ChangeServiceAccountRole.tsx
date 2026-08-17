import { useForm, useStore } from '@tanstack/react-form'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormSelect } from '@/components/common/form/FormSelect'
import { type SelectOption } from '@/components/common/form/Select'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import {
  changeServiceAccountRoleSchema,
  type ChangeServiceAccountRoleValues,
} from './schema'

export const ChangeServiceAccountRoleModal = ({
  accountIdentity,
  currentRole,
  roleOptions,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  accountIdentity: string
  currentRole: string
  roleOptions: SelectOption[]
  isPending: boolean
  error: TAPIError | null
  onSubmit: (params: { role: string }) => void
} & Omit<IModal, 'onSubmit'>) => {
  const form = useForm({
    defaultValues: {
      role: currentRole || roleOptions[0]?.value || '',
    } as ChangeServiceAccountRoleValues,
    validators: {
      onMount: changeServiceAccountRoleSchema,
      onChange: changeServiceAccountRoleSchema,
    },
    onSubmit: ({ value }) => onSubmit({ role: value.role }),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const role = useStore(form.store, (s) => s.values.role)

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="UserCheckIcon" size="24" />
          Change role
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Saving
          </span>
        ) : (
          'Save'
        ),
        disabled: !canSubmit || isPending || role === currentRole,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(e) => e.preventDefault()}
        className="flex flex-col gap-6"
      >
        <FormErrorBanner error={error} fallback="Unable to change role" />

        <Text>
          Change the role for <strong>{accountIdentity}</strong>.
        </Text>

        <form.Field name="role">
          {(field) => (
            <FormSelect
              field={field}
              options={roleOptions}
              disabled={isPending}
              labelProps={{ labelText: 'Role' }}
            />
          )}
        </form.Field>
      </form>
    </Modal>
  )
}
