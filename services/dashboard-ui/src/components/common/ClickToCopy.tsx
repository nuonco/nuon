'use client'

import React, { useEffect, useState, useCallback } from 'react'
import { Icon } from './Icon'
import { cn } from '@/utils/classnames'

export interface IClickToCopy extends React.HTMLAttributes<HTMLSpanElement> {
  noticeClassName?: string
}

type TUseCopyStateResult = {
  isCopied: boolean
  handleCopy: (text: string) => void
}

const useCopyState = (): TUseCopyStateResult => {
  const [isCopied, setIsCopied] = useState(false)

  useEffect(() => {
    if (!isCopied) return
    const timeout = setTimeout(() => setIsCopied(false), 5000)
    return () => clearTimeout(timeout)
  }, [isCopied])

  const handleCopy = useCallback((text: string) => {
    if (!text) return
    navigator.clipboard.writeText(text)
    setIsCopied(true)
  }, [])

  return { isCopied, handleCopy }
}

function CopiedNotice({ className }: { className?: string }) {
  return (
    <span
      className={cn(
        'bg-foreground text-background',
        'text-xs leading-none px-2 py-1.5 rounded shadow-lg max-w-96 absolute z-10 -top-6 right-0',
        className
      )}
    >
      Copied
    </span>
  )
}

export function ClickToCopy({
  className,
  children,
  noticeClassName,
  ...props
}: IClickToCopy) {
  // Try to infer text to copy from children, fallback to rendering children as is.
  let text = ''
  if (typeof children === 'string' || typeof children === 'number') {
    text = String(children)
  } else if (
    React.isValidElement(children) &&
    typeof (children.props as any)?.children === 'string'
  ) {
    text = (children.props as any).children
  }

  const { isCopied, handleCopy } = useCopyState()

  return (
    <span
      className={cn(
        'flex items-center gap-2 cursor-pointer relative w-fit',
        className
      )}
      onClick={() => handleCopy(text)}
      title="Click to copy"
      {...props}
    >
      {isCopied && <CopiedNotice className={noticeClassName} />}
      {children}
      <span>
        {isCopied ? <Icon variant="Check" /> : <Icon variant="Copy" />}
      </span>
    </span>
  )
}

export interface IClickToCopyButton extends Omit<IClickToCopy, 'children'> {
  textToCopy: string
}

export function ClickToCopyButton({
  className,
  noticeClassName,
  textToCopy,
  ...props
}: IClickToCopyButton) {
  const { isCopied, handleCopy } = useCopyState()

  return (
    <span
      className={cn(
        'hover:bg-black/10 dark:hover:bg-white/5',
        'flex items-center gap-2 cursor-pointer relative border rounded-md p-1 text-sm',
        className
      )}
      onClick={() => handleCopy(textToCopy)}
      title="Click to copy"
      {...props}
    >
      {isCopied && <CopiedNotice className={noticeClassName} />}
      <span>
        {isCopied ? <Icon variant="Check" /> : <Icon variant="Copy" />}
      </span>
    </span>
  )
}
