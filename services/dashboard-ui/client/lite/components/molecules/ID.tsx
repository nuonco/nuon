import type { HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'
import { CopyButton } from '../atoms/CopyButton'
import { Text } from '../atoms/Text'

export interface IID extends Omit<HTMLAttributes<HTMLSpanElement>, 'children'> {
  value: string
  truncate?: boolean
  head?: number
  tail?: number
  copyable?: boolean
  label?: string
  loading?: boolean
  loadingWidth?: number
}

export const middleTruncate = (value: string, head: number, tail: number) =>
  value.length <= head + tail + 1 ? value : `${value.slice(0, head)}…${value.slice(-tail)}`

export const ID = ({
  value,
  truncate = false,
  head = 10,
  tail = 4,
  copyable = true,
  label = 'Copy ID',
  loading = false,
  loadingWidth,
  className,
  ...props
}: IID) => {
  if (loading) {
    return (
      <Text
        variant="caption"
        family="mono"
        loading
        loadingWidth={loadingWidth ?? 22}
        className={className}
      />
    )
  }

  return (
    <span className={cn('inline-flex w-fit items-center gap-1', className)} {...props}>
      <Text variant="caption" family="mono" color="tertiary" className="select-all">
        {truncate ? middleTruncate(value, head, tail) : value}
      </Text>
      {copyable ? <CopyButton value={value} label={label} /> : null}
    </span>
  )
}
