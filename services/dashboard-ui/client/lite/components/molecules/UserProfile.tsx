import type { HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'
import { Avatar, type TAvatarSize } from '../atoms/Avatar'
import { Text } from '../atoms/Text'

export interface IUserProfileData {
  name?: string
  email?: string
  picture?: string
}

export interface IUserProfile
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  user?: IUserProfileData | null
  loading?: boolean
  compact?: boolean
  avatarSize?: TAvatarSize
}

export const UserProfile = ({
  user,
  loading = false,
  compact = false,
  avatarSize = 'md',
  className,
  ...props
}: IUserProfile) => {
  const name = user?.name?.trim() || user?.email?.trim() || 'User'
  const email = user?.name?.trim() ? user?.email?.trim() : undefined

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
        src={user?.picture}
        size={avatarSize}
        loading={loading}
      />
      {compact ? null : (
        <span className="flex min-w-0 flex-col">
          <Text
            variant="caption"
            weight="semibold"
            loading={loading}
            loadingWidth={10}
            className="max-w-48 truncate leading-tight"
          >
            {name}
          </Text>
          {loading || email ? (
            <Text
              variant="label"
              color="tertiary"
              loading={loading}
              loadingWidth={16}
              className="max-w-48 truncate leading-tight"
            >
              {email}
            </Text>
          ) : null}
        </span>
      )}
    </div>
  )
}
