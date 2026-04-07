import { type ChangeEvent } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import { ShowDriftScan } from './ShowDriftScan'

export const ShowDriftScanContainer = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const showDrifts = searchParams.get('drifts') !== 'false'

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set('drifts', e.target.checked ? 'true' : 'false')
    params.delete('offset')
    navigate(`?${params.toString()}`, { replace: true })
  }

  return <ShowDriftScan showDrifts={showDrifts} onChange={handleChange} />
}
