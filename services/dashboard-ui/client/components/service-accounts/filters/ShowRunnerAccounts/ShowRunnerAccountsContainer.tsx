import { type ChangeEvent } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import { ShowRunnerAccounts } from './ShowRunnerAccounts'

export const ShowRunnerAccountsContainer = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const showRunners = searchParams.get('runners') === 'true'

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const params = new URLSearchParams(searchParams.toString())
    if (e.target.checked) {
      params.set('runners', 'true')
    } else {
      params.delete('runners')
    }
    params.delete('offset')
    navigate(`?${params.toString()}`, { replace: true })
  }

  return <ShowRunnerAccounts showRunners={showRunners} onChange={handleChange} />
}
