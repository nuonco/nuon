import type { HTMLAttributes } from 'react'
import type { TOrg } from '@/types'
import { cn } from '@/utils/classnames'
import { Avatar, type TAvatarSize } from '../atoms/Avatar'
import { Status } from '../atoms/Status'
import { Text } from '../atoms/Text'
import { ID } from './ID'

export type TOrgProfileVariant = 'full' | 'modeline'

export interface IOrgProfile
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  org?: TOrg | null
  loading?: boolean
  avatarSize?: TAvatarSize
  variant?: TOrgProfileVariant
}

export const OrgProfile = ({
  org,
  loading = false,
  avatarSize = 'md',
  variant = 'full',
  className,
  ...props
}: IOrgProfile) => {
  const name = org?.name?.trim() || 'Organization unavailable'
  const status = org?.status_v2?.status ?? org?.status

  const isModeline = variant === 'modeline'

  const identity = (
    <span className="flex min-w-0 items-center gap-1.5">
      <Status status={status} variant="dot" loading={loading} tabIndex={-1} />
      <Text
        variant="caption"
        family={isModeline ? 'mono' : 'sans'}
        weight={isModeline ? 'medium' : 'semibold'}
        loading={loading}
        loadingWidth={12}
        className="max-w-48 truncate leading-tight"
      >
        {name}
      </Text>
    </span>
  )

  if (isModeline) {
    return (
      <div
        className={cn('inline-flex min-w-0 items-center', className)}
        {...props}
      >
        {identity}
      </div>
    )
  }

  return (
    <div
      className={cn(
        'inline-flex min-w-0 items-center gap-2 text-left',
        className
      )}
      {...props}
    >
      <Avatar
        name={name}
        src={org?.logo_url}
        size={avatarSize}
        shape="rounded"
        loading={loading}
      />
      <span className="flex min-w-0 flex-col">
        {identity}
        <ID
          value={org?.id ?? '—'}
          label="Organization ID"
          truncate
          copyable={false}
          loading={loading}
          loadingWidth={14}
        />
      </span>
    </div>
  )
}
