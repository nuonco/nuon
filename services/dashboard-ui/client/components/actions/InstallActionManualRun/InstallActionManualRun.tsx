import { useMemo, type ReactNode } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import type { FormValidateOrFn } from '@tanstack/form-core'
import { Button } from '@/components/common/Button'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAction, TActionConfig, TAPIError } from '@/types'
import { buildManualRunSchema, type ManualRunValues } from './schema'

const NUON_MANAGED_ENV_VARS = new Set([
  'role',
  'TRIGGER_TYPE',
  'COMPONENT_ID',
  'COMPONENT_NAME',
  'FLOW_ID',
  'FLOW_TYPE',
  'INSTALL_WORKFLOW_ID',
  'INSTALL_WORKFLOW_TYPE',
  'ROLE_ID',
  'ROLE_ARN',
  'ROLE_NAME',
  'ROLE_TYPE',
  'CHANGE_TYPE',
  'NUON_ORG_ID',
  'NUON_APP_ID',
  'NUON_INSTALL_ID',
  'NUON_API_URL',
  'NUON_API_TOKEN',
])

function normalizeEnvVars(steps: TActionConfig['steps']) {
  return steps.reduce(
    (acc, step) => {
      const keys = Object.keys(step?.env_vars || {})
      keys.forEach((key) => {
        if (!acc[key]) acc[key] = step?.env_vars[key]
      })
      return acc
    },
    {} as Record<string, string>
  )
}

interface IInstallActionManualRunModal extends Omit<IModal, 'onSubmit'> {
  action: TAction
  isLoading: boolean
  isRerun?: boolean
  error?: TAPIError | null
  onSubmit: (vars: Record<string, string>) => void
  roleSelector: ReactNode
  runEnvVars?: Record<string, string>
}

export const InstallActionManualRunModal = ({
  action,
  isLoading,
  isRerun = false,
  error,
  onSubmit,
  roleSelector,
  runEnvVars,
  ...props
}: IInstallActionManualRunModal) => {
  const config = action?.configs?.[0]
  const envVars = useMemo(
    () => normalizeEnvVars(config?.steps || []),
    [config]
  )

  const { initialValues, customFromRun } = useMemo(() => {
    const runEnvVarEntries = Object.entries(runEnvVars ?? {}).filter(
      ([key]) => !NUON_MANAGED_ENV_VARS.has(key)
    )
    const configOverrides = Object.fromEntries(
      runEnvVarEntries.filter(([key]) => key in envVars)
    )
    const customFromRun = runEnvVarEntries.filter(([key]) => !(key in envVars))
    return {
      initialValues: { ...envVars, ...configOverrides },
      customFromRun,
    }
  }, [runEnvVars, envVars])

  const configVarNames = useMemo(
    () => Object.keys(initialValues),
    [initialValues]
  )
  const schema = useMemo(
    () => buildManualRunSchema(configVarNames),
    [configVarNames]
  )
  const validator = schema as unknown as FormValidateOrFn<ManualRunValues>

  const defaultValues = useMemo<ManualRunValues>(
    () => ({
      configVars: initialValues,
      customVars: customFromRun.map(([name, value]) => ({ name, value })),
    }),
    [initialValues, customFromRun]
  )

  const form = useForm({
    defaultValues,
    validators: { onMount: validator, onChange: validator },
    onSubmit: ({ value }) => {
      const vars: Record<string, string> = {}
      Object.entries(value.configVars).forEach(([key, v]) => {
        if (v !== envVars[key]) vars[key] = v
      })
      value.customVars.forEach(({ name, value: v }) => {
        if (name) vars[name] = v
      })
      onSubmit(vars)
    },
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)

  const hasEnvVars = configVarNames.length > 0 || customFromRun.length > 0

  return (
    <Modal
      heading={`${isRerun ? 'Re-run' : 'Run'} action ${action?.name}`}
      size="lg"
      primaryActionTrigger={{
        children: isLoading ? (
          <>
            <Icon variant="Loading" className="animate-spin" />
            {isRerun ? 'Re-running action...' : 'Running action...'}
          </>
        ) : (
          <>
            <Icon variant="PlayIcon" />
            {isRerun ? 'Re-run action' : 'Run action'}
          </>
        ),
        disabled: !canSubmit || isLoading,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(e) => e.preventDefault()}
        className="flex flex-col gap-4"
      >
        <FormErrorBanner error={error} fallback={`Unable to run ${action?.name}`} />

        {roleSelector}

        <Expand
          id="action-env-vars"
          heading={<Text weight="strong">Edit environment variables</Text>}
          className="border rounded-md"
          isOpen={hasEnvVars}
        >
          <div className="p-4 border-t flex flex-col gap-4">
            <Text variant="subtext">
              Edit or add custom environment variables for this manual action
              workflow run.
            </Text>

            {configVarNames.length > 0 && (
              <div className="flex flex-col gap-4">
                {configVarNames.map((name) => (
                  <form.Field key={name} name={`configVars.${name}`}>
                    {(field) => (
                      <FormInput
                        field={field}
                        id={`cfg-${name}`}
                        type="text"
                        disabled={isLoading}
                        labelProps={{ labelText: name }}
                      />
                    )}
                  </form.Field>
                ))}
              </div>
            )}

            <form.Field name="customVars" mode="array">
              {(customVarsField) => (
                <>
                  {customVarsField.state.value.length > 0 && (
                    <div className="flex flex-col gap-2">
                      {customVarsField.state.value.map((_, index) => (
                        <fieldset
                          key={index}
                          className="flex flex-col gap-2 py-2 border-t relative"
                        >
                          <legend className="text-base font-medium pr-2 mb-2 flex items-center justify-between">
                            <span>Custom env var {index + 1}</span>
                            <Button
                              type="button"
                              variant="ghost"
                              onClick={() => customVarsField.removeValue(index)}
                              className="ml-2 !p-2"
                              size="sm"
                              disabled={isLoading}
                              aria-label={`Remove custom env var ${index + 1}`}
                            >
                              <Icon variant="XIcon" size="12" />
                            </Button>
                          </legend>
                          <form.Field name={`customVars[${index}].name`}>
                            {(field) => (
                              <FormInput
                                field={field}
                                type="text"
                                disabled={isLoading}
                                labelProps={{ labelText: 'Name' }}
                              />
                            )}
                          </form.Field>
                          <form.Field name={`customVars[${index}].value`}>
                            {(field) => (
                              <FormInput
                                field={field}
                                type="text"
                                disabled={isLoading}
                                labelProps={{ labelText: 'Value' }}
                              />
                            )}
                          </form.Field>
                        </fieldset>
                      ))}
                    </div>
                  )}

                  <div>
                    <Button
                      type="button"
                      variant="ghost"
                      disabled={isLoading}
                      onClick={() =>
                        customVarsField.pushValue({ name: '', value: '' })
                      }
                    >
                      <Icon variant="PlusIcon" />
                      Add environment variable
                    </Button>
                  </div>
                </>
              )}
            </form.Field>
          </div>
        </Expand>
      </form>
    </Modal>
  )
}
