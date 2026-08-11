import { useCallback, useEffect, useMemo, useRef, type ReactNode } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormCodeInput } from '@/components/common/form/FormCodeInput'
import { FormInput } from '@/components/common/form/FormInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useDraftPersistence } from '@/hooks/use-draft-persistence'
import type { TRunAdhocActionBody } from '@/lib'
import type { TAPIError } from '@/types'
import { runAdhocActionSchema, type RunAdhocActionValues } from './schema'

interface IRunAdhocActionModal extends Omit<IModal, 'onSubmit'> {
  installId: string
  initialValues?: TRunAdhocActionBody
  isPending: boolean
  error: TAPIError | null
  onSubmit: (body: TRunAdhocActionBody) => void
  onDraftResume: (
    onResume: () => void,
    onStartFresh: () => void,
    onClose: () => void,
    draftTimestamp: string
  ) => void
  roleSelector: ReactNode
}

const toBody = (value: RunAdhocActionValues): TRunAdhocActionBody => {
  const env_vars = value.envVars.reduce(
    (acc, { name, value: v }) => {
      if (name) acc[name] = v
      return acc
    },
    {} as Record<string, string>
  )

  const body: TRunAdhocActionBody = {
    name: value.name || undefined,
    timeout: value.timeout ? Number(value.timeout) : undefined,
    env_vars: Object.keys(env_vars).length > 0 ? env_vars : undefined,
  }

  if (value.inputMode === 'command') {
    body.command = value.command
  } else {
    body.inline_contents = value.inlineContents
  }

  return body
}

