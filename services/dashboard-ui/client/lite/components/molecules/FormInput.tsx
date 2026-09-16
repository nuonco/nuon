import { forwardRef } from 'react'
import type { AnyFieldApi } from '@tanstack/react-form'
import { Input, type IInput } from '../atoms/Input'
import { Field, type IField } from './Field'
import { fieldErrorMessage } from './field-error'

export interface IFormInput
  extends Omit<
      IInput,
      'value' | 'defaultValue' | 'onChange' | 'onBlur' | 'name'
    >,
    Pick<IField, 'label' | 'description' | 'error' | 'optional' | 'className'> {
  field: AnyFieldApi
}

export const FormInput = forwardRef<HTMLInputElement, IFormInput>(
  (
    { field, label, description, error, optional, className, ...props },
    ref
  ) => (
    <Field
      label={label}
      description={description}
      optional={optional}
      error={error ?? fieldErrorMessage(field)}
      className={className}
    >
      <Input
        ref={ref}
        name={field.name}
        value={(field.state.value as string | undefined) ?? ''}
        onChange={(event) => field.handleChange(event.target.value)}
        onBlur={field.handleBlur}
        {...props}
      />
    </Field>
  )
)

FormInput.displayName = 'FormInput'
