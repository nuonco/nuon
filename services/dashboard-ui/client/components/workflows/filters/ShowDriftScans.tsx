import { useNavigate, useSearchParams } from 'react-router'
import { type ChangeEvent } from 'react'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'

export const ShowDriftScan = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const showDrifts = searchParams.get('drifts') !== 'false'

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set('drifts', e.target.checked ? 'true' : 'false')
    params.delete('offset')
    navigate(`?${params.toString()}`, { replace: true })
  }

  return (
    <CheckboxInput
      labelProps={{ labelText: 'Drift scan' }}
      checked={showDrifts}
      onChange={handleChange}
    />
  )
}
