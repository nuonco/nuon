import { type ChangeEvent } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import { SyncedFilter } from './SyncedFilter'

export const SyncedFilterContainer = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const syncedOnly = searchParams.get('synced_only') === 'true'

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const params = new URLSearchParams(searchParams.toString())
    if (e.target.checked) {
      params.set('synced_only', 'true')
    } else {
      params.delete('synced_only')
    }
    params.delete('offset')
    navigate(`?${params.toString()}`, { replace: true })
  }

  return <SyncedFilter syncedOnly={syncedOnly} onChange={handleChange} />
}
