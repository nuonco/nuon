'use client'

import classNames from 'classnames'
import React, { type FC, useEffect, useState } from 'react'
import { Check, Copy } from '@phosphor-icons/react'

export const ClickToCopy: FC<React.HTMLAttributes<HTMLSpanElement>> = ({
  className,
  children,
  ...props
}) => {
  const [isCopied, setIsCopied] = useState(false)
  const text = children?.valueOf()?.['props']?.children || children

  useEffect(() => {
    const copyNotice = () => setIsCopied(false)

    if (isCopied) {
      const displayNotice = setTimeout(copyNotice, 5000)

      return () => {
        clearTimeout(displayNotice)
      }
    }
  }, [isCopied])

  return (
    <span
      className={classNames('flex items-center gap-2 cursor-pointer relative', {
        [`${className}`]: Boolean(className),
      })}
      onClick={() => {
        navigator.clipboard.writeText(text)
        setIsCopied(true)
      }}
      title="Click to copy"
      {...props}
    >
      {isCopied ? (
        <span
          className={classNames(
            'bg-dark text-light dark:bg-light dark:text-dark text-sm leading-none px-2 py-1.5 rounded drop-shadow-md max-w-96 absolute z-10 -top-6 right-0'
          )}
        >
          Copied
        </span>
      ) : null}
      {children}
      {isCopied ? <Check /> : <Copy />}
    </span>
  )
}
