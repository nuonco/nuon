import { BackLink } from '@/components/common/BackLink'
import { EmptyState } from '@/components/common/EmptyState'

export const NotFound = () => {
  return (
    <div className="flex flex-col flex-1 items-center justify-center h-full">
      <EmptyState
        variant="404"
        emptyTitle="Page not found test"
        emptyMessage="The page you're looking for doesn't exist or has been moved."
        action={<BackLink />}
      />
    </div>
  )
}
