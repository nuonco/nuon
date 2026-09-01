import { useForm, useStore } from '@tanstack/react-form'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Code } from '@/components/common/Code'
import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import { branchCISettingsSchema, type BranchCISettingsValues } from './schema'

export const BranchCISettingsCard = ({
  ignoreChangesRegex,
  sendStatusesOnIgnore,
  onEdit,
}: {
  ignoreChangesRegex: string
  sendStatusesOnIgnore: boolean
  onEdit: () => void
}) => (
  <Card className="!p-4 !gap-4">
    <div className="flex items-center justify-between gap-3">
      <div className="flex flex-col gap-0.5">
        <Text weight="strong">CI triggers</Text>
        <Text variant="subtext" theme="neutral">
          Control which GitHub changes start branch runs.
        </Text>
      </div>
      <Button variant="secondary" onClick={onEdit}>
        Edit CI triggers
      </Button>
    </div>
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
      <LabeledValue label="Ignored changes regex">
        {ignoreChangesRegex ? (
          <Code>{ignoreChangesRegex}</Code>
        ) : (
          <Text variant="subtext" theme="neutral">
            Not configured
          </Text>
        )}
      </LabeledValue>
      <LabeledValue label="Status on ignored runs">
        <Text variant="subtext">
          {sendStatusesOnIgnore ? 'Send success status' : 'Do not send'}
        </Text>
      </LabeledValue>
    </div>
  </Card>
)

export const BranchCISettingsModal = ({
  ignoreChangesRegex,
  sendStatusesOnIgnore,
  isPending,
  error,
  onSubmit,
  onCancel,
  ...props
}: {
  ignoreChangesRegex: string
  sendStatusesOnIgnore: boolean
  isPending: boolean
  error: TAPIError | null
  onSubmit: (values: BranchCISettingsValues) => void
  onCancel: () => void
} & Omit<IModal, 'onSubmit'>) => {
  const form = useForm({
    defaultValues: {
      ignoreChangesRegex,
      sendStatusesOnIgnore,
    } as BranchCISettingsValues,
    validators: {
      onMount: branchCISettingsSchema,
      onChange: branchCISettingsSchema,
    },
    onSubmit: ({ value }) => onSubmit(value),
  })
  const values = useStore(form.store, (state) => state.values)
  const canSubmit = useStore(form.store, (state) => state.canSubmit)
  const hasChanges =
    values.ignoreChangesRegex !== ignoreChangesRegex ||
    values.sendStatusesOnIgnore !== sendStatusesOnIgnore

  return (
    <Modal
      heading="Edit CI triggers"
      primaryActionTrigger={{
        children: isPending ? 'Saving...' : 'Save changes',
        disabled: !canSubmit || !hasChanges || isPending,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      secondaryActionTrigger={{
        children: 'Cancel',
        disabled: isPending,
        onClick: onCancel,
      }}
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(event) => event.preventDefault()}
        className="flex flex-col gap-6"
      >
        <FormErrorBanner
          error={error}
          fallback="Unable to update CI triggers"
        />
        <form.Field name="ignoreChangesRegex">
          {(field) => (
            <div className="flex flex-col gap-2">
              <FormInput
                field={field}
                id="ignore-changes-regex"
                labelProps={{ labelText: 'Ignored changes regex (optional)' }}
                helperText="A run is not attempted when every changed file matches this RE2 regex."
                placeholder={'^(docs/|README\\.md)'}
                disabled={isPending}
              />
              <Button
                type="button"
                variant="secondary"
                size="sm"
                className="w-fit"
                onClick={() => field.handleChange('.*')}
                disabled={isPending || field.state.value === '.*'}
              >
                Ignore all changes
              </Button>
            </div>
          )}
        </form.Field>
        <form.Field name="sendStatusesOnIgnore">
          {(field) => (
            <FormCheckbox
              field={field}
              id="send-statuses-on-ignore"
              disabled={isPending}
              labelProps={{
                className: 'items-start gap-3',
                labelText: (
                  <span className="flex flex-col gap-1">
                    <Text weight="strong" className="!leading-none">
                      Send status for ignored runs
                    </Text>
                    <Text
                      variant="subtext"
                      theme="neutral"
                      className="!leading-none"
                    >
                      Post a success status so an ignored run does not block the
                      pull request.
                    </Text>
                  </span>
                ),
                labelTextProps: { as: 'div' },
              }}
            />
          )}
        </form.Field>
      </form>
    </Modal>
  )
}
