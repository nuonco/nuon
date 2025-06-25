'use client'

import { useRouter } from 'next/navigation'
import React from 'react'
import { Icon, Text, type IText } from '@/stratus/components/common'
import { cn } from '@/stratus/components/helpers'

interface IBackLink extends IText {}

export const BackLink = ({
  className,
  children = (
    <>
      <Icon variant="CaretLeft" weight="bold" /> Back
    </>
  ),
  variant = 'base',
  weight = 'strong',
  ...props
}: IBackLink) => {
  const router = useRouter()

  return (
    <Text
      className={cn(
        'flex items-center gap-1.5 link default cursor-pointer',
        className
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
