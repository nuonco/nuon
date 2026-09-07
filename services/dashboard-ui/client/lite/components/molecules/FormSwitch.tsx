import type { AnyFieldApi } from '@tanstack/react-form'
import { Switch, type ISwitch } from '../atoms/Switch'
import { fieldErrorMessage } from './field-error'

export interface IFormSwitch
  extends Omit<ISwitch, 'checked' | 'onChange' | 'onBlur' | 'name' | 'error'> {
  field: AnyFieldApi
}

export const FormSwitch = ({ field, ...props }: IFormSwitch) => (
  <Switch
    name={field.name}
    checked={Boolean(field.state.value)}
    onChange={(checked) => field.handleChange(checked)}
    onBlur={field.handleBlur}
    error={fieldErrorMessage(field)}
    {...props}
  />
)
