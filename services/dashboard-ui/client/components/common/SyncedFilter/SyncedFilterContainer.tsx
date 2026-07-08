import { type ChangeEvent } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import { SyncedFilter } from './SyncedFilter'

export const SyncedFilterContainer = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const showSynced = searchParams.get('synced') !== 'false'

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set('synced', e.target.checked ? 'true' : 'false')
    params.delete('offset')
    navigate(`?${params.toString()}`, { replace: true })
  }

  return <SyncedFilter showSynced={showSynced} onChange={handleChange} />
}
