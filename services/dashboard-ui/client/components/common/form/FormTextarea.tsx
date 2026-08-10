import type { AnyFieldApi } from '@tanstack/react-form'
import { Textarea, type ITextarea } from './Textarea'
import { fieldErrorMessage } from './field-error'

export interface IFormTextarea
  extends Omit<
    ITextarea,
    'value' | 'onChange' | 'onBlur' | 'name' | 'error' | 'errorMessage'
  > {
  field: AnyFieldApi
}

export const FormTextarea = ({ field, ...props }: IFormTextarea) => {
  const errorMessage = fieldErrorMessage(field)

  return (
    <Textarea
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
