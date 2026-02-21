import { useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

export default function OrgDashboard() {
  const { orgId } = useParams()
  const navigate = useNavigate()

  useEffect(() => {
    if (orgId) {
      navigate(`/${orgId}/apps`, { replace: true })
    }
  }, [orgId, navigate])

  return null
}
