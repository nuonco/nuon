import cronstrue from 'cronstrue'
import type { ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import { Code } from '../atoms/Code'
import { Text, type IText } from '../atoms/Text'
import { Tooltip } from '../atoms/Tooltip'

export type TCronFormat = 'human' | 'expression' | 'both'

export interface ICron extends Omit<IText, 'as' | 'children'> {
  value?: string
  format?: TCronFormat
  tooltip?: boolean
  fallback?: ReactNode
}

const describe = (value: string) => {
  try {
    return {
      valid: true,
      human: cronstrue.toString(value, {
        throwExceptionOnParseError: true,
        verbose: false,
      }),
    }
  } catch {
    return { valid: false, human: 'Invalid schedule' }
  }
}

export const Cron = ({
  value,
  format = 'human',
  tooltip = true,
  fallback = '—',
  loading = false,
  loadingWidth,
  variant = 'caption',
  color = 'secondary',
  family,
  className,
  ...props
}: ICron) => {
  if (loading) {
    return (
      <Text
        variant={variant}
        color={color}
        family={family}
        loading
        loadingWidth={loadingWidth ?? 18}
        className={className}
        {...props}
      />
    )
  }

  if (!value) {
    return (
      <Text
        variant={variant}
        color="tertiary"
        family={family}
        className={className}
        {...props}
      >
        {fallback}
      </Text>
    )
  }

  const { valid, human } = describe(value)
  const expressionChip = (
    <Code className={valid ? undefined : 'text-tertiary'}>{value}</Code>
  )
  const expression = (
    <Text variant={variant} className={className} {...props}>
      {expressionChip}
    </Text>
  )
  const humanText = (
    <Text
      variant={variant}
      color={valid ? color : 'tertiary'}
      family={family}
      className={className}
      {...props}
    >
      {human}
    </Text>
  )

  if (format === 'both') {
    return (
      <span
        className={cn(
          'inline-flex min-w-0 flex-wrap items-baseline gap-x-2 gap-y-1',
          className
        )}
        {...props}
      >
        <Text
          variant={variant}
          color={valid ? color : 'tertiary'}
          family={family}
        >
          {human}
        </Text>
        <Text variant={variant}>{expressionChip}</Text>
      </span>
    )
  }

  const content = format === 'expression' ? expression : humanText
  if (!tooltip) return content

  return (
    <Tooltip
      content={
        valid
          ? format === 'expression'
            ? human
            : value
          : `Invalid schedule: ${value}`
      }
    >
      {content}
    </Tooltip>
  )
}
