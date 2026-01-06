import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Input } from '@/components/common/form/Input'
import { CodeInput } from '@/components/common/form/CodeInput'
import { Text } from '@/components/common/Text'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import type { TAppInputConfig, TInstall } from '@/types'
import type { IInputConfigFields } from './types'

const FieldWrapper = ({
  children,
  labelText,
  helpText,
}: {
  children: React.ReactElement
  labelText: string | React.ReactElement
  helpText?: string
}) => {
  return (
    <label className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
      <span className="flex flex-col gap-0">
        <Text variant="body" weight="strong">
          {labelText}
        </Text>
        {helpText ? (
          <Text variant="subtext" className="max-w-72">
            {helpText}
          </Text>
        ) : null}
      </span>
      {children}
    </label>
  )
}

const InputGroupFields = ({
  groupInputs,
  install,
}: {
  groupInputs: TAppInputConfig['input_groups'][0]
  install?: TInstall
}) => {
  const installInputs = install ? install?.install_inputs?.at(0)?.values : {}

  const vendorInputs = groupInputs?.app_inputs?.filter(
    (input) => !input?.source || input?.source === 'vendor'
  )

  if (!vendorInputs || vendorInputs.length === 0) {
    return null
  }

  // Sort inputs first, then separate required from optional
  const sortedInputs = vendorInputs.sort(
    (a, b) => (a?.index || 0) - (b?.index || 0)
  )
  const requiredInputs = sortedInputs.filter((input) => input?.required)
  const optionalInputs = sortedInputs.filter((input) => !input?.required)

  const renderInput = (input: (typeof vendorInputs)[0]) => {
    const isBoolean =
      Boolean(input?.default === 'true' || input?.default === 'false') ||
      input?.type === 'bool'

    if (isBoolean) {
      return (
        <div
          key={input?.id}
          className="grid grid-cols-1 md:grid-cols-2 gap-4 items-start"
        >
          <div />
          <div className="ml-1">
            <input type="hidden" name={`inputs:${input?.name}`} value="off" />
            <CheckboxInput
              defaultChecked={
                installInputs?.[input?.name || '']
                  ? Boolean(installInputs?.[input?.name || ''] === 'true')
                  : Boolean(input?.default === 'true')
              }
              labelProps={{
                labelText: input?.display_name || input?.name || '',
                className:
                  'hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0',
              }}
              name={`inputs:${input?.name}`}
            />
          </div>
        </div>
      )
    }

    return (
      <FieldWrapper
        key={input?.id}
        labelText={
          <>
            {input?.display_name || input?.name || ''}
            <Text
              className="ml-1"
              variant="subtext"
              theme={input?.required ? 'error' : 'neutral'}
            >
              {input?.required ? '*' : '(optional)'}
            </Text>
          </>
        }
        helpText={input?.description}
      >
        {input?.type === 'json' ? (
          <CodeInput
            language="json"
            name={`inputs:${input?.name}`}
            required={input?.required}
            defaultValue={installInputs?.[input?.name || ''] || input?.default}
            placeholder="Enter JSON configuration..."
            helperText="Enter valid JSON configuration"
            minHeight={120}
          />
        ) : (
          <Input
            type={
              input?.type === 'number'
                ? 'number'
                : input?.sensitive
                  ? 'password'
                  : 'text'
            }
            autoComplete="off"
            name={`inputs:${input?.name}`}
            required={input?.required}
            defaultValue={installInputs?.[input?.name || ''] || input?.default}
            placeholder={`Enter ${input?.display_name?.toLowerCase() || 'value'}`}
          />
        )}
      </FieldWrapper>
    )
  }

  return (
    <fieldset className="flex flex-col gap-6 border-t pt-6">
      <legend className="flex flex-col gap-0 pr-6">
        <span className="text-lg font-semibold">
          {groupInputs?.display_name}{' '}
          {requiredInputs?.length ? (
            <Text variant="subtext" theme="error">
              (required)
            </Text>
          ) : (
            <Text variant="subtext" theme="neutral">
              (optional)
            </Text>
          )}
        </span>
        <span className="text-sm font-normal">{groupInputs?.description}</span>
      </legend>

      {/* Required fields - always visible */}
      {requiredInputs.map(renderInput)}

      {/* Optional fields - inside Expand component */}
      {optionalInputs.length > 0 && (
        <Expand
          heading="Advanced"
          headerClassName="!px-4 bg-code"
          id={`${groupInputs.id}-advanced`}
          isOpen={!requiredInputs?.length ? true : false}
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

export const InputConfigFields = ({
  inputConfig,
  install,
}: IInputConfigFields) => {
  if (!inputConfig?.input_groups) {
    return null
  }

  return (
    <>
      {inputConfig.input_groups
        .sort((a, b) => (a?.index || 0) - (b?.index || 0))
        .map((group) => (
          <InputGroupFields
            key={group.id}
            groupInputs={group}
            install={install}
          />
        ))}
    </>
  )
}
