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
  interactiveHeading?: boolean
  toggleLabel?: string
  toggleContent?: React.ReactNode
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
  interactiveHeading = false,
  toggleLabel,
  toggleContent,
  ...props
}: IExpand) => {
  const [isExpanded, setIsExpanded] = useState(isOpen)

  useEffect(() => {
    setIsExpanded(isOpen)
  }, [isOpen])

  const toggle = () => setIsExpanded((prev) => !prev)

  const expandIcon = isExpanded ? (
    <Icon variant="CaretUpIcon" />
  ) : (
    <Icon variant="CaretDownIcon" />
  )

  const headingNode =
    typeof heading === 'string' ? <Text>{heading}</Text> : heading

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
      {interactiveHeading ? (
        <div
          className={cn(
            'flex items-start gap-2 p-2 w-full',
            {
              'justify-between': !isIconBeforeHeading,
            },
            headerClassName
          )}
        >
          {isIconBeforeHeading && (
            <button
              type="button"
              className="flex items-center gap-2 cursor-pointer outline-none rounded px-2 py-1 transition-all hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10"
              aria-expanded={isExpanded}
              aria-controls={`${id}-content`}
              aria-label={toggleLabel}
              id={id}
              onClick={toggle}
            >
              {expandIcon}
              {toggleContent}
            </button>
          )}
          {headingNode}
          {!isIconBeforeHeading && (
            <button
              type="button"
              className="flex items-center gap-2 cursor-pointer outline-none rounded px-2 py-1 transition-all hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10"
              aria-expanded={isExpanded}
              aria-controls={`${id}-content`}
              aria-label={toggleLabel}
              id={id}
              onClick={toggle}
            >
              {toggleContent}
              {expandIcon}
            </button>
          )}
        </div>
      ) : (
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
          onClick={toggle}
        >
          {isIconBeforeHeading && expandIcon}
          {headingNode}
          {!isIconBeforeHeading && expandIcon}
        </button>
      )}

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
