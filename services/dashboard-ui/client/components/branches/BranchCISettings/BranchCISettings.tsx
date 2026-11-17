import { useForm, useStore } from '@tanstack/react-form'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Code } from '@/components/common/Code'
import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormRadioGroup } from '@/components/common/form/FormRadioGroup'
import { FormTextarea } from '@/components/common/form/FormTextarea'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import {
  branchCISettingsSchema,
  parsePathPatterns,
  type BranchCISettingsValues,
} from './schema'

type TPathFilter = Pick<
  BranchCISettingsValues,
  'pathFilterMode' | 'pathPatterns'
>

const pathFilterLabel = (mode: BranchCISettingsValues['pathFilterMode']) =>
  mode === 'paths_ignore' ? 'Ignored paths' : 'Run paths'

export const BranchCISettingsCard = ({
  ignoreChangesRegex,
  sendStatusesOnIgnore,
  pathFilterMode,
  pathPatterns,
  onEdit,
}: {
  ignoreChangesRegex: string
  sendStatusesOnIgnore: boolean
  onEdit: () => void
} & TPathFilter) => {
  const patterns = parsePathPatterns(pathPatterns)

  return (
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
        <LabeledValue label={pathFilterLabel(pathFilterMode)}>
          {patterns.length ? (
            <div className="flex flex-wrap gap-1">
              {patterns.map((pattern) => (
                <Code key={pattern}>{pattern}</Code>
              ))}
            </div>
          ) : (
            <Text variant="subtext" theme="neutral">
              Repo directory
            </Text>
          )}
        </LabeledValue>
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
}

export const BranchCISettingsModal = ({
  ignoreChangesRegex,
  sendStatusesOnIgnore,
  pathFilterMode,
  pathPatterns,
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
} & TPathFilter &
  Omit<IModal, 'onSubmit'>) => {
  const form = useForm({
    defaultValues: {
      ignoreChangesRegex,
      sendStatusesOnIgnore,
      pathFilterMode,
      pathPatterns,
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
    values.sendStatusesOnIgnore !== sendStatusesOnIgnore ||
    values.pathFilterMode !== pathFilterMode ||
    parsePathPatterns(values.pathPatterns).join('\n') !==
      parsePathPatterns(pathPatterns).join('\n')

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
        <form.Field name="pathFilterMode">
          {(field) => (
            <FormRadioGroup
              field={field}
              label="Path filter"
              disabled={isPending}
              options={[
                {
                  value: 'paths',
                  label: 'Run paths',
                  description:
                    'Push and pull request runs start only when a changed file matches. Prefix a pattern with ! to exclude it.',
                },
                {
                  value: 'paths_ignore',
                  label: 'Ignored paths',
                  description:
                    'Push and pull request runs are skipped when every changed file matches.',
                },
              ]}
            />
          )}
        </form.Field>
        <form.Field name="pathPatterns">
          {(field) => (
            <FormTextarea
              field={field}
              id="path-patterns"
              labelProps={{ labelText: 'Path patterns (optional)' }}
              helperText="One glob per line, relative to the repository root. Leave empty to use the repo directory."
              placeholder={'apps/acme/**\nshared/helm/**\n!apps/acme/docs/**'}
              disabled={isPending}
              autoResize
            />
          )}
        </form.Field>
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
