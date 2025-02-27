'use client'

import classNames from 'classnames'
import React, { type FC, useMemo } from 'react'
import { DateTime } from 'luxon'
import { ArrowDown, ArrowUp } from '@phosphor-icons/react'
import {
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  flexRender,
  useReactTable,
  type ColumnDef,
  type ColumnSort,
} from '@tanstack/react-table'
import { LogLineSeverity } from '@/components/RunnerLogLineSeverity'
import { Time } from '@/components/Time'
import { Text } from '@/components/Typography'
import type { TOTELLog } from '@/types'

export interface ILogsPreview {
  data: Array<Record<string, any>>
  globalFilter: string
  sorting: Array<ColumnSort>
}

export const LogsPreview: FC<ILogsPreview> = ({
  data,
  globalFilter,
  sorting,
}) => {
  const lineStyle =
    'text-sm font-mono text-cool-grey-600 dark:text-cool-grey-500'
  const columns: Array<ColumnDef<TOTELLog>> = useMemo(
    () => [
      {
        header: 'Date',
        accessorKey: 'timestamp',
        cell: (props) => (
          <span
            className={classNames(lineStyle, {
              'col-span-4 flex items-center gap-2': true,
            })}
          >
            <LogLineSeverity
              severity_number={props.row.original?.severity_number}
            />
            <Time
              className="!text-sm"
              time={DateTime.fromMillis(props.getValue<number>()).toISO()}
            />
          </span>
        ),
      },
      {
        header: 'Content',
        accessorKey: 'body',
        cell: (props) => (
          <span
            className={classNames(lineStyle, {
              'col-span-7 break-all': true,
            })}
          >
            {props.getValue<string>()}
          </span>
        ),
      },
    ],
    []
  )

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    state: { globalFilter, sorting },
  })

  return (
    <div className="flex flex-col gap-8">
      <div className="divide-y">
        {table.getHeaderGroups().map((group) => (
          <div
            key={group.id}
            className="grid grid-cols-12 items-center justify-start gap-6 py-2 w-full"
          >
            {group.headers.map((header, i) => (
              <Text
                key={header.id}
                className={classNames(
                  '!font-medium text-cool-grey-600 dark:text-cool-grey-500',
                  {
                    'col-span-4': i === 0,
                    'col-span-8': i === 1,
                    'cursor-pointer': header.column.getCanSort(),
                  }
                )}
                onClick={(e) => {
                  header.column.getToggleSortingHandler()(e)
                }}
              >
                <span>{header.column.columnDef.header as React.ReactNode}</span>
                <span>
                  {header.column.getCanSort() &&
                    {
                      asc: <ArrowUp />,
                      desc: <ArrowDown />,
                    }[header.column.getIsSorted() as string]}
                </span>
              </Text>
            ))}
          </div>
        ))}

        {table.getRowModel().rows.map((row) => (
          <span
            key={row.id}
            className="grid grid-cols-12 items-start justify-start gap-6 py-1 w-full"
          >
            {row.getVisibleCells().map((cell) => (
              <React.Fragment key={cell.id}>
                {flexRender(cell.column.columnDef.cell, cell.getContext())}
              </React.Fragment>
            ))}
          </span>
        ))}
      </div>
    </div>
  )
}
