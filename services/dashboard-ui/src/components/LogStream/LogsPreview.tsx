'use client'

import classNames from 'classnames'
import React, { type FC, useMemo } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { LogsViewer } from './LogsViewer'
import { LogLineSeverity } from './LogLineSeverity'
import type { TLogRecord } from './types'
import { Time } from '@/components/Time'

export interface ILogsPreview {
  logs: Array<TLogRecord>
}

export const LogsPreview: FC<ILogsPreview> = ({ logs }) => {
  const lineStyle =
    'text-sm font-mono text-cool-grey-600 dark:text-cool-grey-500'
  const columns: Array<ColumnDef<TLogRecord>> = useMemo(
    () => [
      {
        header: 'Date',
        accessorKey: 'timestamp',
        cell: (props) => {
          return (
            <span
              className={classNames(lineStyle, {
                'col-span-3 flex items-center gap-2': true,
              })}
            >
              <LogLineSeverity
                severity_number={props.row.original?.severity_number}
              />
              <Time
                className="!text-[11px]"
                time={props.getValue<string>()}
                useMicro
              />
            </span>
          )
        },
      },
      {
        header: 'Content',
        accessorKey: 'body',
        cell: (props) => (
          <span
            className={classNames(lineStyle, {
              'col-span-8 break-all': true,
            })}
          >
            {props.getValue<string>()}
          </span>
        ),
      },
    ],
    []
  )

  return <LogsViewer columns={columns} data={logs} />
}
