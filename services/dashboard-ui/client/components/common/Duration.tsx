'use client'

import { DateTime, Duration as LuxonDuration, type DurationUnits } from 'luxon'
import { Icon } from './Icon'
import { Text, type IText } from './Text'

export interface IDuration extends Omit<IText, 'role'> {
  beginTime?: string
  endTime?: string
  durationUnits?: DurationUnits
  format?: 'default' | 'timer'
  listStyle?: 'narrow' | 'short' | 'long'
  nanoseconds?: number
  unitDisplay?: 'narrow' | 'short' | 'long'
}

export const Duration = ({
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
}: IDuration) => {
  let duration: LuxonDuration | undefined

  if (typeof nanoseconds === 'number') {
    if (nanoseconds === 0) {
      return (
        <Text {...props}>
          <Icon variant="Minus" />
        </Text>
      )
    }
    const milliseconds = Math.round(nanoseconds / 1e6)
    duration = LuxonDuration.fromMillis(milliseconds)
  } else if (beginTime && endTime) {
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
        <Icon variant="Minus" />
      )}
    </Text>
  )
}
