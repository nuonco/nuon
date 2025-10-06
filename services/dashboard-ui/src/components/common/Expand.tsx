'use client'

import React, { useEffect, useState } from 'react'
import { Icon } from './Icon'
import { Text } from './Text'
import { TransitionDiv } from './TransitionDiv'
import { cn } from '@/utils/classnames'
import './Expand.css'

export interface IExpand extends React.HTMLAttributes<HTMLDivElement> {
  heading: React.ReactNode | string
  isOpen?: boolean
  isIconBeforeHeading?: boolean
  hasNoHoverStyle?: boolean
  headerClassName?: string
  id: string
}

export const Expand = ({
  className,
  children,
  heading,
  id,
  hasNoHoverStyle = false,
  headerClassName,
  isIconBeforeHeading = false,
  isOpen = false,
  ...props
}: IExpand) => {
  const [isExpanded, setIsExpanded] = useState(isOpen)

  useEffect(() => {
    setIsExpanded(isOpen)
  }, [isOpen])

  const expandIcon = isExpanded ? (
    <Icon variant="CaretUp" />
  ) : (
    <Icon variant="CaretDown" />
  )

  return (
    <div
      className={cn(
        'expand-wrapper shrink-0 flex flex-col w-full overflow-hidden',
        {
          'is-expanded': isExpanded,
        },
        className
      )}
      {...props}
    >
      <button
        type="button"
        className={cn(
          'flex items-center gap-2 cursor-pointer p-2 w-full outline-none transition-all',
          {
            'justify-between': !isIconBeforeHeading,
            'hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10':
              !hasNoHoverStyle,
          },
          headerClassName
        )}
        aria-expanded={isExpanded}
        aria-controls={`${id}-content`}
        id={id}
        onClick={() => setIsExpanded((prev) => !prev)}
      >
        {isIconBeforeHeading && expandIcon}
        {typeof heading === 'string' ? <Text>{heading}</Text> : heading}
        {!isIconBeforeHeading && expandIcon}
      </button>

      <TransitionDiv
        isVisible={isExpanded}
        key={`${id}-content`}
        id={`${id}-content`}
        className="expand w-full overflow-hidden"
      >
        {children}
      </TransitionDiv>
    </div>
  )
}
