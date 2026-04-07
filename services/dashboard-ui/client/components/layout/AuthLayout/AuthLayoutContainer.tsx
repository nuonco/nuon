import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { AuthLayout } from './AuthLayout'

export const AuthLayoutContainer = () => {
  const { authServiceUrl, appUrl } = useConfig()
  const { isAuthenticated, isLoading, error } = useAuth()

  if (!isLoading && !isAuthenticated && !error) {
    window.location.href = `${authServiceUrl}/?url=${appUrl}`
  }

  return (
    <AuthLayout
      isLoading={isLoading}
      isAuthenticated={!!isAuthenticated}
      hasError={!!error}
      onRetry={() => window.location.reload()}
    />
  )
}
