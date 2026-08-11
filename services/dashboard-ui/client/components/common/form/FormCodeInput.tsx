import type { AnyFieldApi } from '@tanstack/react-form'
import { CodeInput, type ICodeInput } from './CodeInput'
import { fieldErrorMessage } from './field-error'

export interface IFormCodeInput
  extends Omit<
    ICodeInput,
    'value' | 'onChange' | 'name' | 'error' | 'errorMessage'
  > {
  field: AnyFieldApi
}

export const FormCodeInput = ({ field, ...props }: IFormCodeInput) => {
  const errorMessage = fieldErrorMessage(field)

  return (
    <CodeInput
      name={field.name}
      value={(field.state.value as string | undefined) ?? ''}
      onChange={(e) => field.handleChange(e.target.value)}
      error={!!errorMessage}
      errorMessage={errorMessage}
      {...props}
    />
  )
}
