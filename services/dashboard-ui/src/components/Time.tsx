// @ts-nocheck
'use client'

import { DateTime, Duration as LuxonDuration, type DurationUnits } from 'luxon'
import React, { type FC } from 'react'
import { Minus } from '@phosphor-icons/react'
import { ToolTip, IToolTip } from '@/components/ToolTip'
import { Text, type IText } from '@/components/Typography'

export interface ITime extends Omit<IText, 'role'> {
  format?: 'default' | 'long' | 'relative' | 'time-only'
  time?: string
  nanos?: number | string | bigint // nanoseconds since epoch
  position?: IToolTip['position']
  alignment?: IToolTip['alignment']
}

function getDateTime({ time, nanos }: { time?: string, nanos?: number | string | bigint }) {
  if (nanos !== undefined && nanos !== null) {
    const ns = typeof nanos === 'bigint' ? nanos : BigInt(nanos)
    // Convert to milliseconds and round to nearest ms
    const ms = Number((ns + 500_000n) / 1_000_000n)
    return DateTime.fromMillis(ms).toLocal()
  } else if (time) {
    return DateTime.fromISO(time).toLocal()
  } else {
    return DateTime.now().toLocal()
  }
}

export const Time: FC<ITime & { useMicro?: boolean }> = ({
  alignment,
  format,
  time,
  nanos,
  position,
  useMicro = false,
  ...props
}) => {
  const datetime = useMicro ? DateTime.fromISO(time) : getDateTime({ time, nanos });
  let formatted: string

  if (format === 'relative') {
    formatted = datetime.toRelative() ?? ''
  } else if (format === 'long') {
    formatted = datetime.toFormat("yyyy-MM-dd HH:mm:ss.SSS")
  } else if (format === 'time-only') {
    formatted = datetime.toFormat("HH:mm:ss.SSS")
  } else {
    // default (old-style): 7/29/2025, 8:21:11:562 PM
    formatted = datetime.toFormat("M/d/yyyy, h:mm:ss.SSSs a")
  }

  const TimeComp = (
    <Text {...props} role="time">
      {formatted}
    </Text>
  )

  return format === 'relative' ? (
    <ToolTip
      tipContent={datetime.toFormat("M/d/yyyy, h:mm:ss:SSS a")}
      alignment={alignment}
      position={position}
    >
      {TimeComp}
    </ToolTip>
  ) : (
    TimeComp
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
    const milliseconds = Math.round(nanoseconds / 1e6)
    duration = LuxonDuration.fromMillis(milliseconds)
  } else {
    const bt = DateTime.fromISO(beginTime)
    const et = DateTime.fromISO(endTime)
    duration = et.diff(bt, durationUnits)
  }

  return (
    <Text {...props} role="time">
      {duration?.isValid ? (
        format === 'timer' ? (
          duration.toFormat('T-hh:mm:ss:SS')
        ) : duration.as('seconds') < 1 ? (
          duration.rescale().toHuman({ listStyle, unitDisplay })
        ) : (
          duration.rescale().set({ milliseconds: 0 }).rescale().toHuman({
            listStyle,
            unitDisplay,
          })
        )
      ) : (
        <Minus />
      )}
    </Text>
  )
}
