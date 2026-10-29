import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import { Banner } from '@/components/common/Banner'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

const reconcileAppVersionSchema = z.object({ planOnly: z.boolean() })

export type TReconcileAppVersionValues = z.infer<typeof reconcileAppVersionSchema>

export interface IReconcileAppVersionModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  appliedConfigId?: string
  targetConfigId?: string
  unappliedCount: number
  isPending: boolean
  error?: TAPIError | null
  onSubmit: (values: TReconcileAppVersionValues) => void
}

export const ReconcileAppVersionModal = ({
  installName,
  appliedConfigId,
  targetConfigId,
  unappliedCount,
  isPending,
  error,
  onSubmit,
  ...props
}: IReconcileAppVersionModal) => {
  const form = useForm({
    defaultValues: { planOnly: false } as TReconcileAppVersionValues,
    validators: { onSubmit: reconcileAppVersionSchema },
    onSubmit: ({ value }) => {
      if (!isPending) onSubmit(value)
    },
  })

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="ArrowsClockwiseIcon" size="24" />
          Reconcile app version
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Starting reconcile
          </span>
        ) : (
          'Run now'
        ),
        disabled: isPending,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        <Text variant="base">
          This deploys the differences between {installName}'s applied app config
          and its pinned app config, using the same approvals as a branch run
          deploy.
        </Text>

        <div className="flex items-end gap-6">
          <LabeledValue label="Applied config" className="flex-1">
            {appliedConfigId ? <ID>{appliedConfigId}</ID> : 'Never provisioned'}
          </LabeledValue>
          <LabeledValue label="Pinned config" className="flex-1">
            {targetConfigId ? <ID>{targetConfigId}</ID> : '—'}
          </LabeledValue>
        </div>

        {unappliedCount > 0 ? (
          <Banner theme="info">
            <Text variant="body">
              {unappliedCount} {unappliedCount === 1 ? 'resource has' : 'resources have'}{' '}
              changes not yet applied.
            </Text>
          </Banner>
        ) : null}

        <form.Field name="planOnly">
          {(field) => (
            <FormCheckbox
              field={field}
              disabled={isPending}
              className="mt-1"
              labelProps={{
                className: 'items-start',
                labelText: (
                  <>
                    <Text as="p" weight="strong">
                      Plan only
                    </Text>
                    <Text as="p" variant="subtext" theme="neutral" className="mt-1">
                      Generate a plan without applying any changes.
                    </Text>
                  </>
                ),
                labelTextProps: { as: 'div' },
              }}
            />
          )}
        </form.Field>

        <FormErrorBanner error={error} fallback="Unable to reconcile app version." />
      </div>
    </Modal>
  )
}
