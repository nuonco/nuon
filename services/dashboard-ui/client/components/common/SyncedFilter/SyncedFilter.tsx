import { type ChangeEvent } from 'react'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'

interface ISyncedFilter {
  showSynced: boolean
  onChange: (e: ChangeEvent<HTMLInputElement>) => void
}

export const SyncedFilter = ({ showSynced, onChange }: ISyncedFilter) => {
  return (
    <CheckboxInput
      labelProps={{ labelText: 'Synced' }}
      checked={showSynced}
      onChange={onChange}
    />
  )
}
