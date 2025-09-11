'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import Image from 'next/image'
import { initialsFromString } from '@/utils'

export const OrgAvatar: FC<{
  name: string
  isSmall?: boolean
  logoURL?: string
}> = ({ name, isSmall = false, logoURL }) => {
  return (
    <span
      className={classNames(
        'flex items-center justify-center rounded-md bg-cool-grey-200 text-cool-grey-600 dark:bg-dark-grey-300 dark:text-white/50 font-medium font-sans',
        {
          'w-[40px] h-[40px]': !isSmall,
          'w-[30px] h-[30px]': isSmall,
          'p-2': !logoURL,
        }
      )}
    >
      {logoURL ? (
        <Image
          className="rounded-md"
          height={isSmall ? 30 : 40}
          width={isSmall ? 30 : 40}
          src={logoURL}
          alt="Logo"
        />
      ) : (
        initialsFromString(name)
      )}
    </span>
  )
}
