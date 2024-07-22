import classNames from 'classnames'
import { DateTime } from 'luxon'
import React, { type FC } from 'react'
import { Code } from '@/components'
import type { TWaypointLog } from '@/types'

export interface ILogs {
  logs: Array<TWaypointLog>
}

export const Logs: FC<ILogs> = ({ logs }) => {
  return (
    <Code>
      {logs?.length
        ? logs.map((term, ii) => {
            // handle complete state

            return (
              <span key={ii}>
                {term?.Terminal?.events?.length
                  ? term?.Terminal?.events?.map((l, i) => {
                      let line = null

                      if (l?.line) {
                        line = (
                          <span
                            key={`${l?.line?.msg}-${i}`}
                            className="block text-xs"
                          >
                            {l?.line?.msg}
                          </span>
                        )
                      }

                      // raw data

                      if (l?.raw?.data) {
                        line = (
                          <span
                            key={`${l?.raw?.data}-${i}`}
                            className="block text-xs"
                          >
                            {atob(l?.raw?.data)}
                          </span>
                        )
                      }

                      if (l?.step) {
                        line = (
                          <span
                            key={`${l?.step?.msg}-${i}`}
                            className="block text-xs"
                          >
                            {l?.step?.msg}
                          </span>
                        )
                      }

                      // status
                      if (l?.status) {
                        line = (
                          <span
                            key={`${l?.status?.msg}-${i}`}
                            className="block text-xs"
                          >
                            {l?.status?.msg}
                          </span>
                        )
                      }

                      return line
                    })
                  : null}{' '}
                {term?.State?.current as string}
              </span>
            )
          })
        : 'no logs to show'}
    </Code>
  )
}

const LogLine: FC<{ line: any }> = ({ line }) => {
  const lineStyle = 'tracking-wider text-[10px] leading-none'

  return (
    <span className="flex items-center justify-start gap-6 w-max">
      <span
        className={classNames('flex w-1.5 h-4', {
          'bg-blue-300': line.ServityNumber <= 4,
          'bg-gray-500': line?.SeverityNumber >= 5 && line?.SeverityNumber <= 8,
          'bg-fuchsia-500':
            line?.SeverityNumber >= 9 && line?.SeverityNumber <= 12,
          'bg-yellow-500':
            line?.SeverityNumber >= 13 && line?.SeverityNumber <= 16,
          'bg-red-400':
            line?.SeverityNumber >= 17 && line?.SeverityNumber <= 20,
          'bg-red-600':
            line?.SeverityNumber >= 21 && line?.SeverityNumber <= 24,
        })}
      />

      <span className={lineStyle}>
        {DateTime.fromMillis(line?.Timestamp).toLocaleString(
          DateTime.DATETIME_MED_WITH_SECONDS
        )}
      </span>

      <span className={lineStyle}>{line?.Resource?.['service.name']}</span>

      <span className={lineStyle + ' font-semibold uppercase'}>
        {line?.SeverityText || 'UNKOWN'}
      </span>

      <span className={lineStyle + ' text-gray-100/75'}>{line?.Body}</span>
    </span>
  )
}

export const OTELLogs: FC<{ logs?: Array<Record<string, unknown>> }> = ({
  logs = [],
}) => {
  return (
    <Code className="gap-4">
      {logs.map((line) => (
        <LogLine key={line?.timestamp?.toString()} line={line} />
      ))}
    </Code>
  )
}
