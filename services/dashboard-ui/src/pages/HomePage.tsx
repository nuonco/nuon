import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { getCookie } from '@/utils/cookies'
import { apiClient } from '@/lib/api-client'
import { useAuth } from '@/hooks/use-auth'
import type { TOrg } from '@/types'

export default function HomePage() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    async function handleOrgRedirect() {
      if (!user) {
        setIsLoading(false)
        return
      }

      try {
        const orgIdFromCookie = getCookie('nuon-org-id')

        if (orgIdFromCookie) {
          const { data: org, error } = await apiClient<TOrg>({
            path: `/api/ctl-api/v1/orgs/current`,
          })

          if (org && !error) {
            navigate(`/${orgIdFromCookie}/apps`, { replace: true })
            return
          }
        }

        // Fetch first org
        const { data: orgs } = await apiClient<TOrg[]>({
          path: '/api/ctl-api/v1/orgs?limit=1',
        })

        if (orgs && orgs.length > 0) {
          navigate(`/${orgs[0].id}/apps`, { replace: true })
          return
        }

        // No orgs - show placeholder
        setIsLoading(false)
      } catch (error) {
        console.error('Error redirecting to org:', error)
        setIsLoading(false)
      }
    }

    handleOrgRedirect()
  }, [user, navigate])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900" />
      </div>
    )
  }

  return (
    <div className="flex items-center justify-center h-screen">
      <div className="text-center">
        <h1 className="text-2xl font-bold mb-4">Welcome to Nuon</h1>
        <p className="text-gray-600">No organizations found. Create one to get started.</p>
      </div>
    </div>
  )
}
