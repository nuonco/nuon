import type { HTMLAttributes } from 'react'
import { Icon } from '@/components/common/Icon'
import { cn } from '@/utils/classnames'
import { useShell } from './shell-context'

export type TIconVariant = Parameters<typeof Icon>[0]['variant']

export interface IBlock extends HTMLAttributes<HTMLDivElement> {
  text?: string
  icon?: TIconVariant
  iconSize?: number
  collapsed?: boolean
}

const textSize = (className?: string) => {
  const match = className?.match(/h-\[(\d+)px\]/)
  const height = match ? Number(match[1]) : 12

  if (height <= 8) return 'text-[9px]'
  if (height <= 10) return 'text-[10px]'
  if (height <= 14) return 'text-xs'
  if (height <= 20) return 'text-sm'
  return 'text-base'
}

const barHeight = (className?: string) => {
  const match = className?.match(/h-\[(\d+)px\]/)
  return Math.min(match ? Number(match[1]) : 10, 12)
}

const barWidth = (text: string) => Math.round(text.length * 6.5)

const withoutSizing = (className?: string) =>
  (className ?? '')
    .split(/\s+/)
    .filter(
      (token) =>
        token && !/^h-\[|^w-\[|^max-w-full$|^rounded|^flex-none$/.test(token)
    )
    .join(' ')

export const Block = ({
  className,
  text,
  icon,
  iconSize = 16,
  collapsed = false,
  style,
  ...props
}: IBlock) => {
  const { showText } = useShell()

  if (icon) {
    return (
      <span
        className={cn(
          'inline-flex items-center font-mono whitespace-nowrap',
          textSize(className),
          withoutSizing(className)
        )}
        {...props}
      >
        <Icon variant={icon} size={iconSize} className="flex-none" />
        {text &&
          (showText ? (
            <span
              className={cn(
                'inline-block overflow-hidden transition-all duration-200 ease-out',
                collapsed
                  ? 'ml-0 max-w-0 opacity-0'
                  : 'ml-1.5 max-w-[12rem] opacity-100'
              )}
            >
              {text}
            </span>
          ) : (
            <span
              className={cn(
                'inline-block rounded-sm bg-cool-grey-400 dark:bg-dark-grey-400 transition-all duration-200 ease-out',
                collapsed ? 'ml-0 opacity-0' : 'ml-1.5 opacity-100'
              )}
              style={{
                height: barHeight(className),
                width: collapsed ? 0 : barWidth(text),
              }}
            />
          ))}
      </span>
    )
  }

  if (showText && text) {
    return (
      <span
        className={cn(
          'font-mono whitespace-nowrap',
          textSize(className),
          withoutSizing(className)
        )}
        {...props}
      >
        {text}
      </span>
    )
  }

  return (
    <div
      className={cn(
        'rounded-sm bg-cool-grey-400 dark:bg-dark-grey-400 block',
        className
      )}
      style={style}
      {...props}
    />
  )
}
