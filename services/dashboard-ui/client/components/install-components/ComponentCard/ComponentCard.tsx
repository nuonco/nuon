import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import type { TComponentType } from '@/types'

export interface IComponentCard {
  name?: string
  type?: TComponentType
  status?: string
  href?: string
  isLoading?: boolean
  error?: string
}

export const ComponentCard = ({
  name,
  type,
  status,
  href,
  isLoading,
  error,
}: IComponentCard) => {
  if (isLoading) {
    return (
      <div className="flex w-fit items-center gap-3 rounded-lg border px-3 py-2.5">
        <Skeleton width="6rem" />
        <Skeleton width="4rem" />
        <Skeleton width="3.5rem" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex w-fit items-center gap-2 rounded-lg border !border-red-200 dark:!border-red-900 px-3 py-2.5">
        <Text variant="subtext" theme="error">
          {error}
        </Text>
      </div>
    )
  }

  const content = (
    <div className="flex w-fit items-center gap-3 rounded-lg border px-3 py-2.5">
      {name && (
        <Text variant="body" className="font-strong">
          {name}
        </Text>
      )}
      {type && (
        <ComponentType type={type} variant="subtext" colorVariant="color" />
      )}
      {status && <Status status={status} variant="badge" />}
    </div>
  )

  if (href) {
    return (
      <Link href={href} variant="ghost" className="flex !p-0 no-underline">
        {content}
      </Link>
    )
  }

  return content
}
