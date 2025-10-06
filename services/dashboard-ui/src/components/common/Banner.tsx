import React from 'react'
import { Icon } from './Icon'
import { Text } from './Text'
import { cn } from '@/utils/classnames'

export type TBannerTheme = 'error' | 'warn' | 'info' | 'success' | 'default'

interface IBanner extends React.HTMLAttributes<HTMLDivElement> {
  theme?: TBannerTheme
}

const ICONS: Record<TBannerTheme, React.ReactNode> = {
  warn: <Icon variant="Warning" size="20" />,
  error: <Icon variant="WarningOctagon" size="20" />,
  success: <Icon variant="CheckCircle" size="20" />,
  info: <Icon variant="Info" size="20" />,
  default: <Icon variant="Info" size="20" />,
}

const THEME_CLASSES: Record<TBannerTheme, string> = {
  default:
    'bg-cool-grey-50 text-cool-grey-800 !border-cool-grey-300 dark:bg-dark-grey-800 dark:!border-cool-grey-600/40 dark:text-cool-grey-500',
  info: 'bg-blue-50 text-blue-800 !border-blue-300 dark:bg-[#0F172A] dark:!border-blue-600/40 dark:text-blue-500',
  warn: 'bg-orange-50 text-orange-800 !border-orange-300 dark:bg-[#2D1E10] dark:!border-orange-600/40 dark:text-orange-500',
  error:
    'bg-red-50 text-red-800 !border-red-300 dark:bg-[#2A0C0D] dark:!border-red-600/40 dark:text-red-500',
  success:
    'bg-green-50 text-green-800 !border-green-300 dark:bg-[#0B1A13] dark:!border-green-600/40 dark:text-green-500',
}

export const Banner = ({
  className,
  children,
  theme = 'default',
  ...props
}: IBanner) => {
  return (
    <div
      className={cn(
        'flex gap-4 w-full p-4 border rounded-lg',
        THEME_CLASSES[theme],
        className
      )}
      {...props}
    >
      <div className="flex mt-0.5 self-start">{ICONS[theme]}</div>
      <div className="!w-full">
        {typeof children === 'string' ? (
          <Text weight="strong">{children}</Text>
        ) : (
          children
        )}
      </div>
    </div>
  )
}
