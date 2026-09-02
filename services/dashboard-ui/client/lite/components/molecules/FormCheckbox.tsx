import type { AnyFieldApi } from '@tanstack/react-form'
import { Checkbox, type ICheckbox } from '../atoms/Checkbox'
import { fieldErrorMessage } from './field-error'

export interface IFormCheckbox
  extends Omit<
    ICheckbox,
    'checked' | 'defaultChecked' | 'onChange' | 'onBlur' | 'name' | 'error'
  > {
  field: AnyFieldApi
}

export const FormCheckbox = ({ field, ...props }: IFormCheckbox) => (
  <Checkbox
    name={field.name}
    checked={Boolean(field.state.value)}
    onChange={(event) => field.handleChange(event.target.checked)}
    onBlur={field.handleBlur}
    error={fieldErrorMessage(field)}
    {...props}
  />
)
