import { Outlet } from 'react-router'
import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'

export const AuthLayout = () => {
  const { authServiceUrl, appUrl } = useConfig()
  const { isAuthenticated, isLoading, error } = useAuth()

  if (isLoading) {
    return null
  }

  if (!isAuthenticated && error) {
    return <APIUnavailable onRetry={() => window.location.reload()} />
  }

  if (!isAuthenticated) {
    window.location.href = `${authServiceUrl}/?url=${appUrl}`
    return null
  }

  return <Outlet />
}

const APIUnavailable = ({ onRetry }: { onRetry: () => void }) => (
  <div className="flex min-h-screen items-center justify-center">
    <div className="flex flex-col items-center gap-16">
      <div className="flex flex-col items-center gap-4">
        <Text variant="heading-2">Unable to connect</Text>
        <Text variant="subtext" theme="neutral">
          The API is currently unavailable. Please try again in a moment.
        </Text>
      </div>
      <Button onClick={onRetry}>Retry</Button>
    </div>
  </div>
)

