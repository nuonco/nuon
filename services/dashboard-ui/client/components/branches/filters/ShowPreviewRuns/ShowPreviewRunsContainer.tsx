import { type ChangeEvent } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import { ShowPreviewRuns } from './ShowPreviewRuns'

export const ShowPreviewRunsContainer = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const showPreviews = searchParams.get('preview') !== 'false'

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set('preview', e.target.checked ? 'true' : 'false')
    params.delete('offset')
    navigate(`?${params.toString()}`, { replace: true })
  }

  return <ShowPreviewRuns showPreviews={showPreviews} onChange={handleChange} />
}
