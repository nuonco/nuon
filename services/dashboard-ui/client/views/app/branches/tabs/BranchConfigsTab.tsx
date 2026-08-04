import { useEffect } from 'react'
import { useNavigate, useParams } from 'react-router'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { BranchConfigs } from '../BranchConfigs'

export const BranchConfigsTab = () => {
  const hasNewAppIA = useNewAppIA()
  const navigate = useNavigate()
  const params = useParams()

  useEffect(() => {
    if (!hasNewAppIA) return
    navigate(
      `/${params.orgId}/apps/${params.appId}/branches/${params.branchId}`,
      { replace: true }
    )
  }, [hasNewAppIA, navigate, params.orgId, params.appId, params.branchId])

  if (!hasNewAppIA) return <BranchConfigs />

  return null
}
