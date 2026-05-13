import type { HTMLAttributes } from 'react'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TTheme } from '@/types'
import { cn } from '@/utils/classnames'

export type TBannerTheme = TTheme

interface IBanner extends HTMLAttributes<HTMLDivElement> {
  theme?: TBannerTheme
  onDismiss?: () => void
}

const ICONS: Record<TBannerTheme, React.ReactNode> = {
  warn: <Icon variant="WarningIcon" size="20" />,
  error: <Icon variant="WarningOctagonIcon" size="20" />,
  success: <Icon variant="CheckCircleIcon" size="20" />,
  info: <Icon variant="InfoIcon" size="20" />,
  neutral: <Icon variant="InfoIcon" size="20" />,
  default: <Icon variant="InfoIcon" size="20" />,
  brand: <Icon variant="WarningCircleIcon" size="20" />,
}

const THEME_CLASSES: Record<TBannerTheme, string> = {
  default:
    'bg-white text-cool-grey-800 !border-cool-grey-300 dark:bg-dark-grey-800 dark:!border-cool-grey-600/40 dark:text-cool-grey-500',
  neutral:
    'bg-cool-grey-50 text-cool-grey-800 !border-cool-grey-300 dark:bg-dark-grey-600 dark:!border-cool-grey-600/40 dark:text-cool-grey-400',
  info: 'bg-blue-50 text-blue-800 !border-blue-300 dark:bg-[#0F172A] dark:!border-blue-600/40 dark:text-blue-500',
  warn: 'bg-orange-50 text-orange-800 !border-orange-300 dark:bg-[#2D1E10] dark:!border-orange-600/40 dark:text-orange-500',
  error:
    'bg-red-50 text-red-800 !border-red-300 dark:bg-[#2A0C0D] dark:!border-red-600/40 dark:text-red-500',
  success:
    'bg-green-50 text-green-800 !border-green-300 dark:bg-[#0B1A13] dark:!border-green-600/40 dark:text-green-500',
  brand:
    'bg-primary-50 text-primary-800 !border-primary-300 dark:bg-[#251932] dark:!border-primary-600/40 dark:text-primary-500',
}

export const Banner = ({
  className,
  children,
  theme = 'default',
  onDismiss,
  ...props
}: IBanner) => {
  return (
    <div
      className={cn(
        'flex gap-4 h-fit w-full p-4 border rounded-lg',
        THEME_CLASSES[theme],
        onDismiss && 'group',
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
      {onDismiss && (
        <button
          type="button"
          onClick={onDismiss}
          className="self-start shrink-0 opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
          aria-label="Dismiss"
        >
          <Icon variant="XIcon" size={16} />
        </button>
      )}
    </div>
  )
}
