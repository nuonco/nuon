import { Expand } from '@/components/common/Expand'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { FormCodeInput } from '@/components/common/form/FormCodeInput'
import { FormInput } from '@/components/common/form/FormInput'
import {
  COMPONENT_OVERRIDE_INPUT_GROUP,
  groupComponentOverrideInputs,
} from '@/utils/install-utils'
import type { TAppInput, TAppInputConfig, TInstall } from '@/types'
import { FieldRow } from './FieldRow'
import { FormComponentOverrideCard } from './FormComponentOverrideCard'
import { isBooleanInput, type InstallInputFieldName } from './schema'
import type { InstallFormApi } from './useInstallForm'

const CODE_INPUT_TYPES = {
  json: { language: 'json', placeholder: 'Enter JSON configuration...' },
  yaml: { language: 'yaml', placeholder: 'Enter YAML configuration...' },
  hcl: { language: 'hcl', placeholder: 'Enter HCL (.tfvars) configuration...' },
} as const

type CodeInputType = keyof typeof CODE_INPUT_TYPES

type InputGroup = NonNullable<TAppInputConfig['input_groups']>[number]

const EditableInput = ({
  form,
  input,
  disabled,
}: {
  form: InstallFormApi
  input: TAppInput
  disabled?: boolean
}) => {
  const name = `inputs.${input.name}` as InstallInputFieldName
  const label = input?.display_name || input?.name || ''

  if (isBooleanInput(input)) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 items-start">
        <div />
        <div className="ml-1">
          <form.Field name={name}>
            {(field) => (
              <FormCheckbox
                field={field}
                disabled={disabled}
                labelProps={{
                  labelText: label,
                  className:
                    'hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0',
                }}
              />
            )}
          </form.Field>
        </div>
      </div>
    )
  }

  const codeType =
    input?.type && input.type in CODE_INPUT_TYPES
      ? CODE_INPUT_TYPES[input.type as CodeInputType]
      : undefined

  return (
    <FieldRow
      labelText={label}
      required={input?.required}
      optional={!input?.required}
      helpText={input?.description}
    >
      {codeType ? (
        <form.Field name={name}>
          {(field) => (
            <FormCodeInput
              field={field}
              language={codeType.language}
              placeholder={codeType.placeholder}
              minHeight={120}
              disabled={disabled}
            />
          )}
        </form.Field>
      ) : (
        <form.Field name={name}>
          {(field) => (
            <FormInput
              field={field}
              type={
                input?.type === 'number'
                  ? 'number'
                  : input?.sensitive
                    ? 'password'
                    : 'text'
              }
              autoComplete="off"
              placeholder={`Enter ${input?.display_name?.toLowerCase() || 'value'}`}
              disabled={disabled}
            />
          )}
        </form.Field>
      )}
    </FieldRow>
  )
}

const CustomerInput = ({
  input,
  install,
}: {
  input: TAppInput
  install?: TInstall
}) => {
  const value = install?.install_inputs?.at(0)?.values?.[input?.name || '']
  const label = input?.display_name || input?.name || ''

  return (
    <FieldRow
      labelText={
        <>
          {label}
          {input?.required && (
            <Text className="ml-1" variant="subtext" theme="warn">
              (required by customer)
            </Text>
          )}
        </>
      }
      helpText={input?.description}
    >
      <Input value={value ?? input?.default ?? ''} disabled readOnly />
    </FieldRow>
  )
}

