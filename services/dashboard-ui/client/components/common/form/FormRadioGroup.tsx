import type { ReactNode } from 'react'
import type { AnyFieldApi } from '@tanstack/react-form'
import { Label } from './Label'
import { RadioInput } from './RadioInput'
import { Text } from '@/components/common/Text'
import { fieldErrorMessage } from './field-error'

export interface IFormRadioGroupOption {
  value: string
  label: ReactNode
  disabled?: boolean
}

export interface IFormRadioGroup {
  field: AnyFieldApi
  options: IFormRadioGroupOption[]
  label?: string
  disabled?: boolean
}

export const FormRadioGroup = ({
  field,
  options,
  label,
  disabled,
}: IFormRadioGroup) => {
  const errorMessage = fieldErrorMessage(field)

  return (
    <div className="flex flex-col gap-2">
      {label ? <Label>{label}</Label> : null}
      <div className="flex flex-col gap-1">
        {options.map((option) => (
          <RadioInput
            key={option.value}
            name={field.name}
            value={option.value}
            checked={field.state.value === option.value}
            onChange={() => field.handleChange(option.value)}
            onBlur={field.handleBlur}
            disabled={disabled || option.disabled}
            labelProps={{ labelText: option.label }}
          />
        ))}
      </div>
      {errorMessage ? (
        <Text variant="subtext" theme="error">
          {errorMessage}
        </Text>
      ) : null}
    </div>
  )
}
