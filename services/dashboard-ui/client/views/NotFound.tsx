import { EmptyState } from '@/components/common/EmptyState'
import { Link } from '@/components/common/Link'
import { PageTitle } from '@/components/navigation/PageTitle'

export const NotFound = () => {
  return (
    <div className="flex flex-col flex-1 items-center justify-center h-full">
      <PageTitle title="Page not found" />
      <EmptyState
        variant="404"
        emptyTitle="Page not found"
        emptyMessage="The page you're looking for doesn't exist or has been moved."
        action={
          <Link href="/" isATag>
            Back to home
          </Link>
        }
      />
    </div>
  )
}
