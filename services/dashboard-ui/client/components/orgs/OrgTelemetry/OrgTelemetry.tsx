import { useForm, useStore } from '@tanstack/react-form'
import { z } from 'zod'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormToggle } from '@/components/common/form/FormToggle'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TOrgTelemetryUpdate } from '@/lib/ctl-api/orgs/update-org-telemetry'
import type { TAPIError } from '@/types'

const schema = z.object({
  enabled: z.boolean(),
  relayEndpoint: z
    .string()
    .max(4096)
    .refine((value) => {
      if (!value) return true
      try {
        const url = new URL(value)
        return (
          value.startsWith('https://') &&
          !!url.hostname &&
          !value.split('/')[2]?.includes('@') &&
          !/[?#${}\\\s]/.test(value)
        )
      } catch {
        return false
      }
    }, 'Enter an HTTPS URL without credentials, query parameters, or fragments'),
})

export const OrgTelemetryModal = ({
  orgName,
  enabled,
  relayEndpoint,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  orgName: string
  enabled: boolean
  relayEndpoint?: string | null
  isPending: boolean
  error: TAPIError | null
  onSubmit: (settings: TOrgTelemetryUpdate) => void
} & Omit<IModal, 'onSubmit'>) => {
  const form = useForm({
    defaultValues: { enabled, relayEndpoint: relayEndpoint ?? '' },
    validators: { onChange: schema },
    onSubmit: ({ value }) => {
      if (isPending) return
      const settings: TOrgTelemetryUpdate = {}
      if (value.enabled !== enabled) settings.enabled = value.enabled
      if (value.relayEndpoint !== (relayEndpoint ?? '')) {
        settings.relay_endpoint = value.relayEndpoint || null
      }
      if (Object.keys(settings).length) onSubmit(settings)
    },
  })
  const canSubmit = useStore(form.store, (state) => state.canSubmit)
  const hasChanges = useStore(
    form.store,
    (state) =>
      state.values.enabled !== enabled ||
      state.values.relayEndpoint !== (relayEndpoint ?? '')
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
        <Text>Configure the telemetry relay and default for {orgName}.</Text>
        <form.Field name="relayEndpoint">
          {(field) => (
            <FormInput
              field={field}
              id="telemetry-relay-endpoint"
              type="url"
              labelProps={{ labelText: 'Telemetry relay URL (optional)' }}
              placeholder="https://relay.example.com/telemetry"
              helperText="Relay used by telemetry-enabled installs in this org. Leave blank to use the deployment default, if configured."
              disabled={isPending}
            />
          )}
        </form.Field>
        <form.Field name="enabled">
          {(field) => (
            <FormToggle
              field={field}
              label="Enable telemetry by default"
              description="Existing and new installs without an override follow this setting."
              disabled={isPending}
            />
          )}
        </form.Field>
      </form>
    </Modal>
  )
}
