'use client'

import { redirect } from "next/navigation"
import { useUser } from '@auth0/nextjs-auth0'
import { Avatar } from '@/components/common/Avatar'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'

export const UserProfile = () => {
  const { user, error, isLoading } = useUser()
  if (error) {
    redirect("/")  
  }

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
            <Avatar src={user?.picture} alt={user.name} />
            <div className="flex flex-col gap-0.5 w-full overflow-hidden">
              <Text className="!leading-none" variant="body" weight="strong">
                {user.name}
              </Text>
              <Text variant="label">{user.email}</Text>
            </div>
          </>
        )
      )}
    </div>
  )
}
