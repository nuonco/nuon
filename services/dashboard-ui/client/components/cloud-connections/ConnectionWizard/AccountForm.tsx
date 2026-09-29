import { useForm, useStore } from '@tanstack/react-form'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Tooltip } from '@/components/common/Tooltip'
import { FormInput } from '@/components/common/form/FormInput'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { accountSchema, type TAccountValues } from './schema'

export const AccountForm = ({
  values,
  onContinue,
}: {
  values?: TAccountValues
  onContinue: (values: TAccountValues) => void
}) => {
  const form = useForm({
    defaultValues: values || {
      name: '',
      target_id: '',
      role_name: 'nuon-cloud-connection',
    },
    validators: { onMount: accountSchema, onChange: accountSchema },
    onSubmit: ({ value }) => onContinue(accountSchema.parse(value)),
  })
  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  return (
    <form
      className="flex max-w-2xl flex-col gap-6"
      onSubmit={(e) => {
        e.preventDefault()
        form.handleSubmit()
      }}
    >
      <SectionHeader
        title="Choose the AWS account"
        description="Nuon assumes an IAM role in this account via OIDC. You decide what that role can do."
      />
      <form.Field name="name">
        {(field) => (
          <FormInput
            field={field}
            id="connection-name"
            labelProps={{ labelText: 'Connection name' }}
            helperText="Use a name that identifies the account and its purpose."
            placeholder="acme-production"
          />
        )}
      </form.Field>
      <form.Field name="target_id">
        {(field) => (
          <FormInput
            field={field}
            id="connection-account"
            inputMode="numeric"
            placeholder="123456789012"
            labelProps={{
              labelText: (
                <span className="flex items-center gap-1.5">
                  AWS account ID{' '}
                  <Tooltip tipContent="The AWS account ID you want Nuon to give access to">
                    <span tabIndex={0} aria-label="AWS account ID information">
                      <Icon variant="InfoIcon" size={14} />
                    </span>
                  </Tooltip>
                </span>
              ),
            }}
          />
        )}
      </form.Field>
      <form.Field name="role_name">
        {(field) => (
          <FormInput
            field={field}
            id="connection-role"
            labelProps={{ labelText: 'Role name' }}
            helperText="The IAM role to create in this AWS account."
          />
        )}
      </form.Field>
      <div className="flex justify-end border-t pt-4">
        <Button type="submit" variant="primary" disabled={!canSubmit}>
          Continue
        </Button>
      </div>
    </form>
  )
}
