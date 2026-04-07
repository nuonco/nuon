import { Avatar } from '@/components/common/Avatar'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'

interface IUserProfile {
  isLoading: boolean
  user?: {
    picture?: string
    name?: string
    email?: string
  } | null
}

export const UserProfile = ({ isLoading, user }: IUserProfile) => {
  return (
    <div className="flex gap-4 items-center min-w-40">
      {isLoading ? (
        <>
          <Avatar isLoading />
          <div className="flex flex-col gap-0.5 w-full overflow-hidden">
            <Skeleton height="14px" />
            <Skeleton height="11px" width="75%" />
          </div>
        </>
      ) : (
        user && (
          <>
            <Avatar src={user?.picture} name={user?.name} alt={user?.name} />
            <div className="flex flex-col gap-0.5 w-full overflow-hidden">
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