const InputGroupFields = ({
  form,
  group,
  install,
  disabled,
}: {
  form: InstallFormApi
  group: InputGroup
  install?: TInstall
  disabled?: boolean
}) => {
  const allInputs = [...(group?.app_inputs || [])].sort(
    (a, b) => (a?.index || 0) - (b?.index || 0)
  )
  if (allInputs.length === 0) return null

  const requiredInputs = allInputs.filter((input) => input?.required)
  const optionalInputs = allInputs.filter((input) => !input?.required)

  const renderInput = (input: TAppInput) =>
    disabled ? (
      <CustomerInput key={input?.id} input={input} install={install} />
    ) : (
      <EditableInput key={input?.id} form={form} input={input} />
    )

  return (
    <fieldset className="flex flex-col gap-6 border-t pt-6">
      <legend className="flex flex-col gap-0 pr-6">
        <span className="text-lg font-semibold">
          {group?.display_name}{' '}
          {!disabled && (
            <Text
              variant="subtext"
              theme={requiredInputs.length ? 'error' : 'neutral'}
            >
              {requiredInputs.length ? '(required)' : '(optional)'}
            </Text>
          )}
        </span>
        <span className="text-sm font-normal">{group?.description}</span>
      </legend>

      {requiredInputs.map(renderInput)}

      {optionalInputs.length > 0 && (
        <Expand
          heading="Advanced"
          headerClassName="!px-4 bg-code"
          id={`${group.id}-advanced`}
          isOpen={!requiredInputs.length}
          className="mt-2 border rounded-md"
        >
          <div className="flex flex-col gap-6 p-4 border-t">
            {optionalInputs.map(renderInput)}
          </div>
        </Expand>
      )}
    </fieldset>
  )
}

const ComponentOverridesFieldset = ({
  form,
  group,
}: {
  form: InstallFormApi
  group: InputGroup
}) => {
  const cards = groupComponentOverrideInputs(group?.app_inputs || [])
  if (cards.length === 0) return null

  return (
    <fieldset className="flex flex-col gap-6 border-t pt-6">
      <legend className="flex flex-col gap-0 pr-6">
        <span className="text-lg font-semibold">Components</span>
        <span className="text-sm font-normal">
          Choose which components deploy on this install and customize each one.
        </span>
      </legend>

      <div className="flex flex-col gap-4">
        {cards.map((card) => (
          <FormComponentOverrideCard
            key={card.component}
            form={form}
            card={card}
          />
        ))}
      </div>
    </fieldset>
  )
}

interface IInstallInputFields {
  form: InstallFormApi
  inputConfig?: TAppInputConfig
  install?: TInstall
}

export const InstallInputFields = ({
  form,
  inputConfig,
  install,
}: IInstallInputFields) => {
  if (!inputConfig?.input_groups) return null

  const sortedGroups = [...inputConfig.input_groups].sort(
    (a, b) => (a?.index || 0) - (b?.index || 0)
  )

  const vendorGroups: { group: InputGroup; inputs: TAppInput[] }[] = []
  const customerGroups: { group: InputGroup; inputs: TAppInput[] }[] = []

  for (const group of sortedGroups) {
    const vendorInputs =
      group?.app_inputs?.filter(
        (input) => !input?.source || input?.source === 'vendor'
      ) || []
    const customerInputs =
      group?.app_inputs?.filter((input) => input?.source === 'customer') || []

    if (vendorInputs.length > 0)
      vendorGroups.push({ group, inputs: vendorInputs })
    if (customerInputs.length > 0)
      customerGroups.push({ group, inputs: customerInputs })
  }

  return (
    <>
      {vendorGroups.map(({ group, inputs }) =>
        group.name === COMPONENT_OVERRIDE_INPUT_GROUP ? (
          <ComponentOverridesFieldset
            key={`vendor-${group.id}`}
            form={form}
            group={{ ...group, app_inputs: inputs }}
          />
        ) : (
          <InputGroupFields
            key={`vendor-${group.id}`}
            form={form}
            group={{ ...group, app_inputs: inputs }}
          />
        )
      )}

      {customerGroups.length > 0 && (
        <>
          <div className="flex flex-col gap-2 border-t pt-6 mt-6">
            <span className="text-lg font-semibold text-cool-grey-600 dark:text-cool-grey-400">
              Customer configuration
            </span>
            <span className="text-sm font-normal text-cool-grey-500 dark:text-cool-grey-500">
              These fields are configured by the customer and cannot be edited
              here.
            </span>
          </div>
          {customerGroups.map(({ group, inputs }) => (
            <InputGroupFields
              key={`customer-${group.id}`}
              form={form}
              group={{ ...group, app_inputs: inputs }}
              install={install}
              disabled
            />
          ))}
        </>
      )}
    </>
  )
}
