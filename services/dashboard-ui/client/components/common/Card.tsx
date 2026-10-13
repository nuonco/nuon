import React from 'react'
import { cn } from '@/utils/classnames'

export type TCardElevation = '0' | '1' | '2' | '3'

const ELEVATION_CLASS: Record<TCardElevation, string> = {
  '0': '',
  '1': 'bg-elevation-1',
  '2': 'bg-elevation-2',
  '3': 'bg-elevation-3',
}

export interface ICard extends React.HTMLAttributes<HTMLDivElement> {
  elevation?: TCardElevation
}

export const Card = ({
  children,
  className,
  elevation = '0',
  ...props
}: ICard) => {
  return (
    <div
      className={cn(
        'flex flex-col gap-6 p-6 border rounded-md shadow-sm',
        ELEVATION_CLASS[elevation],
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}
