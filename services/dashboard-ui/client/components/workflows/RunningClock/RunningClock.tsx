import { useEffect, useState } from 'react'
import { DateTime } from 'luxon'
import { Text, type IText } from '@/components/common/Text'

const formatClock = (ms: number) => {
  const totalSeconds = Math.max(0, Math.floor(ms / 1000))
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = String(totalSeconds % 60).padStart(2, '0')
  return hours > 0
    ? `${hours}:${String(minutes).padStart(2, '0')}:${seconds}`
    : `${minutes}:${seconds}`
}

export interface IRunningClock extends Pick<IText, 'variant' | 'theme'> {
  startedAt?: string
  label?: string
}

export const RunningClock = ({
  startedAt,
  label = 'Running',
  variant = 'subtext',
  theme = 'neutral',
}: IRunningClock) => {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(id)
  }, [])

  const start = startedAt ? DateTime.fromISO(startedAt) : undefined
  if (!start?.isValid || start.year <= 1) return null

  return (
    <Text variant={variant} theme={theme} role="time">
      {label ? `${label} ` : null}
      <span className="tabular-nums">
        {formatClock(now - start.toMillis())}
      </span>
    </Text>
  )
}
