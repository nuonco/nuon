import { useForm, useStore } from '@tanstack/react-form'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { FormInput } from '@/components/common/form/FormInput'
import { Label } from '@/components/common/form/Label'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import { SHARE_SCOPE_OPTIONS, type TInstallShare } from '../types'
import { SharedWithList } from '../SharedWithList/SharedWithList'
import { shareInstallSchema, type ShareInstallFormValues } from './schema'

export const ShareInstallModal = ({
  isPending,
  error,
  shares,
  installName,
  onSubmit,
  onRevoke,
  onResend,
  ...props
}: {
  isPending: boolean
  error: TAPIError | null
  shares?: TInstallShare[]
  installName?: string
  onSubmit: (params: ShareInstallFormValues) => void
  onRevoke?: (shareId: string) => void
  onResend?: (shareId: string) => void
} & Omit<IModal, 'onSubmit'>) => {
  const form = useForm({
    defaultValues: {
      email: '',
      approvals: false,
      readme: true,
      operations: false,
      resources: true,
    } as ShareInstallFormValues,
    validators: {
      onMount: shareInstallSchema,
      onChange: shareInstallSchema,
    },
    onSubmit: ({ value }) => onSubmit(value),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const formError = useStore(form.store, (s) => {
    const first = s.fieldMeta.approvals?.errors?.[0]
    if (first === undefined || first === null) return undefined
    if (typeof first === 'string') return first
    if (typeof first === 'object' && 'message' in first) {
      return String((first as { message: unknown }).message)
    }
    return String(first)
  })

  return (
    <Modal
      size="lg"
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="ShareIcon" size="24" />
          Share install
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Sharing...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="EnvelopeIcon" />
            Share install
          </span>
        ),
        disabled: !canSubmit || isPending,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to share this install'}
          </Banner>
        ) : null}

        <Text variant="body" theme="neutral">
          {installName
            ? `Invite someone to view ${installName} in the customer portal.`
            : 'Invite someone to view this install in the customer portal.'}
        </Text>

        <form
          className="flex flex-col gap-6"
          onSubmit={(e) => {
            e.preventDefault()
            form.handleSubmit()
          }}
        >
          <form.Field name="email">
            {(field) => (
              <FormInput
                field={field}
                id="share-install-email"
                type="email"
                placeholder="customer@example.com"
                labelProps={{
                  labelText: 'Email address',
                }}
              />
            )}
          </form.Field>

          <div className="flex flex-col gap-2">
            <Label>What they can access</Label>
            <Text variant="subtext" theme="neutral">
              Choose which parts of this install to share.
            </Text>
            <div className="flex flex-col gap-1">
              {SHARE_SCOPE_OPTIONS.map((option) => (
                <form.Field key={option.key} name={option.key}>
                  {(field) => (
                    <FormCheckbox
                      field={field}
                      labelProps={{
                        labelText: (
                          <span className="flex flex-col gap-0.5">
                            <Text variant="body">{option.label}</Text>
                            <Text variant="subtext" theme="neutral">
                              {option.description}
                            </Text>
                          </span>
                        ),
                      }}
                    />
                  )}
                </form.Field>
              ))}
            </div>
            {formError ? <Banner theme="error">{formError}</Banner> : null}
          </div>
        </form>

        {shares && shares.length > 0 ? (
          <div className="flex flex-col gap-3">
            <Text variant="base" weight="strong">
              Shared with
            </Text>
            <SharedWithList
              shares={shares}
              onRevoke={onRevoke}
              onResend={onResend}
            />
          </div>
        ) : null}
      </div>
    </Modal>
  )
}
