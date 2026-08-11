import { useForm, useStore } from '@tanstack/react-form'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Label } from '@/components/common/form/Label'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormSelect } from '@/components/common/form/FormSelect'
import { FormToggle } from '@/components/common/form/FormToggle'
import { hasSubCondition } from '@/components/oidc-trust-policies/CreateOIDCTrustPolicy'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TOIDCTrustPolicy } from '@/types'
import { editOIDCSchema, type EditOIDCFormValues } from './schema'

export type EditOIDCTrustPolicyFormInput = EditOIDCFormValues

const conditionsToRows = (
  claimConditions: TOIDCTrustPolicy['claim_conditions']
): EditOIDCFormValues['claimConditions'] => {
  const entries = Object.entries(claimConditions ?? {})
  return entries.length
    ? entries.map(([key, value]) => ({ key, value }))
    : [{ key: 'sub', value: '' }]
}

export const EditOIDCTrustPolicyModal = ({
  policy,
  isPending,
  error,
  roleOptions,
  onSubmit,
  ...props
}: {
  policy: TOIDCTrustPolicy
  isPending: boolean
  error: TAPIError | null
  roleOptions: { value: string; label: string; description?: string }[]
  onSubmit: (input: EditOIDCTrustPolicyFormInput) => void
} & Omit<IModal, 'onSubmit'>) => {
  const form = useForm({
    defaultValues: {
      name: policy.name ?? '',
      issuerUrl: policy.issuer_url ?? '',
      audience: policy.audience ?? '',
      role: policy.role ?? 'org_read_only',
      tokenDurationSeconds: policy.token_duration_seconds
        ? String(policy.token_duration_seconds)
        : '',
      enabled: policy.enabled ?? true,
      claimConditions: conditionsToRows(policy.claim_conditions),
    } as EditOIDCFormValues,
    validators: { onMount: editOIDCSchema, onChange: editOIDCSchema },
    onSubmit: ({ value }) =>
      onSubmit({
        name: value.name.trim(),
        issuerUrl: value.issuerUrl.trim(),
        audience: value.audience.trim(),
        role: value.role,
        tokenDurationSeconds: value.tokenDurationSeconds,
        enabled: value.enabled,
        claimConditions: value.claimConditions,
      }),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const claimConditions = useStore(form.store, (s) => s.values.claimConditions)
  const hasSub = hasSubCondition(claimConditions)

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="ShieldCheckIcon" size="24" />
          Edit trust policy
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Saving changes
          </span>
        ) : (
          'Save changes'
        ),
        disabled: !canSubmit || !hasSub || isPending,
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
        <FormErrorBanner error={error} fallback="Unable to update trust policy" />

        <form.Field name="enabled">
          {(field) => (
            <FormToggle
              field={field}
              label="Enabled"
              description="Disabled policies reject token exchange requests."
            />
          )}
        </form.Field>

        <form.Field name="name">
          {(field) => (
            <FormInput
              field={field}
              id="policy-name"
              disabled={isPending}
              labelProps={{ labelText: 'Name' }}
            />
          )}
        </form.Field>

        <form.Field name="issuerUrl">
          {(field) => (
            <FormInput
              field={field}
              id="policy-issuer-url"
              type="url"
              disabled={isPending}
              labelProps={{ labelText: 'Issuer URL' }}
              helperText="Must be an absolute http or https URL."
            />
          )}
        </form.Field>

        <form.Field name="audience">
          {(field) => (
            <FormInput
              field={field}
              id="policy-audience"
              disabled={isPending}
              labelProps={{ labelText: 'Audience' }}
              helperText="The expected `aud` claim value on the presented token."
            />
          )}
        </form.Field>

        <form.Field name="role">
          {(field) => (
            <FormSelect
              field={field}
              options={roleOptions}
              disabled={isPending}
              labelProps={{ labelText: 'Role' }}
              helperText="Org role granted to tokens exchanged with this policy."
            />
          )}
        </form.Field>

        <form.Field name="tokenDurationSeconds">
          {(field) => (
            <FormInput
              field={field}
              id="policy-token-duration"
              placeholder="3600"
              type="number"
              min={1}
              max={86400}
              disabled={isPending}
              labelProps={{ labelText: 'Token duration in seconds (optional)' }}
              helperText="Maximum is 86400."
            />
          )}
        </form.Field>

        <div className="flex flex-col gap-2">
          <Label>Claim conditions</Label>
          <Text variant="subtext" theme="neutral">
            All conditions must match the presented token. A `sub` condition is
            required.
          </Text>
          <form.Field name="claimConditions" mode="array">
            {(ccField) => (
              <>
                <div className="flex flex-col gap-2">
                  {ccField.state.value.map((_, index) => (
                    <div key={index} className="flex items-center gap-2">
                      <form.Field name={`claimConditions[${index}].key`}>
                        {(f) => (
                          <FormInput
                            field={f}
                            placeholder="sub"
                            disabled={isPending}
                          />
                        )}
                      </form.Field>
                      <form.Field name={`claimConditions[${index}].value`}>
                        {(f) => (
                          <FormInput
                            field={f}
                            placeholder="repo:acme/app:ref:refs/heads/main"
                            disabled={isPending}
                          />
                        )}
                      </form.Field>
                      <Button
                        variant="icon"
                        aria-label="Remove claim condition"
                        disabled={ccField.state.value.length === 1 || isPending}
                        onClick={() => ccField.removeValue(index)}
                      >
                        <Icon variant="TrashIcon" size={14} />
                      </Button>
                    </div>
                  ))}
                </div>
                <Button
                  variant="secondary"
                  size="sm"
                  className="w-fit"
                  disabled={isPending}
                  onClick={() => ccField.pushValue({ key: '', value: '' })}
                >
                  <Icon variant="PlusIcon" size={14} />
                  Add condition
                </Button>
              </>
            )}
          </form.Field>
        </div>
      </form>
    </Modal>
  )
}
