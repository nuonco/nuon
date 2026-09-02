import type { ReactNode } from 'react'
import { cn } from '@/utils/classnames'

export const STEP_GUTTER = 'px-4 sm:px-6'

interface IStepLayoutBlock {
  className?: string
  children: ReactNode
}

export const StepBlock = ({ className, children }: IStepLayoutBlock) => (
  <div className={cn('flex flex-col gap-3 py-4', STEP_GUTTER, className)}>
    {children}
  </div>
)

export const StepRowList = ({ className, children }: IStepLayoutBlock) => (
  <div className={cn('flex flex-col divide-y', className)}>{children}</div>
)

export const StepRow = ({ className, children }: IStepLayoutBlock) => (
  <div className={cn('flex items-center gap-3 py-3', STEP_GUTTER, className)}>
    {children}
  </div>
)
