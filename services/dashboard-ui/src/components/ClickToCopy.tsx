'use client'

import classNames from 'classnames'
import React, { type FC, useEffect, useState } from 'react'
import { Check, Copy } from '@phosphor-icons/react'

interface IClickToCopy extends React.HTMLAttributes<HTMLSpanElement> {
  noticeClassName?: string
}

export const ClickToCopy: FC<IClickToCopy> = ({
  className,
  children,
  noticeClassName,
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
            'bg-dark text-light dark:bg-light dark:text-dark text-sm leading-none px-2 py-1.5 rounded drop-shadow-md max-w-96 absolute z-10 -top-6 right-0',
            {             
              [`${noticeClassName}`]: Boolean(noticeClassName),
            }
          )}
        >
          Copied
        </span>
      ) : null}
      {children}
      <span>
        {isCopied ? <Check /> : <Copy />}
      </span>
    </span>
  )
}

interface IClickToCopyButton extends Omit<IClickToCopy, "children">  {
  textToCopy: string
}

export const ClickToCopyButton: FC<IClickToCopyButton> = ({ className, noticeClassName, textToCopy }) => {
  const [isCopied, setIsCopied] = useState(false)
  

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
      className={classNames('flex items-center gap-2 cursor-pointer relative hover:bg-black/10 dark:hover:bg-white/5 border rounded-md p-1 text-sm', {
        [`${className}`]: Boolean(className),
      })}
      onClick={() => {
        navigator.clipboard.writeText(textToCopy)
        setIsCopied(true)
      }}
      title="Click to copy"
    >
      {isCopied ? (
        <span
          className={classNames(
            'bg-dark text-light dark:bg-light dark:text-dark text-sm leading-none px-2 py-1.5 rounded drop-shadow-md max-w-96 absolute z-10 -top-6 right-0',
            {             
              [`${noticeClassName}`]: Boolean(noticeClassName),
            }
          )}
        >
          Copied
        </span>
      ) : null}
      <span>
        {isCopied ? <Check /> : <Copy />}
      </span>
    </span>
  )
}
