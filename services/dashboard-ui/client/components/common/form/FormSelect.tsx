import type { AnyFieldApi } from '@tanstack/react-form'
import { Select, type ISelect } from './Select'
import { fieldErrorMessage } from './field-error'

export interface IFormSelect
  extends Omit<
    ISelect,
    'value' | 'onChange' | 'name' | 'error' | 'errorMessage'
  > {
  field: AnyFieldApi
}

export const FormSelect = ({ field, ...props }: IFormSelect) => {
  const errorMessage = fieldErrorMessage(field)

  return (
    <Select
      name={field.name}
      value={(field.state.value as string | undefined) ?? ''}
      onChange={(value) => field.handleChange(value)}
      error={!!errorMessage}
      errorMessage={errorMessage}
      {...props}
    />
  )
}
