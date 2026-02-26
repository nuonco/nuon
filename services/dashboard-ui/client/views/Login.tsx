import { Navigate } from 'react-router'
import { getOrgSession } from '@/lib/cookies'
import { useConfig } from '@/hooks/use-config'
import { useAuth } from '@/hooks/use-auth'

export const Login = () => {
  const { authServiceUrl, appUrl } = useConfig()
  const { isAuthenticated, isLoading } = useAuth()

  if (!isLoading && isAuthenticated) {
    const orgId = getOrgSession()
    return <Navigate to={orgId ? `/${orgId}/apps` : '/onboarding'} replace />
  }

  return (
    <div>
      <button
        onClick={() => {
          window.location.href = `${authServiceUrl}/?url=${appUrl}`
        }}
      >
        Sign in
      </button>
    </div>
  )
}
