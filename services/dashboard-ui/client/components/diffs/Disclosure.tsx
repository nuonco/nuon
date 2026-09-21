import {
  useEffect,
  useState,
  type HTMLAttributes,
  type ReactNode,
  type TransitionEvent,
} from 'react'
import { cn } from '@/utils/classnames'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { useDisclosure, type IUseDisclosure } from './use-disclosure'

export interface IDisclosure
  extends IUseDisclosure,
    Omit<HTMLAttributes<HTMLDivElement>, 'id' | 'onChange' | 'title'> {
  title: ReactNode
  description?: ReactNode
  status?: ReactNode
  actions?: ReactNode
  icon?: ReactNode
  headerClassName?: string
  contentClassName?: string
  children: ReactNode
}

const HEADER_CLASSES =
  'flex min-w-0 flex-1 cursor-pointer items-center gap-2 rounded-md px-2 py-2 text-left ' +
  'text-cool-grey-800 dark:text-white/70 outline-none transition-colors ' +
  'hover:bg-cool-grey-500/8 hover:text-foreground active:bg-cool-grey-500/16 ' +
  'focus-visible:outline-1 focus-visible:-outline-offset-2 focus-visible:outline-primary-400/80'

export const Disclosure = ({
  title,
  description,
  status,
  actions,
  icon,
  id,
  open: controlledOpen,
  defaultOpen,
  onOpenChange,
  className,
  headerClassName,
  contentClassName,
  children,
  ...props
}: IDisclosure) => {
  const { open, triggerProps, contentProps } = useDisclosure({
    id,
    open: controlledOpen,
    defaultOpen,
    onOpenChange,
  })

  const [mounted, setMounted] = useState(open)

  useEffect(() => {
    if (open) setMounted(true)
  }, [open])

  const onTransitionEnd = (event: TransitionEvent<HTMLDivElement>) => {
    if (event.propertyName !== 'grid-template-rows') return
    if (!open) setMounted(false)
  }

  return (
    <div className={cn('flex flex-col', className)} {...props}>
      <div className="flex items-center gap-1">
        <button
          type="button"
          className={cn(HEADER_CLASSES, headerClassName)}
          {...triggerProps}
        >
          <Icon
            variant="CaretRightIcon"
            size={16}
            aria-hidden
            className={cn(
              'shrink-0 transition-transform duration-200 ease-out motion-reduce:duration-[1ms]',
              open && 'rotate-90'
            )}
          />
          {icon ? (
            <span className="flex shrink-0 items-center">{icon}</span>
          ) : null}
          <span className="flex min-w-0 flex-1 flex-col">
            <Text weight="strong" className="truncate">
              {title}
            </Text>
            {description ? (
              <Text variant="subtext" theme="neutral" className="truncate">
                {description}
              </Text>
            ) : null}
          </span>
          {status ? <span className="shrink-0">{status}</span> : null}
        </button>
        {actions ? (
          <div className="flex shrink-0 items-center gap-1">{actions}</div>
        ) : null}
      </div>

      <div
        onTransitionEnd={onTransitionEnd}
        className={cn(
          'grid transition-[grid-template-rows] duration-200 ease-out motion-reduce:duration-[1ms]',
          open ? 'grid-rows-[1fr]' : 'grid-rows-[0fr]'
        )}
      >
        <div className="overflow-hidden">
          {mounted ? (
            <div className={contentClassName} {...contentProps}>
              {children}
            </div>
          ) : null}
        </div>
      </div>
    </div>
  )
}
