import { useForm, useStore } from '@tanstack/react-form'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TAPIError, TCloudConnection } from '@/types'
import { cn } from '@/utils/classnames'
import { accessSchema } from './schema'

export const AccessForm = ({
  onContinue,
  onBack,
  isPending = false,
  error,
}: {
  onContinue: (preset: TCloudConnection['preset']) => void
  onBack: () => void
  isPending?: boolean
  error?: TAPIError | null
}) => {
  const form = useForm({
    defaultValues: { preset: '' as TCloudConnection['preset'] | '' },
    validators: { onMount: accessSchema, onChange: accessSchema },
    onSubmit: ({ value }) => onContinue(accessSchema.parse(value).preset),
  })
  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  return (
    <form
      className="flex flex-col gap-6"
      onSubmit={(e) => {
        e.preventDefault()
        form.handleSubmit()
      }}
    >
      <SectionHeader
        title="Choose what Nuon can access"
        description="Choose managed stack access or provide the permissions policy in AWS. The OIDC trust stays limited to this connection."
      />
      <FormErrorBanner error={error} fallback="Connection creation failed" />
      <form.Field name="preset">
        {(field) => (
          <div className="grid gap-6 xl:grid-cols-2">
            <Card
              className={cn(
                field.state.value === 'stacks' && 'ring-2 ring-primary-500'
              )}
            >
              <Icon variant="StackIcon" size={28} theme="brand" />
              <div className="flex flex-wrap items-center gap-2">
                <Text variant="base" weight="strong">
                  Manage install stacks
                </Text>
                <Badge theme="info">Recommended</Badge>
              </div>
              <Text>
                <strong>What you grant:</strong> Nuon can create, update,
                delete, and inspect the install's CloudFormation stack,
                including its outputs.
              </Text>
              <Text variant="subtext" theme="neutral">
                <strong>Why Nuon needs it:</strong> Nuon keeps the install
                infrastructure in sync without requiring an operator to apply
                every stack change by hand.
              </Text>
              <Text variant="subtext" theme="neutral">
                Review the exact IAM actions and resource scope in the rendered
                permissions policy before running the setup.
              </Text>
              <Button
                type="button"
                variant="secondary"
                aria-pressed={field.state.value === 'stacks'}
                disabled={isPending}
                onClick={() => field.handleChange('stacks')}
              >
                {field.state.value === 'stacks'
                  ? 'Selected'
                  : 'Select stack access'}
              </Button>
            </Card>
            <Card
              className={cn(
                field.state.value === 'custom' && 'ring-2 ring-primary-500'
              )}
            >
              <Icon variant="SlidersHorizontalIcon" size={28} />
              <Text variant="base" weight="strong">
                Custom
              </Text>
              <Text>
                Nuon renders the trust policy so this connection can assume the
                role. You attach whatever permissions policy you want.
              </Text>
              <Button
                type="button"
                variant="secondary"
                aria-pressed={field.state.value === 'custom'}
                disabled={isPending}
                onClick={() => field.handleChange('custom')}
              >
                {field.state.value === 'custom'
                  ? 'Selected'
                  : 'Select custom access'}
              </Button>
            </Card>
          </div>
        )}
      </form.Field>
      <Text theme="neutral" variant="subtext">
        Continuing saves a Pending connection. If you leave before verification,
        you can resume setup or delete it from the connection page.
      </Text>
      <div className="flex justify-end gap-2 border-t pt-4">
        <Button
          type="button"
          variant="secondary"
          disabled={isPending}
          onClick={onBack}
        >
          Back
        </Button>
        <Button
          type="submit"
          variant="primary"
          disabled={!canSubmit || isPending}
          tooltipProps={
            !canSubmit
              ? {
                  tipContent: 'Cannot continue — choose stack or custom access',
                }
              : undefined
          }
        >
          {isPending ? 'Creating connection' : 'Continue'}
        </Button>
      </div>
    </form>
  )
}
