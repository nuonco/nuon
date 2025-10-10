'use client'

import { useRouter } from 'next/navigation'
import { cn } from '@/utils/classnames'
import { Icon } from './Icon'
import { Text, type IText } from './Text'

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
        '!flex items-center gap-1.5 cursor-pointer w-fit',
        'text-primary-600 dark:text-primary-500',
        'hover:text-primary-800 hover:dark:text-primary-400',
        'focus:text-primary-800 focus:dark:text-primary-400',
        'active:text-primary-900 active:dark:text-primary-600',
        'focus-visible:rounded',
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
