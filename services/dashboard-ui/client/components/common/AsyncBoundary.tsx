import { ReactNode, Suspense } from 'react'
import { ErrorBoundary } from './ErrorBoundary'
import { Skeleton } from './Skeleton'
import { EmptyState } from './EmptyState'

type TFallbackComponent = ReactNode | ((props: { error: Error }) => ReactNode)

interface IAsyncBoundary {
  children: ReactNode
  errorFallback?: TFallbackComponent
  loadingFallback?: ReactNode
}

const DefaultErrorFallback = () => (
  <EmptyState
    emptyTitle="Something went wrong"
    emptyMessage="An error occurred while loading this content."
    variant="table"
  />
)

const DefaultSkeletonFallback = () => <Skeleton height="80px" />

export function AsyncBoundary({
  children,
  errorFallback = <DefaultErrorFallback />,
  loadingFallback = <DefaultSkeletonFallback />,
}: IAsyncBoundary) {
  return (
    <ErrorBoundary fallback={errorFallback}>
      <Suspense fallback={loadingFallback}>{children}</Suspense>
    </ErrorBoundary>
  )
}
