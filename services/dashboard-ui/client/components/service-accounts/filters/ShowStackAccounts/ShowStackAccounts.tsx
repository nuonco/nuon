import { type ChangeEvent } from 'react'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'

interface IShowStackAccounts {
  showStacks: boolean
  onChange: (e: ChangeEvent<HTMLInputElement>) => void
}

export const ShowStackAccounts = ({
  showStacks,
  onChange,
}: IShowStackAccounts) => {
  return (
    <CheckboxInput
      labelProps={{ labelText: 'Show stack accounts' }}
      checked={showStacks}
      onChange={onChange}
    />
  )
}
