import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { InitPylonChat } from '@/lib/pylon-chat'
import { AuthLayout } from './AuthLayout'

export const AuthLayoutContainer = () => {
  const { authServiceUrl, appUrl, pylonAppId } = useConfig()
  const { isAuthenticated, isLoading, error } = useAuth()

  if (!isLoading && !isAuthenticated && !error) {
    window.location.href = `${authServiceUrl}/?url=${appUrl}`
  }

  return (
    <>
      {pylonAppId && isAuthenticated && (
        <InitPylonChat PYLON_APP_ID={pylonAppId} />
      )}
      <AuthLayout
        isLoading={isLoading}
        isAuthenticated={!!isAuthenticated}
        hasError={!!error}
        onRetry={() => window.location.reload()}
      />
    </>
  )
}
