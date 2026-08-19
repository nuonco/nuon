import { useNavigate, useSearchParams } from 'react-router'

const STORAGE_KEY = 'nuon:synced-only-filter'

function readStored(): boolean {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored !== null) return stored === 'true'
  } catch {}
  return true
}

export function useSyncedOnlyFilter() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const param = searchParams.get('synced_only')
  const syncedOnly = param !== null ? param === 'true' : readStored()

  const setSyncedOnly = (checked: boolean) => {
    try {
      localStorage.setItem(STORAGE_KEY, String(checked))
    } catch {}
    const params = new URLSearchParams(searchParams.toString())
    if (checked) {
      params.set('synced_only', 'true')
    } else {
      params.delete('synced_only')
    }
    params.delete('offset')
    navigate(`?${params.toString()}`, { replace: true })
  }

  return { syncedOnly, setSyncedOnly }
}
