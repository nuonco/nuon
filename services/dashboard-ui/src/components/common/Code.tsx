'use client'

import { cn } from '@/utils/classnames'

export interface ICode extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'default' | 'preformated' | 'inline'
}

const BASE_CLASSES =
  'text-sm bg-code text-blue-800 dark:text-blue-500 font-mono break-all flex flex-col rounded shadow-sm overflow-auto'
const VARIANT_CLASSES: Record<string, string> = {
  default: 'p-4 min-h-[3rem] max-h-[40rem]',
  preformated: 'p-4 min-h-[3rem] max-h-[40rem]',
  inline:
    '!p-1 leading-3 min-h-min overflow-x-scroll w-fit inline-block align-middle',
}

export const Code = ({
  className,
  children,
  variant = 'default',
  ...props
}: ICode) => {
  const classes = cn(BASE_CLASSES, VARIANT_CLASSES[variant] || '', className)

  if (variant === 'preformated') {
    return (
      <pre className={classes} {...props}>
        {children}
      </pre>
    )
  }

  return (
    <code className={classes} {...props}>
      {variant === 'inline' ? (
        <span className="block min-w-max">{children}</span>
      ) : (
        children
      )}
    </code>
  )
}
