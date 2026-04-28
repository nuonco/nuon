import React from 'react'
import { cn } from '@/utils/classnames'

type TVariant = 'column' | 'row' | 'primary' | 'secondary' | 'tertiary'

interface IPageContent extends React.HTMLAttributes<HTMLDivElement> {
  isScrollable?: boolean
  variant?: TVariant
}

const isRowVariant = (v: TVariant) =>
  v === 'row' || v === 'secondary' || v === 'tertiary'

export const PageContent = ({
  className,
  children,
  isScrollable: _isScrollable,
  variant = 'column',
  ...props
}: IPageContent) => {
  return (
    <div
      className={cn(
        'flex-1 flex flex-col min-w-0',
        { 'md:flex-row': isRowVariant(variant) },
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}
