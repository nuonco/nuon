import { type ChangeEvent } from 'react'
import { useSyncedOnlyFilter } from '@/hooks/use-synced-only-filter'
import { SyncedFilter } from './SyncedFilter'

export const SyncedFilterContainer = () => {
  const { syncedOnly, setSyncedOnly } = useSyncedOnlyFilter()

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    setSyncedOnly(e.target.checked)
  }

  return <SyncedFilter syncedOnly={syncedOnly} onChange={handleChange} />
}
