'use client'

import React, { type FC } from 'react'
import { SignOut } from '@phosphor-icons/react'
import Image from 'next/image'
import { useUser } from '@auth0/nextjs-auth0'
import { Text } from '@/components/old/Typography'

export const Profile: FC<{ isSidebarOpen?: boolean }> = ({
  isSidebarOpen = true,
}) => {
  const { user, error, isLoading } = useUser()

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
  const { user } = useUser()
  return (
    user && (
      <span className="flex items-center justify-between w-full gap-2 overflow-hidden">
        <Profile isSidebarOpen={isSidebarOpen} />
        {isSidebarOpen ? (
          <a
            href="/api/auth/logout"
            className="hover:bg-black/5 dark:hover:bg-white/5 w-[48px] h-[48px] p-1 flex text-sm leading-5 text-left rounded-lg"
            title="Sign out"
          >
            <SignOut className="m-auto" size={16} />
          </a>
        ) : null}
      </span>
    )
  )
}
