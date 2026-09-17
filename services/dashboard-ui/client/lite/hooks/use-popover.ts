import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  type CSSProperties,
} from 'react'

export type TPopoverSide = 'top' | 'bottom' | 'left' | 'right'
export type TPopoverAlign = 'start' | 'center' | 'end'

export interface IUsePopover {
  open: boolean
  side?: TPopoverSide
  align?: TPopoverAlign
  gap?: number
  margin?: number
  arrowInset?: number
}

const OPPOSITE: Record<TPopoverSide, TPopoverSide> = {
  top: 'bottom',
  bottom: 'top',
  left: 'right',
  right: 'left',
}

const clamp = (value: number, min: number, max: number) =>
  Math.min(Math.max(value, min), Math.max(min, max))

const fits = (
  side: TPopoverSide,
  trigger: DOMRect,
  content: DOMRect,
  gap: number,
  margin: number
) => {
  if (side === 'top') return trigger.top - gap - content.height >= margin
  if (side === 'bottom')
    return trigger.bottom + gap + content.height <= window.innerHeight - margin
  if (side === 'left') return trigger.left - gap - content.width >= margin
  return trigger.right + gap + content.width <= window.innerWidth - margin
}

export const usePopover = <
  TTrigger extends HTMLElement = HTMLElement,
  TContent extends HTMLElement = HTMLElement,
>({
  open,
  side = 'top',
  align = 'center',
  gap = 8,
  margin = 8,
  arrowInset = 12,
}: IUsePopover) => {
  const triggerRef = useRef<TTrigger>(null)
  const contentRef = useRef<TContent>(null)
  const [placedSide, setPlacedSide] = useState<TPopoverSide>(side)
  const [style, setStyle] = useState<CSSProperties>()

  const place = useCallback(() => {
    const triggerEl = triggerRef.current
    const contentEl = contentRef.current
    if (!triggerEl || !contentEl) return

    const trigger = triggerEl.getBoundingClientRect()
    const content = contentEl.getBoundingClientRect()

    const resolved =
      fits(side, trigger, content, gap, margin) ||
      !fits(OPPOSITE[side], trigger, content, gap, margin)
        ? side
        : OPPOSITE[side]

    const isVertical = resolved === 'top' || resolved === 'bottom'

    let top: number
    let left: number

    if (isVertical) {
      top =
        resolved === 'top'
          ? trigger.top - content.height - gap
          : trigger.bottom + gap
      left =
        align === 'start'
          ? trigger.left
          : align === 'end'
            ? trigger.right - content.width
            : trigger.left + trigger.width / 2 - content.width / 2
    } else {
      left =
        resolved === 'left'
          ? trigger.left - content.width - gap
          : trigger.right + gap
      top =
        align === 'start'
          ? trigger.top
          : align === 'end'
            ? trigger.bottom - content.height
            : trigger.top + trigger.height / 2 - content.height / 2
    }

    left = clamp(left, margin, window.innerWidth - content.width - margin)
    top = clamp(top, margin, window.innerHeight - content.height - margin)

    const arrow = isVertical
      ? clamp(
          trigger.left + trigger.width / 2 - left,
          arrowInset,
          content.width - arrowInset
        )
      : clamp(
          trigger.top + trigger.height / 2 - top,
          arrowInset,
          content.height - arrowInset
        )

    const available = isVertical
      ? resolved === 'top'
        ? trigger.top - gap - margin
        : window.innerHeight - trigger.bottom - gap - margin
      : window.innerHeight - margin * 2

    setPlacedSide(resolved)
    setStyle({
      top: `${top}px`,
      left: `${left}px`,
      '--arrow': `${arrow}px`,
      '--available': `${Math.max(0, available)}px`,
    } as CSSProperties)
  }, [side, align, gap, margin, arrowInset])

  useLayoutEffect(() => {
    if (!open) return
    place()
  }, [open, place])

  useEffect(() => {
    if (!open) return

    const trigger = triggerRef.current
    const content = contentRef.current
    const observer = new ResizeObserver(place)
    if (trigger) observer.observe(trigger)
    if (content) observer.observe(content)

    window.addEventListener('resize', place)
    window.addEventListener('scroll', place, true)
    return () => {
      observer.disconnect()
      window.removeEventListener('resize', place)
      window.removeEventListener('scroll', place, true)
    }
  }, [open, place])

  return { triggerRef, contentRef, side: placedSide, style, place }
}
