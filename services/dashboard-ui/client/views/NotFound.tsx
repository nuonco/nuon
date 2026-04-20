import { EmptyState } from '@/components/common/EmptyState'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'

export const NotFound = () => {
  return (
    <div className="flex flex-col flex-1 items-center justify-center h-full">
      <EmptyState
        variant="404"
        emptyTitle="Page not found"
        emptyMessage="The page you're looking for doesn't exist or has been moved."
        action={
          <Text variant="subtext">
            <Link href="/" isATag>
              Back to home
            </Link>
          </Text>
        }
      />
    </div>
  )
}
