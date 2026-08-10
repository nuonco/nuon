import type { AnyFieldApi } from '@tanstack/react-form'
import { InterestsPicker } from './InterestsPicker'

export interface IFormInterestsPicker {
  field: AnyFieldApi
  disabled?: boolean
}

export const FormInterestsPicker = ({
  field,
  disabled,
}: IFormInterestsPicker) => (
  <InterestsPicker
    value={field.state.value}
    onChange={(next) => field.handleChange(next)}
    disabled={disabled}
  />
)
