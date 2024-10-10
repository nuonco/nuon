'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { LogLineSeverity, Text, Time } from '@/components'
import type { TOTELLog } from '@/types'

const LogLinePreview: FC<{ line: TOTELLog; isPreview?: boolean }> = ({
  line,
  isPreview = false,
}) => {
  const lineStyle =
    'tracking-wider text-sm font-mono leading-loose text-cool-grey-600 dark:text-cool-grey-500'

  return (
    <span className="grid grid-cols-12 items-center justify-start gap-6 py-2 w-full">
      {isPreview ? null : (
        <span className={classNames('flex items-center gap-2')}>
          <LogLineSeverity severity_number={line.severity_number} />
          <span className={lineStyle + ' font-semibold uppercase'}>
            {line?.severity_text || 'UNKOWN'}
          </span>
        </span>
      )}

      <span
        className={classNames(lineStyle, {
          'col-span-2': !isPreview,
          'col-span-4 flex items-center gap-2': isPreview,
        })}
      >
        {isPreview && (
          <LogLineSeverity severity_number={line.severity_number} />
        )}
        <Time className="!text-sm" time={line.timestamp} />
      </span>

      {!isPreview && (
        <span
          className={classNames(lineStyle, {
            'col-span-2': !isPreview,
            'col-span-3': isPreview,
          })}
        >
          {line?.resource_attributes?.['service.name']}
        </span>
      )}

      <span
        className={classNames(lineStyle, {
          'col-span-7': !isPreview,
          'col-span-8 truncate': isPreview,
        })}
      >
        {line?.body}
      </span>
    </span>
  )
}

export interface ILogsPreview {
  logs: Array<TOTELLog>
}

export const LogsPreview: FC<ILogsPreview> = ({ logs }) => {
  return (
    <div className="divide-y">
      <div className="grid grid-cols-12 items-center justify-start gap-6 py-2 w-full">
        <Text className="!font-medium text-cool-grey-600 dark:text-cool-grey-500 col-span-4">
          Date
        </Text>

        <Text className="!font-medium text-cool-grey-600 dark:text-cool-grey-500 col-span-8">
          Content
        </Text>
      </div>
      {logs.slice(0, 15).map((line) => (
        <LogLinePreview key={line?.timestamp as string} line={line} isPreview />
      ))}
    </div>
  )
}
