import { type ChangeEvent } from 'react'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'

interface ISyncedFilter {
  syncedOnly: boolean
  onChange: (e: ChangeEvent<HTMLInputElement>) => void
}

export const SyncedFilter = ({ syncedOnly, onChange }: ISyncedFilter) => {
  return (
    <CheckboxInput
      labelProps={{ labelText: 'Synced only' }}
      checked={syncedOnly}
      onChange={onChange}
    />
  )
}
