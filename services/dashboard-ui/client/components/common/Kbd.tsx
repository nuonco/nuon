import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

export type TKbdSize = 'sm' | 'md'

interface IKbd {
  children: React.ReactNode
  size?: TKbdSize
  className?: string
}

const SIZE_CLASSES: Record<TKbdSize, string> = {
  sm: 'min-w-[16px] h-[16px] px-1 text-[10px]',
  md: 'min-w-[20px] h-[20px] px-1.5 text-[11px]',
}

export const Kbd = ({ children, size = 'md', className }: IKbd) => (
  <kbd
    className={cn(
      'inline-flex items-center justify-center rounded border font-medium leading-none',
      'border-cool-grey-200 dark:border-dark-grey-600',
      'bg-cool-grey-50 dark:bg-dark-grey-800',
      'text-cool-grey-700 dark:text-cool-grey-300',
      SIZE_CLASSES[size],
      className
    )}
  >
    {children}
  </kbd>
)

export interface IKbdShortcut {
  shortcut: string
  size?: TKbdSize
  separator?: 'plus' | 'then' | 'none'
  className?: string
}

const SEPARATOR_TEXT: Record<'plus' | 'then', string> = {
  plus: '+',
  then: 'then',
}

export const KbdShortcut = ({
  shortcut,
  size = 'md',
  separator = 'none',
  className,
}: IKbdShortcut) => {
  const parts = shortcut.split(' ').filter(Boolean)
  return (
    <span className={cn('inline-flex items-center gap-1', className)}>
      {parts.map((part, i) => (
        <span key={i} className="inline-flex items-center gap-1">
          <Kbd size={size}>{part.toUpperCase()}</Kbd>
          {separator !== 'none' && i < parts.length - 1 ? (
            <Text
              variant="subtext"
              theme="neutral"
              className={cn(
                'leading-none',
                size === 'sm' ? 'text-[10px]' : 'text-xs'
              )}
            >
              {SEPARATOR_TEXT[separator]}
            </Text>
          ) : null}
        </span>
      ))}
    </span>
  )
}
