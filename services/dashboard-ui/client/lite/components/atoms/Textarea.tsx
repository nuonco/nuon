import {
  forwardRef,
  useEffect,
  useImperativeHandle,
  useRef,
  type TextareaHTMLAttributes,
} from 'react'
import { cn } from '@/utils/classnames'
import { FIELD_CONTROL_CLASSES, type TFieldSize } from './Input'

export interface ITextarea
  extends Omit<
    TextareaHTMLAttributes<HTMLTextAreaElement>,
    'size' | 'required'
  > {
  size?: TFieldSize
  autoResize?: boolean
  minRows?: number
  maxRows?: number
  loading?: boolean
}

export const Textarea = forwardRef<HTMLTextAreaElement, ITextarea>(
  (
    {
      size = 'md',
      autoResize = false,
      minRows = 3,
      maxRows = 10,
      loading = false,
      disabled,
      className,
      onInput,
      value,
      defaultValue,
      ...props
    },
    forwardedRef
  ) => {
    const ref = useRef<HTMLTextAreaElement>(null)
    useImperativeHandle(forwardedRef, () => ref.current as HTMLTextAreaElement)

    const resize = () => {
      const element = ref.current
      if (!element || !autoResize) return
      const lineHeight = Number.parseFloat(getComputedStyle(element).lineHeight)
      const verticalPadding =
        Number.parseFloat(getComputedStyle(element).paddingTop) +
        Number.parseFloat(getComputedStyle(element).paddingBottom)
      element.style.height = 'auto'
      element.style.height = `${Math.min(
        Math.max(element.scrollHeight, lineHeight * minRows + verticalPadding),
        lineHeight * maxRows + verticalPadding
      )}px`
    }

    useEffect(resize, [autoResize, maxRows, minRows, value])

    return (
      <textarea
        ref={ref}
        rows={minRows}
        disabled={disabled || loading}
        aria-busy={loading || undefined}
        value={value}
        defaultValue={defaultValue}
        onInput={(event) => {
          resize()
          onInput?.(event)
        }}
        className={cn(
          FIELD_CONTROL_CLASSES,
          size === 'sm' ? 'px-2.5 py-1.5 text-caption' : 'px-3 py-2 text-body',
          autoResize ? 'resize-none overflow-y-auto' : 'resize-y',
          loading &&
            'skeleton resize-none text-transparent placeholder:text-transparent',
          className
        )}
        {...props}
      />
    )
  }
)

Textarea.displayName = 'Textarea'
