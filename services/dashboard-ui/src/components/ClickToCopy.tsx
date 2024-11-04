'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { Copy } from '@phosphor-icons/react'

export const ClickToCopy: FC<React.HTMLAttributes<HTMLSpanElement>> = ({
  className,
  children,
  ...props
}) => {
  const text = children?.valueOf()?.['props']?.children || children

  return (
    <span
      className={classNames('flex items-center gap-2 cursor-pointer', {
        [`${className}`]: Boolean(className),
      })}
      onClick={() => {
        console.log('logging', text)
        navigator.clipboard.writeText(text)
      }}
      title="Click to copy"
      {...props}
    >
      {children}
      <Copy />
    </span>
  )
}
