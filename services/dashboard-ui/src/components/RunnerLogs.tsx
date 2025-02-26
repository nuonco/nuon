'use client'

import classNames from 'classnames'
import React, { type FC, useMemo, useState } from 'react'
import { DateTime } from 'luxon'
import { ArrowDown, ArrowUp, ArrowsOutSimple } from '@phosphor-icons/react'
import {
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  flexRender,
  useReactTable,
  type ColumnDef,
  type ColumnFilter,
  type ColumnSort,
} from '@tanstack/react-table'
import { Button } from '@/components/Button'
import { ClickToCopy } from '@/components/ClickToCopy'
import { Expand } from '@/components/Expand'
import { LogsPreview } from '@/components/RunnerLogsPreview'
import { LogLineSeverity } from '@/components/RunnerLogLineSeverity'
import { Modal } from '@/components/Modal'
import { Section } from '@/components/Card'
import { RunnerLogsActions } from '@/components/RunnerLogsActions'
import { Time } from '@/components/Time'
import { Code, Text } from '@/components/Typography'
import type { TOTELLog } from '@/types'

interface IOTELLogs {
  data: Array<Record<string, any>>
  columns: Array<ColumnDef<any>>
  columnFilters: Array<ColumnFilter>
  globalFilter: string
  sorting: Array<ColumnSort>
  isAllExpanded?: boolean
}

