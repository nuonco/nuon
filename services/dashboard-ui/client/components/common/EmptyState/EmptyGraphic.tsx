import Image from 'next/image'
import type { TEmptyVariant } from '@/types'
import { cn } from '@/utils/classnames'

interface IEmptyGraphic {
  isDarkModeOnly?: boolean
  size?: 'default' | 'sm'
  variant?: TEmptyVariant
}

export const EmptyGraphic = ({
  isDarkModeOnly = false,
  size = 'default',
  variant = '404',
}: IEmptyGraphic) => {
  const sizeSuffix = size === 'sm' ? '-small' : ''
  const variants = {
    light: `/empty-graphics/${variant}-light${sizeSuffix}.svg`,
    dark: `/empty-graphics/${variant}-dark${sizeSuffix}.svg`,
  }

  return (
    <>
      <Image
        className={cn('w-auto relative block', {
          hidden: isDarkModeOnly,
          'dark:hidden': !isDarkModeOnly,
        })}
        src={variants.light}
        alt=""
        height={90}
        width={150}
        draggable={false}
        priority
      />
      <Image
        className={cn('w-auto relative dark:block', {
          block: isDarkModeOnly,
          hidden: !isDarkModeOnly,
        })}
        src={variants.dark}
        alt=""
        height={90}
        width={150}
        draggable={false}
        priority
      />
    </>
  )
}
