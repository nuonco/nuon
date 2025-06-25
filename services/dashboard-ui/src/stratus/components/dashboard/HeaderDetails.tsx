import React from 'react'
import { cn } from '@/stratus/components/helpers'

export interface IHeaderDetails extends React.HTMLAttributes<HTMLDivElement> {}

export const HeaderDetails = ({
  className,
  children,
  ...props
}: IHeaderDetails) => {
  const childrenLength = React.Children.toArray(children)?.length

  return (
    <div
      className={cn(
        'flex flex-wrap items-center gap-3 md:gap-0 md:divide-x divide-dotted',
        className
      )}
      {...props}
    >
      {React.Children.map(children, (c, i) => {
        const clxs = {
          'md:pr-3': i === 0,
          'md:pl-3': i === childrenLength - 1,
          'md:px-3': i !== 0 && i !== childrenLength - 1,
        }
        return React.isValidElement(c) && c.type === 'div' ? (
          React.cloneElement<React.HTMLAttributes<HTMLDivElement>>(c, {
            className: cn(c.props.className, clxs),
          })
        ) : (
          <div className={cn(clxs)}>{c}</div>
        )
      })}
    </div>
  )
}
