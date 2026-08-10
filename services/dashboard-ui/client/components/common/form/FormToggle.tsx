import type { AnyFieldApi } from '@tanstack/react-form'
import { Toggle, type IToggle } from './Toggle'

export interface IFormToggle extends Omit<IToggle, 'checked' | 'onChange'> {
  field: AnyFieldApi
}

export const FormToggle = ({ field, ...props }: IFormToggle) => (
  <Toggle
    checked={!!field.state.value}
    onChange={(checked) => field.handleChange(checked)}
    {...props}
  />
)
