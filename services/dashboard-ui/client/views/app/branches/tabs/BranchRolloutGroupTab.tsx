import { useEffect } from 'react'
import { useNavigate, useParams } from 'react-router'
import { BranchRolloutGroup } from '@/components/branches/BranchOverview'
import { useNewAppIA } from '@/hooks/use-new-app-ia'

export const BranchRolloutGroupTab = () => {
  const hasNewAppIA = useNewAppIA()
  const navigate = useNavigate()
  const params = useParams()

  useEffect(() => {
    if (hasNewAppIA) return
    navigate(
      `/${params.orgId}/apps/${params.appId}/branches/${params.branchId}`,
      { replace: true }
    )
  }, [hasNewAppIA, navigate, params.orgId, params.appId, params.branchId])

  return hasNewAppIA ? <BranchRolloutGroup /> : null
}
