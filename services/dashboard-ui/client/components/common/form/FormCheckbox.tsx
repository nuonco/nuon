import type { AnyFieldApi } from '@tanstack/react-form'
import { CheckboxInput, type ICheckboxInput } from './CheckboxInput'

export interface IFormCheckbox
  extends Omit<
    ICheckboxInput,
    'checked' | 'onChange' | 'onBlur' | 'name' | 'value'
  > {
  field: AnyFieldApi
}

export const FormCheckbox = ({ field, ...props }: IFormCheckbox) => (
  <CheckboxInput
    name={field.name}
    checked={!!field.state.value}
    onChange={(e) => field.handleChange(e.target.checked)}
    onBlur={field.handleBlur}
    {...props}
  />
)
