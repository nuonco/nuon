import { isRouteErrorResponse, useRouteError } from 'react-router'
import { EmptyState } from '@/components/common/EmptyState'
import { Link } from '@/components/common/Link'

export const RouteError = () => {
  const error = useRouteError()
  const is404 = isRouteErrorResponse(error) && error.status === 404

  return (
    <div className="flex flex-col flex-1 items-center justify-center h-full">
      <EmptyState
        variant="404"
        emptyTitle={is404 ? 'Page not found' : 'Something went wrong'}
        emptyMessage={
          is404
            ? "The page you're looking for doesn't exist or has been moved."
            : 'An unexpected error occurred. Try again.'
        }
        action={
          <Link href="/" isATag>
            Back to home
          </Link>
        }
      />
    </div>
  )
}
