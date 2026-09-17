import type { HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'
import { Text } from './Text'

export interface ICode extends HTMLAttributes<HTMLElement> {
  loading?: boolean
  loadingWidth?: number
}

const CODE_CLASSES =
  'rounded-[0.3em] bg-code-inline-bg px-[0.35em] py-[0.1em] font-mono text-[0.9em] leading-[1.45] text-primary'

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
