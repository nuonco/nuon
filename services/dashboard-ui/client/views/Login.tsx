import { Navigate } from 'react-router'
import { useConfig } from '@/hooks/use-config'
import { useAuth } from '@/hooks/use-auth'

export const Login = () => {
  const { authServiceUrl, appUrl } = useConfig()
  const { isAuthenticated, isLoading } = useAuth()

  if (!isLoading && isAuthenticated) {
    window.location.replace('/')
    return null
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
