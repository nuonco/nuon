import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { useConsent } from '@/hooks/use-consent'
import { InitPylonChat } from '@/lib/pylon-chat'
import { InitPostHog } from '@/lib/posthog-analytics'
import { ConsentProvider } from '@/providers/consent-provider'
import { ToastProvider } from '@/providers/toast-provider'
import { ConsentToastContainer } from '@/components/consent/ConsentToast'
import { AuthLayout } from './AuthLayout'

const AnalyticsWithConsent = ({ apiKey }: { apiKey: string }) => {
  const { consent } = useConsent()

  if (consent !== 'granted') return null

  return <InitPostHog apiKey={apiKey} />
}

export const AuthLayoutContainer = () => {
  const { authServiceUrl, appUrl, pylonAppId, posthogKey } = useConfig()
  const { isAuthenticated, isLoading, error } = useAuth()

  if (!isLoading && !isAuthenticated && !error) {
    window.location.href = `${authServiceUrl}/?url=${appUrl}`
  }

  return (
    <ConsentProvider>
      {pylonAppId && isAuthenticated && (
        <InitPylonChat PYLON_APP_ID={pylonAppId} />
      )}
      {posthogKey && isAuthenticated && (
        <AnalyticsWithConsent apiKey={posthogKey} />
      )}
      {posthogKey && isAuthenticated && (
        <ToastProvider>
          <ConsentToastContainer />
        </ToastProvider>
      )}
      <AuthLayout
        isLoading={isLoading}
        isAuthenticated={!!isAuthenticated}
        hasError={!!error}
        onRetry={() => window.location.reload()}
      />
    </ConsentProvider>
  )
}
