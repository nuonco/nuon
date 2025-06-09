import classNames from 'classnames'
import React, { type FC } from 'react'

interface ISection extends React.HTMLAttributes<HTMLDivElement> {}

export const Section: FC<ISection> = ({ className, children, ...props }) => {
  return (
    <section
      className={classNames('p-4 md:p-6 w-full flex flex-col', {
        [`${className}`]: Boolean(className),
      })}
      {...props}
    >
      {children}
    </section>
  )
}
