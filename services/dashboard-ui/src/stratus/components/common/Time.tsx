import { DateTime, Duration as LuxonDuration, type DurationUnits } from 'luxon'
import React from 'react'

export interface ITime
  extends Omit<React.HTMLAttributes<HTMLSpanElement>, 'role'> {
  format?: 'default' | 'long' | 'relative' | 'time-only'
  time?: string
}

export const Time = ({ format, time, ...props }: ITime) => {
  const datetime = time ? DateTime.fromISO(time) : DateTime.now()

  return (
    <span {...props} role="time">
      {format === 'relative'
        ? datetime.toRelative()
        : datetime.toLocaleString(
            format === 'long'
              ? DateTime.DATETIME_FULL_WITH_SECONDS
              : format === 'time-only'
                ? DateTime.TIME_SIMPLE
                : DateTime.DATETIME_SHORT_WITH_SECONDS
          )}
    </span>
  )
}
