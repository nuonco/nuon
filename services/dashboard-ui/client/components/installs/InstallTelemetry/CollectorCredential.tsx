import { useForm, useStore } from '@tanstack/react-form'
import { useId } from 'react'
import { z } from 'zod'
import { Banner } from '@/components/common/Banner'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormSelect } from '@/components/common/form/FormSelect'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

const schema = z.object({
  duration: z.enum(['24h', '168h', '720h', '2160h', '8760h']),
})

export type TCollectorCredentialValues = z.infer<typeof schema>

export const CollectorCredentialModal = ({
  installId,
  isPending,
  error,
  credential,
  onSubmit,
  onDone,
  ...props
}: {
  installId: string
  isPending: boolean
  error: TAPIError | null
  credential: (TCollectorCredentialValues & { token: string }) | null
  onSubmit: (values: TCollectorCredentialValues) => void
  onDone: () => void
} & Omit<IModal, 'onSubmit'>) => {
  const id = useId()
  const form = useForm({
    defaultValues: { duration: '8760h' },
    validators: { onMount: schema, onChange: schema },
    onSubmit: ({ value }) => {
      if (!isPending) onSubmit(value as TCollectorCredentialValues)
    },
  })
  const canSubmit = useStore(form.store, (state) => state.canSubmit)

  if (credential) {
    return (
      <Modal
        heading="Bootstrap credential created"
        primaryActionTrigger={{
          children: 'Done',
          variant: 'primary',
          onClick: onDone,
        }}
        {...props}
      >
        <Banner theme="warn">
          Copy this token before closing the dialog. You cannot retrieve it
          again. Do not commit it or share it in logs or messages.
        </Banner>
        <Text>
          Save this token in the app's sensitive collector input for install{' '}
          {installId}, then deploy the collector through the app's normal
          workflow. The component creates its Kubernetes Secret.
        </Text>
        <div className="flex items-center gap-2 min-w-0">
          <Text family="mono" className="break-all min-w-0">
            {credential.token}
          </Text>
          <ClickToCopyButton
            textToCopy={credential.token}
            aria-label="Copy bootstrap token"
            className="shrink-0"
          />
        </div>
        <Text variant="subtext">
          Sensitive inputs are stored by Nuon and masked in the input UI. The
          deployed Secret and Helm release history also contain this token.
        </Text>
        <Text variant="subtext">
          Existing credentials remain valid. For rotation, bring the runner
          online, update the sensitive input, deploy the change, and verify the
          replacement works before revoking the old credential.
        </Text>
      </Modal>
    )
  }

  return (
    <Modal
      heading="Create bootstrap credential"
      primaryActionTrigger={{
        children: isPending ? 'Creating...' : 'Create credential',
        variant: 'primary',
        disabled: isPending || !canSubmit,
        onClick: () => form.handleSubmit(),
      }}
      {...props}
    >
      <Text>
        Create a token for this install's telemetry collector. Supply it through
        the app's sensitive install input; the component manages the Kubernetes
        Secret and namespace.
      </Text>
      <form
        autoComplete="off"
        noValidate
        onSubmit={(event) => event.preventDefault()}
        className="flex flex-col gap-4"
      >
        <FormErrorBanner error={error} fallback="Credential creation failed" />
        <form.Field name="duration">
          {(field) => (
            <FormSelect
              field={field}
              id={`${id}-duration`}
              labelProps={{ labelText: 'Expires after' }}
              helperText="Defaults to 1 year. Non-expiring credentials are not supported; rotate before expiry."
              disabled={isPending}
              options={[
                { value: '24h', label: '1 day' },
                { value: '168h', label: '1 week' },
                { value: '720h', label: '30 days' },
                { value: '2160h', label: '90 days' },
                { value: '8760h', label: '1 year' },
              ]}
            />
          )}
        </form.Field>
      </form>
      <Text variant="subtext" theme="neutral">
        Existing credentials are not invalidated. Creating access does not
        enable forwarding or deploy the collector.
      </Text>
      {error ? (
        <Text variant="subtext">
          Check issued credentials before retrying. If the response was lost, a
          credential may already have been issued; revoke it if its value is
          unavailable.
        </Text>
      ) : null}
    </Modal>
  )
}
