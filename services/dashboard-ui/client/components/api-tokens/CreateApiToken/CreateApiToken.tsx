import { useForm, useStore } from '@tanstack/react-form'
import { Banner } from '@/components/common/Banner'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Label } from '@/components/common/form/Label'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormSelect } from '@/components/common/form/FormSelect'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import { createApiTokenSchema, type CreateApiTokenValues } from './schema'

export const DURATION_OPTIONS = [
  { value: '24h', label: '1 day' },
  { value: '168h', label: '1 week' },
  { value: '720h', label: '30 days' },
  { value: '2160h', label: '90 days' },
  { value: '8760h', label: '1 year' },
]

export const CreateApiTokenModal = ({
  isPending,
  error,
  createdToken,
  roleOptions,
  rolesLoading,
  onSubmit,
  onDone,
  ...props
}: {
  isPending: boolean
  error: TAPIError | null
  createdToken: string | null
  roleOptions: { value: string; label: string; description?: string }[]
  rolesLoading?: boolean
  onSubmit: (params: CreateApiTokenValues) => void
  onDone: () => void
} & Omit<IModal, 'onSubmit'>) => {
  const defaultRole = roleOptions.some((o) => o.value === 'org_read_only')
    ? 'org_read_only'
    : (roleOptions[0]?.value ?? 'org_read_only')

  const form = useForm({
    defaultValues: {
      name: '',
      role: defaultRole,
      duration: '720h',
    } as CreateApiTokenValues,
    validators: {
      onMount: createApiTokenSchema,
      onChange: createApiTokenSchema,
    },
    onSubmit: ({ value }) => onSubmit(value),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)

  if (createdToken) {
    return (
      <Modal
        heading={
          <Text flex className="gap-4" variant="h3" weight="strong">
            <Icon variant="KeyIcon" size="24" />
            API token created
          </Text>
        }
        primaryActionTrigger={{
          children: 'Done',
          onClick: () => onDone(),
          variant: 'primary',
        }}
        {...props}
      >
        <div className="flex flex-col gap-4">
          <Banner theme="warn">
            Save this token somewhere safe. You won't be able to view it again.
          </Banner>
          <div className="flex items-center gap-2">
            <Text variant="body" family="mono" className="break-all">
              {createdToken}
            </Text>
            <ClickToCopyButton textToCopy={createdToken} />
          </div>
        </div>
      </Modal>
    )
  }

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="KeyIcon" size="24" />
          Create API token
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Creating token
          </span>
        ) : (
          'Create token'
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
        <FormErrorBanner error={error} fallback="Unable to create API token" />

        <form.Field name="name">
          {(field) => (
            <FormInput
              field={field}
              id="token-name"
              placeholder="e.g. ci-deploy"
              type="text"
              disabled={isPending}
              labelProps={{ labelText: 'Name' }}
            />
          )}
        </form.Field>

        {rolesLoading ? (
          <div className="flex flex-col gap-1">
            <Label htmlFor="token-role">
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

        <form.Field name="duration">
          {(field) => (
            <FormSelect
              field={field}
              options={DURATION_OPTIONS}
              disabled={isPending}
              labelProps={{ labelText: 'Expires after' }}
            />
          )}
        </form.Field>
      </form>
    </Modal>
  )
}
