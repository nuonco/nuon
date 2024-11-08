'use client'

import React, { type FC } from 'react'
import { SignOut } from "@phosphor-icons/react"
import Image from 'next/image'
import { useUser } from '@auth0/nextjs-auth0/client'
import { Text } from '@/components/Typography'

export const Profile: FC = () => {
  const { user, error, isLoading } = useUser()

  if (isLoading) return <div>Loading...</div>
  if (error) return <div>{error.message}</div>

  return (
    user && (
      <div className="flex gap-4 items-center">
        <Image
          className="rounded-lg"
          height={40}
          width={40}
          src={user.picture as string}
          alt={user.name as string}
        />
        <div className="flex flex-col gap-0">
          <Text className="truncate" variant="med-14">
            {user.name}
          </Text>
          <Text className="truncate" variant="reg-12">
            {user.email}
          </Text>
        </div>
      </div>
    )
  )
}

export const SignOutButton: FC = () => {
  const { user } = useUser()
  return user && (
    <a
      href="/api/auth/logout"
      className="hover:bg-black/5 dark:hover:bg-white/5 h-[48px] p-1 flex items-center justify-between w-full text-sm leading-5 text-left gap-2 rounded-lg"
    >
      <Profile />
      <SignOut size={16} />
    </a>
  )
}
