import { useMemo, useState, type ReactNode } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import type { FormValidateOrFn } from '@tanstack/form-core'
import { Badge } from '@/components/common/Badge'
import { type IButtonAsButton } from '@/components/common/Button'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { FormCodeInput } from '@/components/common/form/FormCodeInput'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { WizardNavComponent } from '@/components/onboarding/WizardNav'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import type { TRunbookInput } from '@/lib/ctl-api/apps/runbooks'
import type {
  TInstallRunbook,
  TRunRunbookBody,
} from '@/lib/ctl-api/installs/runbooks'
import {
  buildRunbookSchema,
  isBooleanInput,
  type RunbookFormValues,
} from './schema'

interface IRunRunbookForm extends Omit<IModal, 'onSubmit'> {
  installRunbook: TInstallRunbook
  isPending: boolean
  error: TAPIError | null
  onSubmit: (body: TRunRunbookBody) => void
  roleSelector: ReactNode
}

type RunbookFormApi = ReturnType<typeof useForm<RunbookFormValues>>

const RunbookInputField = ({
  form,
  input,
}: {
  form: RunbookFormApi
  input: TRunbookInput
}) => {
  const name = `inputs.${input.name}`
  const label = input.display_name || input.name

  if (isBooleanInput(input)) {
    return (
      <form.Field name={name}>
        {(field) => (
          <div className="flex flex-col gap-1">
            <FormCheckbox field={field} labelProps={{ labelText: label }} />
            {input.description ? (
              <Text variant="subtext">{input.description}</Text>
            ) : null}
          </div>
        )}
      </form.Field>
    )
  }

  const labelText = `${label}${input.required ? ' *' : ' (optional)'}`

  if (input.type === 'json') {
    return (
      <form.Field name={name}>
        {(field) => (
          <FormCodeInput
            field={field}
            language="json"
            labelProps={{ labelText }}
            helperText={input.description}
          />
        )}
      </form.Field>
    )
  }

  return (
    <form.Field name={name}>
      {(field) => (
        <FormInput
          field={field}
          id={`runbook-input-${input.name}`}
          type={
            input.sensitive
              ? 'password'
              : input.type === 'number'
                ? 'number'
                : 'text'
          }
          labelProps={{ labelText }}
          helperText={input.description}
        />
      )}
    </form.Field>
  )
}

