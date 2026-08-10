import type { AnyFieldApi } from '@tanstack/react-form'
import { MatchPicker } from './MatchPicker'

export interface IFormMatchPicker {
  field: AnyFieldApi
  disabled?: boolean
}

export const FormMatchPicker = ({ field, disabled }: IFormMatchPicker) => (
  <MatchPicker
    value={field.state.value}
    onChange={(next) => field.handleChange(next)}
    disabled={disabled}
  />
)
