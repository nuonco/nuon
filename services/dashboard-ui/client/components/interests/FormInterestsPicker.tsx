import type { AnyFieldApi } from '@tanstack/react-form'
import { InterestsPicker } from './InterestsPicker'
import type { PresetModalOutput } from './InterestsModal'

export interface IFormInterestsPicker {
  field: AnyFieldApi
  matchField?: AnyFieldApi
  disabled?: boolean
}

export const FormInterestsPicker = ({
  field,
  matchField,
  disabled,
}: IFormInterestsPicker) => (
  <InterestsPicker
    value={field.state.value}
    matchValue={matchField?.state.value}
    onChange={({ interests, match }: PresetModalOutput) => {
      field.handleChange(interests)
      if (matchField) {
        matchField.handleChange(match)
      }
    }}
    disabled={disabled}
  />
)
