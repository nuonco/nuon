import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { InitPylonChat } from '@/lib/pylon-chat'
import { InitPostHog } from '@/lib/posthog'
import { AuthLayout } from './AuthLayout'

export const AuthLayoutContainer = () => {
  const { authServiceUrl, appUrl, pylonAppId, postHogApiKey, postHogHost } =
    useConfig()
  const { isAuthenticated, isLoading, error } = useAuth()

  if (!isLoading && !isAuthenticated && !error) {
    window.location.href = `${authServiceUrl}/?url=${appUrl}`
  }

  return (
    <>
      {postHogApiKey && (
        <InitPostHog apiKey={postHogApiKey} apiHost={postHogHost} />
      )}
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
