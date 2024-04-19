import { DateTime, type DurationUnits } from 'luxon'
import React, { type FC } from 'react'
import { Text, type IText } from '@/components'

export interface ITime extends IText {
  format?: 'default' | 'long' | 'relative'
  time?: string
}

export const Time: FC<ITime> = ({ format, time, ...props }) => {
  const datetime = time ? DateTime.fromISO(time) : DateTime.now()

  return (
    <Text {...props}>
      {format === 'relative'
        ? datetime.toRelative()
        : datetime.toLocaleString(
            format === 'long'
              ? DateTime.DATETIME_FULL_WITH_SECONDS
              : DateTime.DATETIME_SHORT_WITH_SECONDS
          )}
    </Text>
  )
}

export interface IDuration extends IText {
  beginTime: string
  endTime: string
  durationUnits?: DurationUnits
  listStyle?: 'narrow' | 'short' | 'long'
  unitDisplay?: 'narrow' | 'short' | 'long'
  format?: 'default' | 'timer'
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
  unitDisplay = 'narrow',
  ...props
}) => {
  const bt = DateTime.fromISO(beginTime)
  const et = DateTime.fromISO(endTime)
  const duration = et.diff(bt, durationUnits)

  return (
    <Text {...props}>
      {format === 'timer'
        ? duration.toFormat('T-hh:mm:ss:SS')
        : duration.rescale().toHuman({
            listStyle,
            unitDisplay,
          })}
    </Text>
  )
}
