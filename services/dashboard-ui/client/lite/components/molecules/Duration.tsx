import { DateTime, Duration as LuxonDuration } from 'luxon'
import type { ReactNode } from 'react'
import { useNow } from '../../hooks/use-now'
import type { TPopoverSide } from '../../hooks/use-popover'
import { Text, type IText } from '../atoms/Text'
import { Tooltip } from '../atoms/Tooltip'

export type TDurationFormat = 'compact' | 'long' | 'timer'

export interface IDuration extends Omit<IText, 'as' | 'children'> {
  nanoseconds?: number
  start?: string
  end?: string
  format?: TDurationFormat
  tooltip?: boolean
  tooltipSide?: TPopoverSide
  live?: boolean
  fallback?: ReactNode
}

type TDurationPart = {
  long: string
  compact: string
}

const UNITS: Array<{
  key: 'days' | 'hours' | 'minutes' | 'seconds'
  one: string
  many: string
  short: string
}> = [
  { key: 'days', one: 'day', many: 'days', short: 'd' },
  { key: 'hours', one: 'hour', many: 'hours', short: 'h' },
  { key: 'minutes', one: 'minute', many: 'minutes', short: 'm' },
  { key: 'seconds', one: 'second', many: 'seconds', short: 's' },
]

const durationParts = (milliseconds: number): TDurationPart[] => {
  if (milliseconds === 0) return [{ compact: '0s', long: '0 seconds' }]
  if (milliseconds < 1) {
    return [{ compact: '< 1ms', long: 'less than 1 millisecond' }]
  }
  if (milliseconds < 1000) {
    const value = Math.round(milliseconds * 10) / 10
    return [
      {
        compact: `${value}ms`,
        long: `${value} ${value === 1 ? 'millisecond' : 'milliseconds'}`,
      },
    ]
  }

  const shifted = LuxonDuration.fromMillis(milliseconds).shiftTo(
    'days',
    'hours',
    'minutes',
    'seconds'
  )

  const parts = UNITS.flatMap(({ key, one, many, short }) => {
    const raw = shifted.get(key)
    const value = key === 'seconds' ? Math.floor(raw) : raw
    if (!value) return []
    return [
      {
        compact: `${value}${short}`,
        long: `${value} ${value === 1 ? one : many}`,
      },
    ]
  })
  const remainingMilliseconds = Math.round(milliseconds % 1000)
  if (remainingMilliseconds) {
    parts.push({
      compact: `${remainingMilliseconds}ms`,
      long: `${remainingMilliseconds} ${
        remainingMilliseconds === 1 ? 'millisecond' : 'milliseconds'
      }`,
    })
  }
  return parts
}

const timerValue = (milliseconds: number) => {
  const totalSeconds = Math.floor(milliseconds / 1000)
  const days = Math.floor(totalSeconds / 86_400)
  const hours = Math.floor((totalSeconds % 86_400) / 3_600)
  const minutes = Math.floor((totalSeconds % 3_600) / 60)
  const seconds = totalSeconds % 60
  const clock = [hours, minutes, seconds]
    .map((value) => String(value).padStart(2, '0'))
    .join(':')
  return days ? `${days}d ${clock}` : clock
}

const resolveMilliseconds = ({
  nanoseconds,
  start,
  end,
  now,
}: {
  nanoseconds?: number
  start?: string
  end?: string
  now: number
}) => {
  if (typeof nanoseconds === 'number') return nanoseconds / 1e6
  if (!start) return undefined

  const begin = DateTime.fromISO(start)
  const finish = end ? DateTime.fromISO(end) : DateTime.fromMillis(now)
  if (!begin.isValid || !finish.isValid) return undefined
  return finish.diff(begin).as('milliseconds')
}

export const Duration = ({
  nanoseconds,
  start,
  end,
  format = 'compact',
  tooltip = true,
  tooltipSide = 'top',
  live = Boolean(start && !end),
  fallback = '—',
  loading = false,
  loadingWidth,
  variant = 'caption',
  color = 'secondary',
  family,
  className,
  ...props
}: IDuration) => {
  const now = useNow(live, 1000)
  const milliseconds = resolveMilliseconds({ nanoseconds, start, end, now })
  const valid =
    milliseconds !== undefined &&
    Number.isFinite(milliseconds) &&
    milliseconds >= 0

  if (loading) {
    return (
      <Text
        variant={variant}
        color={color}
        family={family}
        loading
        loadingWidth={loadingWidth ?? 8}
        className={className}
        {...props}
      />
    )
  }

  if (!valid) {
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

  const parts = durationParts(milliseconds)
  const exact = parts.map(({ long }) => long).join(', ')
  const display =
    format === 'timer'
      ? timerValue(milliseconds)
      : parts
          .slice(0, format === 'long' ? 3 : 2)
          .map((part) => part[format])
          .join(format === 'long' ? ', ' : ' ')
  const iso = LuxonDuration.fromMillis(milliseconds).toISO()
  const content = (
    <time dateTime={iso ?? undefined} className="inline-flex">
      <Text
        variant={variant}
        color={color}
        family={family ?? (format === 'timer' ? 'mono' : 'sans')}
        className={className}
        {...props}
      >
        {display}
      </Text>
    </time>
  )

  if (!tooltip) return content
  return (
    <Tooltip side={tooltipSide} content={exact}>
      {content}
    </Tooltip>
  )
}
