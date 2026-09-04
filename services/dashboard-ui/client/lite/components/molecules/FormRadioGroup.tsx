import type { AnyFieldApi } from '@tanstack/react-form'
import { useId, type FieldsetHTMLAttributes, type ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import { Radio } from '../atoms/Radio'
import { Text } from '../atoms/Text'
import { fieldErrorMessage } from './field-error'

export interface IFormRadioOption {
  value: string
  label: ReactNode
  description?: ReactNode
  disabled?: boolean
}

export interface IFormRadioGroup
  extends Omit<FieldsetHTMLAttributes<HTMLFieldSetElement>, 'onChange'> {
  field: AnyFieldApi
  options: IFormRadioOption[]
  label?: ReactNode
  description?: ReactNode
  disabled?: boolean
}

export const FormRadioGroup = ({
  field,
  options,
  label,
  description,
  disabled = false,
  className,
  ...props
}: IFormRadioGroup) => {
  const error = fieldErrorMessage(field)
  const errorId = useId()

  return (
    <fieldset
      disabled={disabled}
      aria-invalid={!!error || undefined}
      aria-describedby={error ? errorId : undefined}
      className={cn('flex min-w-0 flex-col gap-2', className)}
      {...props}
    >
      {label ? (
        <Text as="legend" variant="label" color="secondary" className="mb-1">
          {label}
        </Text>
      ) : null}
      {description ? (
        <Text variant="caption" color="tertiary">
          {description}
        </Text>
      ) : null}
      <span className="flex flex-col gap-2">
        {options.map((option) => (
          <Radio
            key={option.value}
            name={field.name}
            value={option.value}
            checked={field.state.value === option.value}
            disabled={disabled || option.disabled}
            label={option.label}
            description={option.description}
            onChange={() => field.handleChange(option.value)}
            onBlur={field.handleBlur}
          />
        ))}
      </span>
      {error ? (
        <Text id={errorId} variant="caption" className="text-field-invalid">
          {error}
        </Text>
      ) : null}
    </fieldset>
  )
}
