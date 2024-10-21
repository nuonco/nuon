'use client'

import classNames from 'classnames'
import React, { type FC, useEffect, useState } from 'react'

export interface IExpand extends React.HTMLAttributes<HTMLDivElement> {
  expandContent: React.ReactElement
  heading: React.ReactElement | React.ReactNode
  isOpen?: boolean
  id: string
}

export const Expand: FC<IExpand> = ({
  className,
  expandContent,
  heading,
  id,
  isOpen = false,
}) => {
  const [isExpanded, setIsExpanded] = useState(isOpen)

  useEffect(() => {
    setIsExpanded(isOpen)
  }, [isOpen])

  return (
    <div className={classNames('')}>
      <div
        className={classNames(
          'cursor-pointer hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10',
          {
            [`${className}`]: Boolean(className),
          }
        )}
        onClick={() => {
          setIsExpanded(!isExpanded)
        }}
      >
        {heading}
      </div>
      {isExpanded && (
        <div key={`${id}-content`} className="w-full">
          {expandContent}
        </div>
      )}
    </div>
  )
}
