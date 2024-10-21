'use client'

import classNames from 'classnames'
import React, { type FC, useEffect, useState } from 'react'
import { CaretDown, CaretUp } from '@phosphor-icons/react'

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
          'flex items-center justify-between cursor-pointer hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10 pr-2'
        )}
        onClick={() => {
          setIsExpanded(!isExpanded)
        }}
      >
        <div
          className={classNames({
            [`${className}`]: Boolean(className),
          })}
        >
          {heading}
        </div>

        {isExpanded ? <CaretUp /> : <CaretDown />}
      </div>
      {isExpanded && (
        <div key={`${id}-content`} className="w-full">
          {expandContent}
        </div>
      )}
    </div>
  )
}
