import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import { CodeBlock } from '@/components/common/CodeBlock'
import type { TAppInputConfig, TInstall } from '@/types'
import type { IInputConfigFields } from './types'

const FieldWrapper = ({ children, labelText, helpText }: {
  children: React.ReactElement
  labelText: string
  helpText?: string
}) => {
  return (
    <label className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
      <span className="flex flex-col gap-0">
        <Text variant="body" weight="strong">{labelText}</Text>
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
  install 
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

  return (
    <fieldset className="flex flex-col gap-6 border-t pt-6">
      <legend className="flex flex-col gap-0 mb-6 pr-6">
        <span className="text-lg font-semibold">
          {groupInputs?.display_name}
        </span>
        <span className="text-sm font-normal">{groupInputs?.description}</span>
      </legend>

      {vendorInputs
        ?.sort((a, b) => (a?.index || 0) - (b?.index || 0))
        ?.map((input) => {
          const isBoolean = Boolean(input?.default === 'true' || input?.default === 'false') || input?.type === 'bool'
          
          if (isBoolean) {
            return (
              <div
                key={input?.id}
                className="grid grid-cols-1 md:grid-cols-2 gap-4 items-start"
              >
                <div />
                <div className="ml-1">
                  <input
                    type="hidden"
                    name={`inputs:${input?.name}`}
                    value="off"
                  />
                  <CheckboxInput
                    defaultChecked={
                      installInputs?.[input?.name || '']
                        ? Boolean(installInputs?.[input?.name || ''] === 'true')
                        : Boolean(input?.default === 'true')
                    }
                    labelProps={{
                      labelText: input?.display_name || input?.name || '',
                      className: 'hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0'
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
              labelText={`${input?.display_name}${input?.required ? ' *' : ''}`}
              helpText={input?.description}
            >
              {input?.type === 'json' ? (
                <div className="flex flex-col gap-2">
                  <textarea
                    className="w-full rounded-md border border-cool-grey-300 dark:border-dark-grey-600 bg-white dark:bg-dark-grey-900 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                    name={`inputs:${input?.name}`}
                    required={input?.required}
                    defaultValue={installInputs?.[input?.name || ''] || input?.default}
                    rows={6}
                    placeholder="Enter JSON configuration..."
                  />
                  <Text variant="subtext" theme="neutral">
                    Enter valid JSON configuration
                  </Text>
                </div>
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
        })}
    </fieldset>
  )
}

export const InputConfigFields = ({ inputConfig, install }: IInputConfigFields) => {
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