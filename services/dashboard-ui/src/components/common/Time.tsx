'use client'

import { DateTime, type DateTimeFormatOptions } from 'luxon'
import { Text, type IText } from './Text'
import { Tooltip, type ITooltip } from './Tooltip'

const SHORT_DATETIME_FORMAT: DateTimeFormatOptions = {
  year: 'numeric',
  month: 'numeric',
  day: 'numeric',
  hour: 'numeric',
  minute: '2-digit',
  second: '2-digit',
  hour12: true,
}

const LONG_DATETIME_FORMAT: DateTimeFormatOptions = {
  year: 'numeric',
  month: 'long',
  day: 'numeric',
  hour: 'numeric',
  minute: '2-digit',
  second: '2-digit',
  timeZoneName: 'short',
  hour12: true,
}

const TIME_ONLY_FORMAT: DateTimeFormatOptions = {
  hour: 'numeric',
  minute: '2-digit',
  hour12: true,
}

// Updated to include timezone abbreviation
const LOG_DATETIME_FORMAT_STRING = 'M/d/yyyy, h:mm:ss.SSSs a ZZZZ'

export interface ITime extends Omit<IText, 'role'> {
  format?:
    | 'short-datetime'
    | 'long-datetime'
    | 'relative'
    | 'time-only'
    | 'log-datetime'
  time?: string
  seconds?: number
  tooltipProps?: ITooltip
}

export const Time = ({
  format = 'short-datetime',
  time,
  seconds,
  ...props
}: ITime) => {
  let datetime: DateTime

  if (typeof seconds === 'number') {
    datetime = DateTime.fromSeconds(seconds)
  } else if (time) {
    datetime = DateTime.fromISO(time)
  } else {
    datetime = DateTime.now()
  }

  const getFormattedTime = () => {
    switch (format) {
      case 'relative':
        return datetime.toRelative()

      case 'long-datetime':
        return datetime.toLocaleString(LONG_DATETIME_FORMAT)

      case 'time-only':
        return datetime.toLocaleString(TIME_ONLY_FORMAT)

      case 'log-datetime':
        return datetime.toFormat(LOG_DATETIME_FORMAT_STRING)

      case 'short-datetime':
      default:
        return datetime.toLocaleString(SHORT_DATETIME_FORMAT)
    }
  }

  return format === 'relative' ? (
    <Tooltip tipContent={datetime.toLocaleString(LONG_DATETIME_FORMAT)}>
      <Text {...props} role="time">
        {getFormattedTime()}
      </Text>
    </Tooltip>
  ) : (
    <Text {...props} role="time">
      {getFormattedTime()}
    </Text>
  )
}
