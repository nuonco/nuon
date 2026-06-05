import { Outlet } from 'react-router'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'

interface IAuthLayout {
  isLoading: boolean
  isAuthenticated: boolean
  hasError: boolean
  onRetry: () => void
}

export const AuthLayout = ({
  isLoading,
  isAuthenticated,
  hasError,
  onRetry,
}: IAuthLayout) => {
  if (isLoading) return null

  if (!isAuthenticated && hasError) {
    return <APIUnavailable onRetry={onRetry} />
  }

  if (!isAuthenticated) return null

  return <Outlet />
}

const APIUnavailable = ({ onRetry }: { onRetry: () => void }) => (
  <div className="flex min-h-screen items-center justify-center">
    <div className="flex flex-col items-center gap-16">
      <div className="flex flex-col items-center gap-4">
        <Text variant="h2">Unable to connect</Text>
        <Text variant="subtext" theme="neutral">
          The API is currently unavailable. Try again in a moment.
        </Text>
      </div>
      <Button onClick={onRetry}>Retry</Button>
    </div>
  </div>
)
