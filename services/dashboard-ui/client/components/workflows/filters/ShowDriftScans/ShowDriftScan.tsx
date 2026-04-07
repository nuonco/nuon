import { type ChangeEvent } from 'react'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'

interface IShowDriftScan {
  showDrifts: boolean
  onChange: (e: ChangeEvent<HTMLInputElement>) => void
}

export const ShowDriftScan = ({ showDrifts, onChange }: IShowDriftScan) => {
  return (
    <CheckboxInput
      labelProps={{ labelText: 'Drift scan' }}
      checked={showDrifts}
      onChange={onChange}
    />
  )
}