export const RunAdhocActionModal = ({
  installId,
  initialValues,
  isPending,
  error,
  onSubmit,
  onDraftResume,
  roleSelector,
  ...props
}: IRunAdhocActionModal) => {
  const draftShownRef = useRef(false)
  const clearDraftRef = useRef<() => void>(() => {})

  const defaultValues = useMemo<RunAdhocActionValues>(
    () => ({
      name: initialValues?.name ?? '',
      inputMode: initialValues?.inline_contents ? 'script' : 'command',
      command: initialValues?.command ?? '',
      inlineContents: initialValues?.inline_contents ?? '',
      timeout: initialValues?.timeout?.toString() ?? '300',
      envVars: Object.entries(initialValues?.env_vars ?? {}).map(
        ([name, value]) => ({ name, value })
      ),
    }),
    [initialValues]
  )

  const form = useForm({
    defaultValues,
    validators: { onMount: runAdhocActionSchema, onChange: runAdhocActionSchema },
    onSubmit: ({ value }) => {
      clearDraftRef.current()
      onSubmit(toBody(value))
    },
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const values = useStore(form.store, (s) => s.values)
  const inputMode = values.inputMode

  const { hasDraft, draftTimestamp, draftValues, clearDraft } =
    useDraftPersistence<RunAdhocActionValues>({
      storageKey: `adhoc-action-draft:${installId}`,
      values,
      enabled: true,
    })
  clearDraftRef.current = clearDraft

  const restoreDraft = useCallback(() => {
    if (draftValues) form.reset({ ...defaultValues, ...draftValues })
  }, [form, defaultValues, draftValues])

  useEffect(() => {
    if (hasDraft && !draftShownRef.current && draftTimestamp) {
      draftShownRef.current = true
      onDraftResume(
        () => restoreDraft(),
        () => {
          clearDraft()
          draftShownRef.current = false
        },
        () => {},
        draftTimestamp
      )
    }
  }, [hasDraft, draftTimestamp, restoreDraft, clearDraft, onDraftResume])

  return (
    <Modal
      className="!max-h-[80vh]"
      childrenClassName="flex-auto overflow-y-auto"
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="TerminalWindowIcon" size="24" />
          Run adhoc action
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Running action
          </span>
        ) : initialValues ? (
          'Rerun action'
        ) : (
          'Run action'
        ),
        onClick: () => form.handleSubmit(),
        disabled: !canSubmit || isPending,
        variant: 'primary',
      }}
      size="lg"
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(e) => e.preventDefault()}
        className="flex flex-col gap-4"
      >
        <FormErrorBanner error={error} fallback="Unable to run adhoc action" />

        <form.Field name="name">
          {(field) => (
            <FormInput
              field={field}
              id="adhoc-name"
              type="text"
              placeholder="Display name for this action"
              maxLength={255}
              disabled={isPending}
              labelProps={{ labelText: 'Name (optional)' }}
            />
          )}
        </form.Field>

        <div className="flex gap-2">
          <Button
            type="button"
            variant={inputMode === 'command' ? 'primary' : 'secondary'}
            size="sm"
            disabled={isPending}
            onClick={() => form.setFieldValue('inputMode', 'command')}
          >
            Single command
          </Button>
          <Button
            type="button"
            variant={inputMode === 'script' ? 'primary' : 'secondary'}
            size="sm"
            disabled={isPending}
            onClick={() => form.setFieldValue('inputMode', 'script')}
          >
            Bash script
          </Button>
        </div>

        {inputMode === 'command' ? (
          <form.Field name="command">
            {(field) => (
              <FormInput
                field={field}
                id="adhoc-command"
                type="text"
                placeholder="echo 'Hello, world!'"
                className="!font-mono"
                disabled={isPending}
                labelProps={{ labelText: 'Command *' }}
                helperText="Single-line shell command to execute"
              />
            )}
          </form.Field>
        ) : (
          <form.Field name="inlineContents">
            {(field) => (
              <FormCodeInput
                field={field}
                language="bash"
                minHeight={200}
                disabled={isPending}
                labelProps={{ labelText: 'Bash script *' }}
                helperText="Multi-line bash script to execute"
              />
            )}
          </form.Field>
        )}

        <form.Field name="timeout">
          {(field) => (
            <FormInput
              field={field}
              id="adhoc-timeout"
              type="number"
              min={1}
              max={3600}
              disabled={isPending}
              labelProps={{ labelText: 'Timeout (seconds)' }}
              helperText="Execution timeout (1-3600 seconds, default: 300)"
            />
          )}
        </form.Field>

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <Text variant="label" weight="strong">
              Environment variables
            </Text>
            <form.Field name="envVars" mode="array">
              {(envVarsField) => (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  disabled={isPending}
                  onClick={() =>
                    envVarsField.pushValue({ name: '', value: '' })
                  }
                >
                  <Icon variant="PlusIcon" size="16" />
                  Add variable
                </Button>
              )}
            </form.Field>
          </div>

          <form.Field name="envVars" mode="array">
            {(envVarsField) =>
              envVarsField.state.value.length === 0 ? (
                <Text variant="subtext">No environment variables added</Text>
              ) : (
                <>
                  {envVarsField.state.value.map((_, idx) => (
                    <fieldset
                      key={idx}
                      className="grid grid-cols-[1fr_1fr_auto] gap-2 items-end border-t pt-2"
                    >
                      <form.Field name={`envVars[${idx}].name`}>
                        {(field) => (
                          <FormInput
                            field={field}
                            type="text"
                            placeholder="VAR_NAME"
                            disabled={isPending}
                            labelProps={{ labelText: 'Name' }}
                          />
                        )}
                      </form.Field>
                      <form.Field name={`envVars[${idx}].value`}>
                        {(field) => (
                          <FormInput
                            field={field}
                            type="text"
                            placeholder="value"
                            disabled={isPending}
                            labelProps={{ labelText: 'Value' }}
                          />
                        )}
                      </form.Field>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        disabled={isPending}
                        onClick={() => envVarsField.removeValue(idx)}
                        className="mb-1"
                      >
                        <Icon variant="XIcon" size="16" />
                      </Button>
                    </fieldset>
                  ))}
                </>
              )
            }
          </form.Field>
        </div>

        {roleSelector}
      </form>
    </Modal>
  )
}
