import { useForm, useStore } from '@tanstack/react-form'
import { useId, type ReactNode } from 'react'
import { z } from 'zod'
import { Banner } from '@/components/common/Banner'
import { Code } from '@/components/common/Code'
import { Link } from '@/components/common/Link'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormSelect } from '@/components/common/form/FormSelect'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TInstallTelemetrySettings } from '@/types'

const schema = z.object({ forwarding: z.enum(['default', 'on', 'off']) })

export const InstallTelemetryModal = ({
  installName,
  settings,
  isManagedByConfig,
  isPending,
  error,
  onSubmit,
  children,
  ...props
}: {
  installName: string
  settings: TInstallTelemetrySettings
  isManagedByConfig: boolean
  isPending: boolean
  error: TAPIError | null
  onSubmit: (enabled: boolean | null) => void
  children?: ReactNode
} & Omit<IModal, 'onSubmit'>) => {
  const id = useId()
  const initial =
    settings.override == null ? 'default' : settings.override ? 'on' : 'off'
  const form = useForm({
    defaultValues: { forwarding: initial },
    validators: { onMount: schema, onChange: schema },
    onSubmit: ({ value }) => {
      if (isPending || isManagedByConfig) return
      onSubmit(
        value.forwarding === 'default' ? null : value.forwarding === 'on'
      )
    },
  })
  const forwarding = useStore(form.store, (state) => state.values.forwarding)
  const canSubmit = useStore(form.store, (state) => state.canSubmit)

  return (
    <Modal
      heading="Manage telemetry"
      size="lg"
      primaryActionTrigger={
        isManagedByConfig
          ? undefined
          : {
              children: isPending ? 'Saving...' : 'Save settings',
              variant: 'primary',
              disabled: isPending || !canSubmit || forwarding === initial,
              tooltipProps:
                forwarding === initial
                  ? { tipContent: 'No changes to save' }
                  : undefined,
              onClick: () => form.handleSubmit(),
            }
      }
      {...props}
    >
      <Text>Manage application telemetry for {installName}.</Text>
      <form
        autoComplete="off"
        noValidate
        onSubmit={(event) => event.preventDefault()}
        className="flex flex-col gap-4"
      >
        <FormErrorBanner error={error} fallback="Telemetry update failed" />
        <form.Field name="forwarding">
          {(field) => (
            <FormSelect
              field={field}
              id={`${id}-forwarding`}
              labelProps={{ labelText: 'Telemetry forwarding' }}
              helperText="Allow this install to send application telemetry to your org's configured backend. This does not deploy a collector."
              options={[
                {
                  value: 'default',
                  label: `Use org default (${settings.org_default ? 'On' : 'Off'})`,
                },
                { value: 'on', label: 'On' },
                { value: 'off', label: 'Off' },
              ]}
              disabled={isPending || isManagedByConfig}
            />
          )}
        </form.Field>
        {isManagedByConfig ? (
          <>
            <Text variant="subtext">
              Managed by install config. Change and sync the install config to
              update forwarding.
            </Text>
            <Code variant="preformated">
              {'[telemetry]\nenabled = true # or false'}
            </Code>
            <Text variant="subtext">
              Removing this section preserves the existing override; it does not
              restore the org default.
            </Text>
          </>
        ) : null}
        {!settings.relay_configured ? (
          <Banner theme="warn">
            Relay setup is required before collectors can forward telemetry.
            Saving a policy does not configure the relay.
          </Banner>
        ) : null}
        <LabeledValue
          label={
            <Text variant="subtext" theme="neutral">
              <Tooltip
                showIcon
                tabIndex={0}
                aria-describedby={`${id}-relay-help`}
                className="cursor-help rounded-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-400/80"
                tipContentClassName="max-w-72 !whitespace-normal"
                tipContent={
                  <Text variant="subtext" id={`${id}-relay-help`}>
                    Nuon's shared relay forwards application telemetry from
                    either collector to your org's configured backend.
                  </Text>
                }
              >
                Relay configuration
              </Tooltip>
            </Text>
          }
        >
          <Status status={settings.relay_configured ? 'success' : 'warning'}>
            <span>
              {settings.relay_configured
                ? 'Configured'
                : 'Setup required — contact your Nuon administrator'}
            </span>
          </Status>
        </LabeledValue>
        <Text variant="subtext" theme="neutral">
          Applies to both collectors. Turning forwarding off keeps collector
          access and infrastructure in place.
        </Text>
      </form>
      {children}
      <Link
        href="https://docs.nuon.co/guides/manage-install-telemetry"
        isExternal
      >
        View telemetry guide
      </Link>
    </Modal>
  )
}
