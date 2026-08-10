import type { AnyFieldApi } from '@tanstack/react-form'
import { Input, type IInput } from './Input'
import { fieldErrorMessage } from './field-error'

export interface IFormInput
  extends Omit<
    IInput,
    'value' | 'onChange' | 'onBlur' | 'name' | 'error' | 'errorMessage'
  > {
  field: AnyFieldApi
}

export const FormInput = ({ field, ...props }: IFormInput) => {
  const errorMessage = fieldErrorMessage(field)

  return (
    <Input
      name={field.name}
      value={(field.state.value as string | undefined) ?? ''}
      onChange={(e) => field.handleChange(e.target.value)}
      onBlur={field.handleBlur}
      error={!!errorMessage}
      errorMessage={errorMessage}
      {...props}
    />
  )
}
