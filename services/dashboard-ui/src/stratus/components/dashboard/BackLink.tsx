'use client'

import classNames from 'classnames'
import { useRouter } from 'next/navigation'
import React, { type FC } from 'react'
import { CaretLeft } from '@phosphor-icons/react'
import { Text, type IText } from '@/stratus/components/common'

interface IBackLink extends IText {}

export const BackLink: FC<IBackLink> = ({
  className,
  children = (
    <>
      <CaretLeft weight="bold" /> Back
    </>
  ),
  variant = 'base',
  weight = 'strong',
  ...props
}) => {
  const router = useRouter()

  return (
    <Text
      className={classNames(
        'flex items-center gap-1.5 link default cursor-pointer',
        {
          [`${className}`]: Boolean(className),
        }
      )}
      onClick={() => {
        router.back()
      }}
      variant={variant}
      weight={weight}
      {...props}
    >
      {children}
    </Text>
  )
}