export const OTELLogs: FC<IOTELLogs> = ({
  data,
  columns,
  columnFilters,
  globalFilter,
  isAllExpanded = false,
  sorting,
}) => {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    state: { columnFilters, globalFilter, sorting },
  })

  return (
    <div className="divide-y">
      {table.getHeaderGroups().map((group) => (
        <div
          key={`header-${group.id}`}
          className="grid grid-cols-12 items-center justify-start gap-5 py-2 w-[calc(100%-20px)]"
        >
          {group.headers.map((header, i) => (
            <Text
              key={header.id}
              className={classNames(
                '!font-medium text-cool-grey-600 dark:text-cool-grey-500',
                {
                  'col-span-1': i === 0 || i === 2,
                  'col-span-2': i === 1,
                  'col-span-8': i === 3,
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

      {table.getRowModel().rows.map((row) => {
        const logAttributes = row.original.log_attributes
        const resourceAttributes = row.original.resource_attributes

        return (
          <Expand
            key={row.original?.id}
            id={row.original?.id}
            headerClass={classNames({
              'bg-primary-900/5 dark:bg-primary-400/5':
                row.original?.service_name === 'api',
            })}
            className={classNames(
              'grid grid-cols-12 items-start justify-start gap-6 py-2 w-full'
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
                      <Text className="text-base !font-medium leading-normal p-4">
                        Log attributes
                      </Text>
                    }
                    expandContent={
                      <div className="divide-y p-4 bg-black/5 dark:bg-white/5">
                        <div className="grid grid-cols-3 gap-4 pb-3">
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
                            className="grid grid-cols-3 gap-4 py-3"
                          >
                            <Text className="font-mono text-sm break-all !inline truncate max-w-[250px]">
                              {key}
                            </Text>

                            <span className="col-span-2">
                              {key === 'intermediate-data' ? (
                                <Code variant="preformated">
                                  <ClickToCopy insetNotice>
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
                    <Text className="text-base !font-medium leading-normal p-4">
                      Resource attributes
                    </Text>
                  }
                  expandContent={
                    <div className="divide-y p-4 bg-black/5 dark:bg-white/5">
                      <div className="grid grid-cols-3 gap-4 pb-3">
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
                          className="grid grid-cols-3 gap-4 py-3"
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
        )
      })}
    </div>
  )
}

export type TLogRecord = Omit<TOTELLog, 'timestamp'> & { timestamp: number }

// convert otel log timestamp from string to milliseconds
export function parseOTELLog(logs: Array<TOTELLog>): Array<TLogRecord> {
  return logs?.length
    ? logs?.map((l) => ({
        ...l,
        timestamp: DateTime.fromISO(l.timestamp).toMillis(),
      }))
    : []
}

export interface IRunnerLogs {
  heading: React.ReactNode
  actions?: React.ReactNode
  logs: Array<TLogRecord>
  withOutBorder?: boolean
}

export const RunnerLogs: FC<IRunnerLogs> = ({
  heading,
  actions,
  logs,
  withOutBorder,
}) => {
  const [isDetailsOpen, setIsDetailsOpen] = useState<boolean>(false)
  const [columnFilters, setColumnFilters] = useState([
    {
      id: 'severity_text',
      value: ['Trace', 'Debug', 'Info', 'Warn', 'Error', 'Fatal'],
    },
  ])
  const [globalFilter, setGlobalFilter] = useState('')
  const [columnSort, setColumnSort] = useState([
    { id: 'timestamp', desc: true },
  ])
  const [isAllExpanded, setIsAllExpanded] = useState(false)
  const lineStyle =
    'text-sm font-mono text-cool-grey-600 dark:text-cool-grey-500'

  const columns: Array<ColumnDef<TOTELLog>> = useMemo(
    () => [
      {
        header: 'Severity',
        accessorKey: 'severity_text',
        cell: (props) => (
          <span className={classNames('flex items-center gap-2')}>
            <LogLineSeverity
              severity_number={props.row.original?.severity_number}
            />
            <span className={lineStyle + ' font-semibold uppercase'}>
              {props.getValue<string>() || 'UNKOWN'}
            </span>
          </span>
        ),
        enableColumFilter: true,
        filterFn: 'arrIncludesSome',
      },
      {
        header: 'Date',
        accessorKey: 'timestamp',
        cell: (props) => (
          <span
            className={classNames(lineStyle, {
              'col-span-2 flex items-center gap-2': true,
            })}
          >
            <Time
              className="!text-sm"
              time={DateTime.fromMillis(props.getValue<number>()).toISO()}
            />
          </span>
        ),
      },
      {
        header: 'Service',
        accessorKey: 'service_name',
        enableColumSort: false,
        cell: (props) => (
          <span
            className={classNames(lineStyle, {
              'col-span-1 flex items-center': true,
            })}
          >
            <span>{props.getValue<string>()}</span>
          </span>
        ),
      },
      {
        header: 'Content',
        accessorKey: 'body',
        cell: (props) => (
          <span
            className={classNames(lineStyle, {
              'col-span-8 flex items-center': true,
            })}
          >
            <span>{props.getValue<string>()}</span>
          </span>
        ),
      },
    ],
    []
  )

  const handleStatusFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { checked, value } = e.target
    setColumnFilters((state) => {
      const values = [...state?.at(0)?.value]
      const index = values?.indexOf(value)
      if (checked && index < 0) {
        values.push(value)
      } else if (index > -1) {
        values.splice(index, 1)
      }
      return [{ id: 'severity_text', value: values }]
    })
  }

  const handleStatusOnlyFilter = (e) => {
    setColumnFilters([
      { id: 'severity_text', value: [e?.currentTarget?.value] },
    ])
  }

  const clearStatusFilter = () => {
    setColumnFilters([
      {
        id: 'severity_text',
        value: ['Trace', 'Debug', 'Info', 'Warn', 'Error', 'Fatal'],
      },
    ])
  }

  const handleGlobleFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    setGlobalFilter(e.target.value)
  }

  const handleColumnSort = () => {
    setColumnSort([{ id: 'timestamp', desc: !columnSort?.[0].desc }])
  }

  const handleExpandAll = () => {
    setIsAllExpanded(!isAllExpanded)
  }

  return (
    <>
      <Modal
        actions={
          <RunnerLogsActions
            columnSort={columnSort}
            columnFilters={columnFilters}
            globalFilter={globalFilter}
            handleGlobalFilter={handleGlobleFilter}
            handleStatusFilter={handleStatusFilter}
            handleStatusOnlyFilter={handleStatusOnlyFilter}
            handleColumnSort={handleColumnSort}
            handleExpandAll={handleExpandAll}
            clearStatusFilter={clearStatusFilter}
            isAllExpanded={isAllExpanded}
            shouldShowExpandAll
            id="modal"
          />
        }
        hasFixedHeight
        heading={heading}
        isOpen={isDetailsOpen}
        onClose={() => {
          setIsDetailsOpen(false)
        }}
      >
        <OTELLogs
          data={logs}
          columns={columns}
          columnFilters={columnFilters}
          globalFilter={globalFilter}
          sorting={columnSort}
          isAllExpanded={isAllExpanded}
        />
      </Modal>
      <Section
        className={classNames({
          'border-r': !withOutBorder,
        })}
        isHeadingFixed
        actions={
          <div className="flex items-center divide-x">
            {logs?.length > 0 ? (
              <>
                {actions ? <div className="pr-4">{actions}</div> : null}
                <RunnerLogsActions
                  columnSort={columnSort}
                  columnFilters={columnFilters}
                  globalFilter={globalFilter}
                  handleGlobalFilter={handleGlobleFilter}
                  handleStatusOnlyFilter={handleStatusOnlyFilter}
                  handleStatusFilter={handleStatusFilter}
                  handleColumnSort={handleColumnSort}
                  id="preview"
                  shouldHideFilter
                />
                <div className="ml-4 pl-4">
                  <Button
                    className="flex items-center gap-2 text-sm !font-medium"
                    onClick={() => {
                      setIsDetailsOpen(true)
                    }}
                  >
                    <ArrowsOutSimple size="14" />
                    View all logs
                  </Button>
                </div>
              </>
            ) : null}
          </div>
        }
        heading={heading}
      >
        {logs?.length ? (
          <LogsPreview
            data={logs}
            globalFilter={globalFilter}
            sorting={columnSort}
          />
        ) : (
          <Text className="text-base">No logs found</Text>
        )}
      </Section>
    </>
  )
}
