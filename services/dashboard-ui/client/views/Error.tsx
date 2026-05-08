import { useSearchParams } from 'react-router'
import { EmptyState } from '@/components/common/EmptyState'
import { Button } from '@/components/common/Button'

const messages: Record<string, { title: string; message: string }> = {
  'orgs-failed': {
    title: 'Unable to load your organizations',
    message:
      'We had trouble connecting to our servers. This is usually temporary.',
  },
  'api-error': {
    title: 'Unable to connect to the API',
    message: 'Something went wrong on our end. This is usually temporary.',
  },
}

const fallback = {
  title: 'Something went wrong',
  message: 'An unexpected error occurred. Please try again.',
}

export const Error = () => {
  const [params] = useSearchParams()
  const reason = params.get('reason') ?? ''
  const { title, message } = messages[reason] ?? fallback

  return (
    <div className="flex flex-col flex-1 items-center justify-center h-full">
      <EmptyState
        variant="404"
        emptyTitle={title}
        emptyMessage={message}
        action={
          <Button
            variant="secondary"
            size="sm"
            onClick={() => {
              window.location.href = '/'
            }}
          >
            Try again
          </Button>
        }
      />
    </div>
  )
}
