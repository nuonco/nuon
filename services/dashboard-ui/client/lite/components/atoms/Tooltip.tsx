import { useId, useState, type HTMLAttributes, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { cn } from '@/utils/classnames'
import { usePopover, type TPopoverSide } from '../../hooks/use-popover'
import { Text } from './Text'

export interface ITooltip
  extends Omit<HTMLAttributes<HTMLSpanElement>, 'content'> {
  content: ReactNode
  side?: TPopoverSide
  open?: boolean
  defaultOpen?: boolean
  onOpenChange?: (open: boolean) => void
  disableHover?: boolean
  contentClassName?: string
}

export const Tooltip = ({
  content,
  side = 'top',
  open: controlledOpen,
  defaultOpen = false,
  onOpenChange,
  disableHover = false,
  contentClassName,
  className,
  children,
  ...props
}: ITooltip) => {
  const [uncontrolledOpen, setUncontrolledOpen] = useState(defaultOpen)
  const isControlled = controlledOpen !== undefined
  const isOpen = isControlled ? controlledOpen : uncontrolledOpen
  const tooltipId = useId()

  const { triggerRef, contentRef, side: placedSide, style } = usePopover<
    HTMLSpanElement,
    HTMLDivElement
  >({ open: isOpen, side })

  const setOpen = (next: boolean) => {
    if (disableHover) return
    if (!isControlled) setUncontrolledOpen(next)
    onOpenChange?.(next)
  }

  return (
    <span
      ref={triggerRef}
      className={cn('inline-flex w-fit', className)}
      aria-describedby={isOpen ? tooltipId : undefined}
      onMouseEnter={() => setOpen(true)}
      onMouseLeave={() => setOpen(false)}
      onFocus={() => setOpen(true)}
      onBlur={() => setOpen(false)}
      {...props}
    >
      {children}
      {createPortal(
        <div
          ref={contentRef}
          id={tooltipId}
          role="tooltip"
          data-state={isOpen ? 'open' : 'closed'}
          data-side={placedSide}
          style={style}
          className={cn(
            'popover popover-arrow max-w-xs rounded-md px-2 py-1 text-caption',
            contentClassName
          )}
        >
          {typeof content === 'string' ? (
            <Text as="p" variant="caption">
              {content}
            </Text>
          ) : (
            content
          )}
        </div>,
        document.body
      )}
    </span>
  )
}