export const RunRunbookForm = ({
  installRunbook,
  isPending,
  error,
  onSubmit,
  roleSelector,
  ...props
}: IRunRunbookForm) => {
  const runbookName = installRunbook.runbook?.name ?? 'runbook'
  const config = installRunbook.runbook?.configs?.[0]

  const steps = useMemo(
    () => (config?.steps ?? []).slice().sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0)),
    [config]
  )
  const inputs = useMemo(
    () => (config?.inputs ?? []).slice().sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0)),
    [config]
  )
  const hasInputs = inputs.length > 0

  const schema = useMemo(() => buildRunbookSchema(inputs), [inputs])
  const validator = schema as unknown as FormValidateOrFn<RunbookFormValues>
  const defaultValues = useMemo<RunbookFormValues>(
    () => ({
      inputs: Object.fromEntries(
        inputs.map((input) => [
          input.name,
          isBooleanInput(input) ? input.default === 'true' : (input.default ?? ''),
        ])
      ),
    }),
    [inputs]
  )

  const [page, setPage] = useState<0 | 1 | 2>(0)
  const [stepEnabled, setStepEnabled] = useState<Record<string, boolean>>(() =>
    Object.fromEntries(steps.map((s) => [s.id ?? '', true]))
  )
  const isStepEnabled = (id?: string) => stepEnabled[id ?? ''] ?? true
  const enabledCount = steps.filter((s) => isStepEnabled(s.id)).length
  const noStepsEnabled = enabledCount === 0

  const form = useForm({
    defaultValues,
    validators: { onMount: validator, onChange: validator },
    onSubmit: ({ value }) => {
      const inputsMap = Object.fromEntries(
        Object.entries(value.inputs).map(([key, v]) => [
          key,
          typeof v === 'boolean' ? String(v) : v,
        ])
      )
      onSubmit({
        ...(hasInputs ? { inputs: inputsMap } : {}),
        steps: steps.map((s) => ({
          step_id: s.id ?? '',
          enabled: isStepEnabled(s.id),
        })),
      })
    },
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const values = useStore(form.store, (s) => s.values)

  const showInputsForm = hasInputs && page === 0
  const showStepsSummary = !hasInputs || page === 1
  const showInputsSummary = hasInputs && page === 2
  const isSubmitView = !hasInputs || page === 2

  const primaryActionTrigger: IButtonAsButton = isSubmitView
    ? {
        children: isPending ? (
          <>
            <Icon variant="Loading" className="animate-spin" />
            Running...
          </>
        ) : (
          <>
            Run runbook
            <Icon variant="PlayIcon" />
          </>
        ),
        disabled: isPending || noStepsEnabled || !canSubmit,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }
    : {
        children: 'Next',
        onClick: showInputsForm ? () => setPage(1) : () => setPage(2),
        disabled: showInputsForm ? !canSubmit : noStepsEnabled,
        variant: 'primary',
      }

  const secondaryActionTrigger: IButtonAsButton | undefined =
    hasInputs && page > 0
      ? {
          children: 'Back',
          onClick: () => setPage(page === 2 ? 1 : 0),
          disabled: isPending,
          variant: 'secondary',
        }
      : undefined

  return (
    <Modal
      size="lg"
      heading={`Run ${runbookName}${isSubmitView ? '?' : ''}`}
      primaryActionTrigger={primaryActionTrigger}
      secondaryActionTrigger={secondaryActionTrigger}
      {...props}
    >
      <div className="flex flex-col gap-4">
        <FormErrorBanner error={error} fallback={`Unable to run ${runbookName}`} />

        {hasInputs ? (
          <WizardNavComponent
            steps={[
              { id: 'inputs', title: 'Inputs' },
              { id: 'steps', title: 'Steps' },
              { id: 'confirm', title: 'Confirm' },
            ]}
            currentStepIndex={page}
            completedSteps={new Set(['inputs', 'steps'].slice(0, page) as string[])}
            skipHref={null}
            onGoToStep={(index) => {
              if (index <= page) setPage(index as 0 | 1 | 2)
            }}
          />
        ) : null}

        {hasInputs ? (
          <form
            autoComplete="off"
            noValidate
            onSubmit={(e) => e.preventDefault()}
            className={showInputsForm ? 'flex flex-col gap-4' : 'hidden'}
          >
            <Text>Provide inputs for {runbookName}:</Text>
            {inputs.map((input) => (
              <RunbookInputField
                key={input.id ?? input.name}
                form={form}
                input={input}
              />
            ))}
          </form>
        ) : null}

        {isSubmitView ? roleSelector : null}

        {showStepsSummary ? (
          <Expand
            id="runbook-select-steps"
            isOpen
            heading={
              <Text weight="strong">
                Steps ({enabledCount}/{steps.length})
              </Text>
            }
          >
            <div className="flex flex-col gap-1 p-2">
              {steps.map((step, i) => (
                <div
                  key={step.id ?? i}
                  className="flex items-center justify-between gap-2"
                >
                  <CheckboxInput
                    checked={isStepEnabled(step.id)}
                    onChange={(e) =>
                      setStepEnabled((prev) => ({
                        ...prev,
                        [step.id ?? '']: e.target.checked,
                      }))
                    }
                    labelProps={{ labelText: `${i + 1}. ${step.name}` }}
                  />
                  <Badge variant="code" size="sm" theme="neutral">
                    {step.type}
                  </Badge>
                </div>
              ))}
              {noStepsEnabled ? (
                <Text variant="subtext" theme="error">
                  Enable at least one step to run the runbook.
                </Text>
              ) : null}
            </div>
          </Expand>
        ) : null}

        {showInputsSummary ? (
          <>
            <Expand
              id="runbook-review-inputs"
              isOpen
              heading={<Text weight="strong">Inputs ({inputs.length})</Text>}
            >
              <dl className="flex flex-col gap-2 p-2">
                {inputs.map((input) => {
                  const raw = values.inputs?.[input.name]
                  const value = typeof raw === 'boolean' ? String(raw) : (raw ?? '')
                  return (
                    <div
                      key={input.id ?? input.name}
                      className="grid grid-cols-2 gap-2"
                    >
                      <Text as="dt" variant="subtext">
                        {input.display_name || input.name}
                      </Text>
                      <Text as="dd" variant="body">
                        {input.sensitive
                          ? '••••••••'
                          : value !== ''
                            ? value
                            : '—'}
                      </Text>
                    </div>
                  )
                })}
              </dl>
            </Expand>

            <Expand
              id="runbook-review-steps"
              isOpen
              heading={
                <Text weight="strong">
                  Steps ({enabledCount}/{steps.length})
                </Text>
              }
            >
              <ol className="flex flex-col gap-1 p-2">
                {steps.map((step, i) => {
                  const on = isStepEnabled(step.id)
                  return (
                    <li key={step.id ?? i} className="flex items-center gap-2">
                      <Text
                        as="span"
                        variant="body"
                        className={on ? undefined : 'line-through opacity-60'}
                      >
                        {i + 1}. {step.name}
                      </Text>
                      <Badge variant="code" size="sm" theme="neutral">
                        {step.type}
                      </Badge>
                      {!on ? (
                        <Badge size="sm" theme="neutral">
                          skipped
                        </Badge>
                      ) : null}
                    </li>
                  )
                })}
              </ol>
            </Expand>
          </>
        ) : null}
      </div>
    </Modal>
  )
}
