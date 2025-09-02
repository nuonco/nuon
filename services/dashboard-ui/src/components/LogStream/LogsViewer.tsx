'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { ArrowDown, ArrowUp } from '@phosphor-icons/react'
import {
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  flexRender,
  useReactTable,
  type ColumnDef,
} from '@tanstack/react-table'
import { useLogs } from './logs-context'
import { useLogsViewer } from './logs-viewer-context'
import type { TLogRecord } from './types'
import { ClickToCopy } from '@/components/ClickToCopy'
import { Expand } from '@/components/Expand'
import { SpinnerSVG } from '@/components/Loading'
import { Code, Text } from '@/components/Typography'

interface ILogsViewer {
  data: Array<TLogRecord>
  columns: Array<ColumnDef<any>>
  showLogAttr?: boolean
  enableLogFilter?: boolean
}

export const LogsViewer: FC<ILogsViewer> = ({
  data,
  columns,
  showLogAttr = false,
  enableLogFilter = false,
}) => {
  const { isLoading, isPolling } = useLogs()
  const { columnFilters, columnSort, globalFilter, isAllExpanded } =
    useLogsViewer()
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    state: {
      columnFilters: enableLogFilter ? columnFilters : undefined,
      globalFilter,
      sorting: columnSort,
    },
  })

  return (
    <div className="divide-y">
      {table.getHeaderGroups().map((group) => (
        <div
          key={`header-${group.id}`}
          className="grid grid-cols-12 items-center justify-start gap-3 py-2 w-[calc(100%-20px)]"
        >
          {group.headers.map((header, i) => (
            <Text
              key={header.id}
              className={classNames(
                '!font-medium text-cool-grey-600 dark:text-cool-grey-500',
                showLogAttr
                  ? {
                      'col-span-1': i === 0 || i === 2,
                      'col-span-3': i === 1,
                      'col-span-7': i === 3,
                    }
                  : {
                      'col-span-4': i === 0,
                      'col-span-8': i === 1,
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

      {data?.length && isLoading && !isPolling ? (
        <div className="w-full py-2 flex on-enter">
          <Text
            className="flex items-center gap-3 m-auto text-cool-grey-600 dark:text-white/70"
            variant="reg-14"
          >
            <SpinnerSVG /> Loading more log lines...
          </Text>
        </div>
      ) : null}

      {table.getRowModel().rows.map((row) => {
        const logAttributes = row.original.log_attributes
        const resourceAttributes = row.original.resource_attributes

        return showLogAttr ? (
          <Expand
            key={row.original?.id}
            id={row.original?.id}
            headerClass={classNames({
              'bg-primary-900/5 dark:bg-primary-400/5':
                row.original?.service_name === 'api',
            })}
            className={classNames(
              'grid grid-cols-12 items-start justify-start gap-4 py-1 w-full'
            )}
            heading={
              row
                .getVisibleCells()
                .map((cell) => (
                  <React.Fragment key={cell.id}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </React.Fragment>
                )) as unknown as React.ReactElement
            }
            expandContent={
              <div className="flex flex-col bg-black/5 dark:bg-white/5">
                {Object.keys(logAttributes)?.length ? (
                  <Expand
                    id={`${row.original?.id}-log-attr`}
                    heading={
                      <Text className="text-base !font-medium leading-normal px-4 py-2">
                        Log attributes
                      </Text>
                    }
                    expandContent={
                      <div className="divide-y px-4 py-2 bg-black/5 dark:bg-white/5">
                        <div className="grid grid-cols-3 gap-4 pb-2">
                          <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
                            Key
                          </Text>
                          <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
                            Value
                          </Text>
                        </div>

                        {Object.keys(logAttributes).map((key) => (
                          <div
                            key={`${key}-${row.original?.id}`}
                            className="grid grid-cols-3 gap-4 py-2"
                          >
                            <Text className="font-mono text-sm break-all !inline truncate max-w-[250px]">
                              {key}
                            </Text>

                            <span className="col-span-2">
                              {key === 'intermediate-data' ||
                              key === 'diagnostic' ||
                              key === 'change' ||
                              key === 'changes' ||
                              key === 'outputs' ||
                              key === 'hook' ? (
                                <Code variant="preformated">
                                  <ClickToCopy
                                    className="!items-start justify-between"
                                    noticeClassName="-top-1 right-5"
                                  >
                                    {JSON.stringify(
                                      JSON.parse(logAttributes[key]),
                                      null,
                                      2
                                    )}
                                  </ClickToCopy>
                                </Code>
                              ) : (
                                <Text className="text-sm font-mono text-pretty!inline break-all">
                                  {logAttributes[key]}
                                </Text>
                              )}
                            </span>
                          </div>
                        ))}
                      </div>
                    }
                    isOpen
                  />
                ) : null}

                <Expand
                  id={`${row.original?.id}-resource-attr`}
                  heading={
                    <Text className="text-base !font-medium leading-normal px-4 py-2">
                      Resource attributes
                    </Text>
                  }
                  expandContent={
                    <div className="divide-y px-4 py-2 bg-black/5 dark:bg-white/5">
                      <div className="grid grid-cols-3 gap-4 pb-2">
                        <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
                          Key
                        </Text>
                        <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
                          Value
                        </Text>
                      </div>

                      {Object.keys(resourceAttributes).map((key) => (
                        <div
                          key={`${key}-${row.original?.id}`}
                          className="grid grid-cols-3 gap-4 py-2"
                        >
                          <Text className="font-mono text-sm break-all !inline truncate max-w-[250px]">
                            {key}
                          </Text>
                          <Text className="text-sm font-mono text-pretty col-span-2 !inline break-all">
                            {resourceAttributes[key]}
                          </Text>
                        </div>
                      ))}
                    </div>
                  }
                />
              </div>
            }
            isOpen={isAllExpanded}
          />
        ) : (
          <span
            key={row.id}
            className="grid grid-cols-12 items-start justify-start gap-4 py-1 w-full"
          >
            {row.getVisibleCells().map((cell) => (
              <React.Fragment key={cell.id}>
                {flexRender(cell.column.columnDef.cell, cell.getContext())}
              </React.Fragment>
            ))}
          </span>
        )
      })}
    </div>
  )
}
