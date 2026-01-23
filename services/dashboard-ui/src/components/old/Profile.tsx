'use client'

import React, { type FC } from 'react'
import { SignOutIcon } from '@phosphor-icons/react'
import Image from 'next/image'
import { useAuth } from '@/hooks/use-auth'
import { Text } from '@/components/old/Typography'

export const Profile: FC<{ isSidebarOpen?: boolean }> = ({
  isSidebarOpen = true,
}) => {
  const { user, error, isLoading } = useAuth()

  if (isLoading) return <div>Loading...</div>
  if (error) return <div>{error.message}</div>

  return (
    user && (
      <div className="flex gap-4 items-center">
        <Image
          className="rounded-lg"
          height={39}
          width={39}
          src={user.picture as string}
          alt={user.name as string}
        />
        {isSidebarOpen ? (
          <div className="w-full overflow-hidden">
            <Text className="truncate" variant="med-14">
              {user.name}
            </Text>
            <Text className="truncate" variant="reg-12">
              {user.email}
            </Text>
          </div>
        ) : null}
      </div>
    )
  )
}

export const SignOutButton: FC<{ isSidebarOpen?: boolean }> = ({
  isSidebarOpen = true,
}) => {
  const { user, useAuthService, authServiceUrl } = useAuth()
  return (
    user && (
      <span className="flex items-center justify-between w-full gap-2 overflow-hidden">
        <Profile isSidebarOpen={isSidebarOpen} />
        {isSidebarOpen ? (
          <a
            href={
              useAuthService ? `${authServiceUrl}/logout` : '/api/auth/logout'
            }
            className="hover:bg-black/5 dark:hover:bg-white/5 w-[48px] h-[48px] p-1 flex text-sm leading-5 text-left rounded-lg"
            title="Sign out"
          >
            <SignOutIcon className="m-auto" size={16} />
          </a>
        ) : null}
      </span>
    )
  )
}
