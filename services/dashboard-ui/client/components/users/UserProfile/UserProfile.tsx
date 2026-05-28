import { Avatar } from '@/components/common/Avatar'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

interface IUserProfile {
  collapsible?: boolean
  isCollapsed?: boolean
  isLoading: boolean
  user?: {
    picture?: string
    name?: string
    email?: string
  } | null
}

export const UserProfile = ({
  collapsible,
  isCollapsed,
  isLoading,
  user,
}: IUserProfile) => {
  return (
    <div className="flex items-center w-fit overflow-hidden">
      {isLoading ? (
        <>
          <Avatar size={collapsible ? 'sidebar' : 'md'} isLoading />
          <div
            className={cn(
              'flex flex-col gap-0.5 w-full overflow-hidden ml-2',
              collapsible && 'transition-all duration-fast',
              collapsible && isCollapsed && 'md:opacity-0 md:w-0 md:!ml-0',
              collapsible && !isCollapsed && 'md:opacity-100'
            )}
          >
            <Skeleton height="14px" />
            <Skeleton height="11px" width="75%" />
          </div>
        </>
      ) : (
        user && (
          <>
            <Avatar
              size={collapsible ? 'sidebar' : 'md'}
              src={user?.picture}
              name={user?.name}
              alt={user?.name}
            />
            <div
              className={cn(
                'flex flex-col gap-0.5 w-full overflow-hidden ml-2',
                collapsible && 'transition-all duration-fast whitespace-nowrap',
                collapsible && isCollapsed && 'md:opacity-0 md:w-0 md:!ml-0',
                collapsible && !isCollapsed && 'md:opacity-100'
              )}
            >
              <Text className="!leading-none" variant="body" weight="strong">
                {user?.name}
              </Text>
              <Text variant="label">{user?.email}</Text>
            </div>
          </>
        )
      )}
    </div>
  )
}
