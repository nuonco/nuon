import type { SVGAttributes } from 'react'
import { cn } from '@/utils/classnames'

export interface ISpinner extends SVGAttributes<SVGSVGElement> {
  size?: number
  label?: string
}

export const Spinner = ({ size = 16, label, className, ...props }: ISpinner) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 16 16"
    fill="none"
    role={label ? 'status' : undefined}
    aria-label={label}
    aria-hidden={label ? undefined : true}
    className={cn('animate-spin motion-reduce:animate-none', className)}
    {...props}
  >
    <circle
      cx="8"
      cy="8"
      r="6.5"
      stroke="currentColor"
      strokeOpacity="0.25"
      strokeWidth="2"
    />
    <path
      d="M8 1.5a6.5 6.5 0 0 1 6.5 6.5"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
    />
  </svg>
)
