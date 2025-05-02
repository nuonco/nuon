'use client'

import classNames from 'classnames'
import React, { type FC, useEffect, useState } from 'react'
import { CaretDown, CaretUp } from '@phosphor-icons/react'

export interface IExpand extends React.HTMLAttributes<HTMLDivElement> {
  expandContent: React.ReactElement | Array<React.ReactElement>
  heading: React.ReactElement | React.ReactNode
  isOpen?: boolean
  isIconBeforeHeading?: boolean
  hasHeadingStyle?: boolean
  hasNoHoverStyle?: boolean
  headerClass?: string
  id: string
  parentClass?: string
}

export const Expand: FC<IExpand> = ({
  className,
  expandContent,
  heading,
  id,
  hasHeadingStyle = false,
  hasNoHoverStyle = false,
  headerClass,
  isIconBeforeHeading = false,
  isOpen = false,
  parentClass,
}) => {
  const [isExpanded, setIsExpanded] = useState(isOpen)

  useEffect(() => {
    setIsExpanded(isOpen)
  }, [isOpen])

  const ExpandIcon = isExpanded ? (
    <CaretUp className="text-sm" />
  ) : (
    <CaretDown className="text-sm" />
  )

  return (
    <div
      className={classNames('w-full', {
        [`${parentClass}`]: Boolean(parentClass),
      })}
    >
      <div
        className={classNames(
          'flex items-center justify-between cursor-pointer pr-2',
          {
            'border-t border-b bg-cool-grey-50 dark:bg-dark-grey-200 text-cool-grey-600 dark:text-cool-grey-500':
              hasHeadingStyle,
            'hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10':
              !hasNoHoverStyle,
            [`${headerClass}`]: Boolean(headerClass),
          }
        )}
        onClick={() => {
          setIsExpanded(!isExpanded)
        }}
      >
        {isIconBeforeHeading ? ExpandIcon : null}
        <div className={classNames({ [`${className}`]: Boolean(className) })}>
          {heading}
        </div>
        {isIconBeforeHeading ? null : ExpandIcon}
      </div>
      {isExpanded && (
        <div key={`${id}-content`} className="w-full">
          {expandContent}
        </div>
      )}
    </div>
  )
}
