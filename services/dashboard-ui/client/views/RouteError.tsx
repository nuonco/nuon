import { isRouteErrorResponse, useRouteError } from 'react-router'
import { BackLink } from '@/components/common/BackLink'
import { EmptyState } from '@/components/common/EmptyState'

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
            : 'An unexpected error occurred. Please try again.'
        }
        action={<BackLink />}
      />
    </div>
  )
}
