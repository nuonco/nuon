import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import type { TComponentType } from '@/types'
import { cn } from '@/utils/classnames'

export interface IComponentCard {
  name?: string
  type?: TComponentType
  status?: string
  href?: string
  isLoading?: boolean
  error?: string
}

const Skeleton = () => (
  <div className="flex items-center gap-3 rounded-lg border border-cool-grey-200 dark:border-cool-grey-800 px-3 py-2.5 animate-pulse">
    <div className="h-4 w-24 rounded bg-cool-grey-200 dark:bg-cool-grey-800" />
    <div className="h-4 w-16 rounded bg-cool-grey-200 dark:bg-cool-grey-800" />
    <div className="h-4 w-14 rounded bg-cool-grey-200 dark:bg-cool-grey-800" />
  </div>
)

export const ComponentCard = ({
  name,
  type,
  status,
  href,
  isLoading,
  error,
}: IComponentCard) => {
  if (isLoading) return <Skeleton />

  if (error) {
    return (
      <div className="flex items-center gap-2 rounded-lg border border-red-200 dark:border-red-900 px-3 py-2.5">
        <Text variant="subtext" className="text-red-600 dark:text-red-400">
          {error}
        </Text>
      </div>
    )
  }

  const content = (
    <div
      className={cn(
        'flex items-center gap-3 rounded-lg border px-3 py-2.5',
        'border-cool-grey-200 dark:border-cool-grey-800',
        href && 'hover:border-cool-grey-400 dark:hover:border-cool-grey-600 transition-colors'
      )}
    >
      {name && (
        <Text variant="body" className="font-strong">
          {name}
        </Text>
      )}
      {type && <ComponentType type={type} variant="subtext" colorVariant="color" />}
      {status && <Status status={status} variant="badge" />}
    </div>
  )

  if (href) {
    return <Link href={href} variant="ghost" className="!p-0 no-underline">{content}</Link>
  }

  return content
}
