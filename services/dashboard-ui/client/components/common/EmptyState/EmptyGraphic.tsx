import type { TEmptyVariant } from '@/types/dashboard.types'
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
  const dims = size === 'sm' ? { w: 94, h: 64 } : { w: 154, h: 94 }
  const sizeClass = size === 'sm' ? 'w-[94px] h-[64px]' : 'w-[154px] h-[94px]'
  const variants = {
    light: `/empty-graphics/${variant}-light${sizeSuffix}.svg`,
    dark: `/empty-graphics/${variant}-dark${sizeSuffix}.svg`,
  }

  return (
    <>
      <img
        className={cn(sizeClass, 'relative block', {
          hidden: isDarkModeOnly,
          'dark:hidden': !isDarkModeOnly,
        })}
        src={variants.light}
        alt=""
        width={dims.w}
        height={dims.h}
        draggable={false}
      />
      <img
        className={cn(sizeClass, 'relative dark:block', {
          block: isDarkModeOnly,
          hidden: !isDarkModeOnly,
        })}
        src={variants.dark}
        alt=""
        width={dims.w}
        height={dims.h}
        draggable={false}
      />
    </>
  )
}
