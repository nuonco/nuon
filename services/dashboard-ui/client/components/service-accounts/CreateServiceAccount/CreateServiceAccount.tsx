import { useForm, useStore } from '@tanstack/react-form'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Label } from '@/components/common/form/Label'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormSelect } from '@/components/common/form/FormSelect'
import { type SelectOption } from '@/components/common/form/Select'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import {
  createServiceAccountSchema,
  type CreateServiceAccountValues,
} from './schema'

export const CreateServiceAccountModal = ({
  roleOptions,
  rolesLoading,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  roleOptions: SelectOption[]
  rolesLoading?: boolean
  isPending: boolean
  error: TAPIError | null
  onSubmit: (params: CreateServiceAccountValues) => void
} & Omit<IModal, 'onSubmit'>) => {
  const defaultRole = roleOptions.some((o) => o.value === 'org_read_only')
    ? 'org_read_only'
    : (roleOptions[0]?.value ?? 'org_read_only')

  const form = useForm({
    defaultValues: {
      name: '',
      role: defaultRole,
    } as CreateServiceAccountValues,
    validators: {
      onMount: createServiceAccountSchema,
      onChange: createServiceAccountSchema,
    },
    onSubmit: ({ value }) => onSubmit(value),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="RobotIcon" size="24" />
          Create service account
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Creating service account
          </span>
        ) : (
          'Create service account'
        ),
        disabled: !canSubmit || isPending,
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
        <FormErrorBanner
          error={error}
          fallback="Unable to create service account"
        />

        <Text>
          Service accounts are non-human identities for automating access to the
          Nuon API.
        </Text>

        <form.Field name="name">
          {(field) => (
            <FormInput
              field={field}
              id="service-account-name"
              placeholder="e.g. ci-deploy"
              type="text"
              disabled={isPending}
              labelProps={{ labelText: 'Name' }}
            />
          )}
        </form.Field>

        {rolesLoading ? (
          <div className="flex flex-col gap-1">
            <Label htmlFor="service-account-role">
              <Text variant="body" className="font-medium">
                Role
              </Text>
            </Label>
            <Skeleton height="36px" />
          </div>
        ) : (
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
        )}
      </form>
    </Modal>
  )
}
