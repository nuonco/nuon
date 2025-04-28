'use client'

import { DateTime, Duration as LuxonDuration, type DurationUnits } from 'luxon'
import React, { type FC, useState, useEffect } from 'react'
import { Minus } from '@phosphor-icons/react/dist/ssr'
import { Text, type IText } from '@/components/Typography'

export interface ITime extends Omit<IText, 'role'> {
  format?: 'default' | 'long' | 'relative' | 'time-only'
  time?: string
}

export const Time: FC<ITime> = ({ format, time, ...props }) => {
  const [datetime, setDateTime] = useState(
    time ? DateTime.fromISO(time) : DateTime.now()
  )

  useEffect(() => {
    if (format === 'relative') {
      const intervalId = setInterval(() => {
        setDateTime(time ? DateTime.fromISO(time) : DateTime.now())
      }, 1000)

      return () => clearInterval(intervalId)
    }
  }, [format])

  return (
    <Text {...props} role="time">
      {format === 'relative'
        ? datetime.toRelative()
        : datetime.toLocaleString(
            format === 'long'
              ? DateTime.DATETIME_FULL_WITH_SECONDS
              : format === 'time-only'
                ? DateTime.TIME_SIMPLE
                : DateTime.DATETIME_SHORT_WITH_SECONDS
          )}
    </Text>
  )
}

// TODO: normalize around duration format
export interface IDuration extends Omit<IText, 'role'> {
  beginTime?: string
  endTime?: string
  durationUnits?: DurationUnits
  format?: 'default' | 'timer'
  listStyle?: 'narrow' | 'short' | 'long'
  nanoseconds?: number
  unitDisplay?: 'narrow' | 'short' | 'long'
}

export const Duration: FC<IDuration> = ({
  beginTime,
  endTime,
  durationUnits = [
    'years',
    'months',
    'days',
    'hours',
    'minutes',
    'seconds',
    'milliseconds',
  ],
  format = 'default',
  listStyle = 'narrow',
  nanoseconds,
  unitDisplay = 'narrow',
  ...props
}) => {
  let duration: LuxonDuration
  if (nanoseconds !== undefined) {
    if (nanoseconds === 0) {
      return (
        <Text {...props}>
          <Minus />
        </Text>
      )
    }
    duration = LuxonDuration.fromMillis(Math.round(nanoseconds / 1000000))
  } else {
    const bt = DateTime.fromISO(beginTime)
    const et = DateTime.fromISO(endTime)
    duration = et.diff(bt, durationUnits)
  }

  return (
    <Text {...props} role="time">
      {format === 'timer'
        ? duration.toFormat('T-hh:mm:ss:SS')
        : duration.rescale().set({ milliseconds: 0 }).rescale().toHuman({
            listStyle,
            unitDisplay,
          })}
    </Text>
  )
}
