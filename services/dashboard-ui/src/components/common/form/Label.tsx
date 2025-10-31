import { type HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'

export interface ILabel extends HTMLAttributes<HTMLLabelElement> {}

export const Label = ({ children, className, ...props }: ILabel) => {
  return (
    <label className={cn('', className)} {...props}>
      {children}
    </label>
  )
}
