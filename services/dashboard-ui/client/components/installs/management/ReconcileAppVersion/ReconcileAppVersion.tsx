import { useForm } from '@tanstack/react-form'
import { DateTime } from 'luxon'
import { z } from 'zod'
import { Banner } from '@/components/common/Banner'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormRadioGroup } from '@/components/common/form/FormRadioGroup'
import { fieldErrorMessage } from '@/components/common/form/field-error'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TInstallScheduledReconcile } from '@/types'
import { PendingScheduledReconcile } from './PendingScheduledReconcile'
import { ScheduledReconcileHistory } from './ScheduledReconcileHistory'
import {
  parseLocalDateTimeInput,
  scheduleDisabledReason,
  validateScheduledAt,
} from './schedule-utils'

const defaultScheduledAtLocal = () =>
  DateTime.now().plus({ minutes: 15 }).toFormat("yyyy-MM-dd'T'HH:mm")

const reconcileAppVersionSchema = z
  .object({
    mode: z.enum(['now', 'schedule']),
    planOnly: z.boolean(),
    scheduledAtLocal: z.string(),
  })
  .superRefine((values, ctx) => {
    if (values.mode !== 'schedule') return
    if (!values.scheduledAtLocal) {
      ctx.addIssue({
        code: 'custom',
        path: ['scheduledAtLocal'],
        message: 'Choose a date and time.',
      })
      return
    }
    const message = validateScheduledAt(
      parseLocalDateTimeInput(values.scheduledAtLocal)
    )
    if (message) {
      ctx.addIssue({ code: 'custom', path: ['scheduledAtLocal'], message })
    }
  })

export type TReconcileAppVersionValues = z.infer<
  typeof reconcileAppVersionSchema
>

export type TReconcileAppVersionSubmitValues = {
  mode: 'now' | 'schedule'
  planOnly: boolean
  scheduledAt?: string
}

export interface IReconcileAppVersionModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  appliedConfigId?: string
  targetConfigId?: string
  unappliedCount: number
  isPending: boolean
  error?: TAPIError | null
  pending?: TInstallScheduledReconcile
  isCancelingPending?: boolean
  onCancelPending?: () => void
  history?: TInstallScheduledReconcile[]
  getWorkflowHref?: (workflowId: string) => string | undefined
  onSubmit: (values: TReconcileAppVersionSubmitValues) => void
}

export const ReconcileAppVersionModal = ({
  installName,
  appliedConfigId,
  targetConfigId,
  unappliedCount,
  isPending,
  error,
  pending,
  isCancelingPending,
  onCancelPending,
  history = [],
  getWorkflowHref,
  onSubmit,
  ...props
}: IReconcileAppVersionModal) => {
  const scheduleDisabled = scheduleDisabledReason(pending)

  const form = useForm({
    defaultValues: {
      mode: 'now',
      planOnly: false,
      scheduledAtLocal: defaultScheduledAtLocal(),
    } as TReconcileAppVersionValues,
    validators: { onSubmit: reconcileAppVersionSchema },
    onSubmit: ({ value }) => {
      if (isPending) return
      if (value.mode === 'schedule') {
        const utc = parseLocalDateTimeInput(value.scheduledAtLocal).toUTC()
        onSubmit({
          mode: 'schedule',
          planOnly: value.planOnly,
          scheduledAt: utc.toISO() ?? undefined,
        })
        return
      }
      onSubmit({ mode: 'now', planOnly: value.planOnly })
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
          <form.Subscribe selector={(state) => state.values.mode}>
            {(mode) => (mode === 'schedule' ? 'Schedule reconcile' : 'Run now')}
          </form.Subscribe>
        ),
        disabled: isPending,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        <Text variant="base">
          This deploys the differences between {installName}'s applied app
          config and its pinned app config, using the same approvals as a branch
          run deploy. A scheduled reconcile applies whatever app config version
          is pinned when it runs, not the version pinned now.
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
              {unappliedCount}{' '}
              {unappliedCount === 1 ? 'resource has' : 'resources have'} changes
              not yet applied.
            </Text>
          </Banner>
        ) : null}

        {pending ? (
          <Banner theme="info" data-testid="pending-schedule-banner">
            <PendingScheduledReconcile
              pending={pending}
              size="sm"
              isCanceling={isCancelingPending}
              onCancel={() => onCancelPending?.()}
            />
          </Banner>
        ) : null}

        <form.Field name="mode">
          {(field) => (
            <FormRadioGroup
              field={field}
              label="When"
              options={[
                { value: 'now', label: 'Run now' },
                {
                  value: 'schedule',
                  label: 'Schedule',
                  disabled: !!scheduleDisabled,
                },
              ]}
            />
          )}
        </form.Field>

        {scheduleDisabled ? (
          <Text variant="subtext" theme="neutral">
            {scheduleDisabled}
          </Text>
        ) : null}

        <form.Subscribe selector={(state) => state.values.mode}>
          {(mode) =>
            mode === 'schedule' ? (
              <form.Field name="scheduledAtLocal">
                {(field) => {
                  const parsed = field.state.value
                    ? parseLocalDateTimeInput(field.state.value)
                    : undefined
                  const errorMessage = fieldErrorMessage(field)
                  return (
                    <div className="flex flex-col gap-1">
                      <Label htmlFor="scheduled-at">
                        <Text variant="body" weight="strong">
                          Date and time
                        </Text>
                      </Label>
                      <Input
                        id="scheduled-at"
                        type="datetime-local"
                        value={field.state.value}
                        onChange={(e) => field.handleChange(e.target.value)}
                        onBlur={field.handleBlur}
                        error={!!errorMessage}
                        errorMessage={errorMessage}
                      />
                      <Text variant="subtext" theme="neutral">
                        {parsed?.isValid
                          ? `UTC: ${parsed.toUTC().toFormat('yyyy-MM-dd HH:mm')} (${parsed.zoneName})`
                          : 'Enter a date and time in your local timezone.'}
                      </Text>
                    </div>
                  )
                }}
              </form.Field>
            ) : null
          }
        </form.Subscribe>

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
                    <Text
                      as="p"
                      variant="subtext"
                      theme="neutral"
                      className="mt-1"
                    >
                      Generate a plan without applying any changes.
                    </Text>
                  </>
                ),
                labelTextProps: { as: 'div' },
              }}
            />
          )}
        </form.Field>

        <ScheduledReconcileHistory
          history={history}
          getWorkflowHref={getWorkflowHref}
        />

        <FormErrorBanner
          error={error}
          fallback="Unable to reconcile app version."
        />
      </div>
    </Modal>
  )
}
