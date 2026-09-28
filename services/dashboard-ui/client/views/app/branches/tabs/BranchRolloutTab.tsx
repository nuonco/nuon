import { useEffect } from 'react'
import { useNavigate } from 'react-router'
import { BranchRollout } from '@/components/branches/BranchOverview'
import { useNewAppIA } from '@/hooks/use-new-app-ia'

export const BranchRolloutTab = () => {
  const hasNewAppIA = useNewAppIA()
  const navigate = useNavigate()

  useEffect(() => {
    if (!hasNewAppIA) navigate('..', { replace: true, relative: 'path' })
  }, [hasNewAppIA, navigate])

  return hasNewAppIA ? <BranchRollout /> : null
}
