import { useForm, useStore } from '@tanstack/react-form'
import { z } from 'zod'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormToggle } from '@/components/common/form/FormToggle'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

const schema = z.object({ enabled: z.boolean() })

export const OrgTelemetryModal = ({
  orgName,
  enabled,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  orgName: string
  enabled: boolean
  isPending: boolean
  error: TAPIError | null
  onSubmit: (enabled: boolean) => void
} & Omit<IModal, 'onSubmit'>) => {
  const form = useForm({
    defaultValues: { enabled },
    validators: { onChange: schema },
    onSubmit: ({ value }) => {
      if (!isPending && value.enabled !== enabled) onSubmit(value.enabled)
    },
  })
  const canSubmit = useStore(form.store, (state) => state.canSubmit)
  const hasChanges = useStore(
    form.store,
    (state) => state.values.enabled !== enabled
  )

  return (
    <Modal
      heading="Manage telemetry"
      primaryActionTrigger={{
        children: isPending ? 'Saving...' : 'Save settings',
        variant: 'primary',
        disabled: isPending || !canSubmit || !hasChanges,
        tooltipProps:
          !hasChanges && !isPending
            ? { tipContent: 'No changes to save' }
            : undefined,
        onClick: () => form.handleSubmit(),
      }}
      {...props}
    >
      <form
        noValidate
        className="flex flex-col gap-6"
        onSubmit={(event) => {
          event.preventDefault()
          void form.handleSubmit()
        }}
      >
        <FormErrorBanner
          error={error}
          fallback="Unable to update telemetry settings"
        />
        <Text>
          Set the telemetry default for {orgName}. Existing and new installs
          without an override follow this setting.
        </Text>
        <form.Field name="enabled">
          {(field) => (
            <FormToggle
              field={field}
              label="Enable telemetry by default"
              description="Install-specific settings remain unchanged."
              disabled={isPending}
            />
          )}
        </form.Field>
      </form>
    </Modal>
  )
}
