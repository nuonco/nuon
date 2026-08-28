import { type ChangeEvent } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import { ShowStackAccounts } from './ShowStackAccounts'

export const ShowStackAccountsContainer = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const showStacks = searchParams.get('stacks') === 'true'

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const params = new URLSearchParams(searchParams.toString())
    if (e.target.checked) {
      params.set('stacks', 'true')
    } else {
      params.delete('stacks')
    }
    params.delete('offset')
    navigate(`?${params.toString()}`, { replace: true })
  }

  return <ShowStackAccounts showStacks={showStacks} onChange={handleChange} />
}
