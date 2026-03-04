import { type LabelHTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'

export interface ILabel extends LabelHTMLAttributes<HTMLLabelElement> {}

export const Label = ({ children, className, ...props }: ILabel) => {
  return (
    <label className={cn('', className)} {...props}>
      {children}
    </label>
  )
}
