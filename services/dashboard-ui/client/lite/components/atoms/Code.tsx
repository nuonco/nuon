import type { HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'
import { Text } from './Text'

export interface ICode extends HTMLAttributes<HTMLElement> {
  loading?: boolean
  loadingWidth?: number
}

const CODE_CLASSES =
  'rounded bg-code-bg px-1 py-0.5 font-mono text-[0.9em] text-primary'

export const Code = ({
  loading,
  loadingWidth,
  className,
  children,
  ...props
}: ICode) => {
  if (loading) {
    return (
      <Text
        as="code"
        family="mono"
        loading
        loadingWidth={loadingWidth}
        className={className}
      />
    )
  }

  return (
    <code className={cn(CODE_CLASSES, className)} {...props}>
      {children}
    </code>
  )
}
