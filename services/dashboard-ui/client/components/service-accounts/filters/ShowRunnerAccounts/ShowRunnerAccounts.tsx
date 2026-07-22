import { type ChangeEvent } from 'react'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'

interface IShowRunnerAccounts {
  showRunners: boolean
  onChange: (e: ChangeEvent<HTMLInputElement>) => void
}

export const ShowRunnerAccounts = ({
  showRunners,
  onChange,
}: IShowRunnerAccounts) => {
  return (
    <CheckboxInput
      labelProps={{ labelText: 'Show runner accounts' }}
      checked={showRunners}
      onChange={onChange}
    />
  )
}
