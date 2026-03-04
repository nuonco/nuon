import { EmptyState } from '@/components/common/EmptyState'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TAPIError } from '@/types'

export const ProviderError = ({ error }: { error: TAPIError }) => {
  const isNotFound = error.status === 404 || error.status === 403
  return (
    <div className="flex items-center justify-center w-full h-screen">
      <EmptyState
        variant={isNotFound ? '404' : 'diagram'}
        emptyTitle={isNotFound ? 'Not found' : 'Something went wrong'}
        emptyMessage={
          isNotFound
            ? "The resource you're looking for doesn't exist."
            : (error.error ?? 'An unexpected error occurred.')
        }
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
