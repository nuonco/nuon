'use client'

import classNames from 'classnames'
import React, { type FC, useMemo, useState } from 'react'
import {
  ArrowDown,
  ArrowUp,
  ArrowsOutSimple,
  MagnifyingGlass,
  FunnelSimple,
  Funnel,
} from '@phosphor-icons/react'
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
import {
  Button,
  Dropdown,
  Text,
  Time,
  Section,
  Modal,
  LogsPreview,
  RadioInput,
} from '@/components'
import type { TOTELLog } from '@/types'

export const LogLineSeverity: FC<{ severity_number: number }> = ({
  severity_number,
}) => {
  return (
    <span
      className={classNames('flex w-0.5 h-3', {
        'bg-primary-400 dark:bg-primary-300': severity_number <= 4,
        'bg-cool-grey-600 dark:bg-cool-grey-500':
          severity_number >= 5 && severity_number <= 8,
        'bg-blue-600 dark:bg-blue-500':
          severity_number >= 9 && severity_number <= 12,
        'bg-orange-600 dark:bg-orange-500':
          severity_number >= 13 && severity_number <= 16,
        'bg-red-600 dark:bg-red-500':
          severity_number >= 17 && severity_number <= 20,
        'bg-red-700 dark:bg-red-600':
          severity_number >= 21 && severity_number <= 24,
      })}
    />
  )
}

interface IOTELLogs {
  data: Array<Record<string, any>>
  columns: Array<ColumnDef<any>>
  columnFilters: Array<ColumnFilter>
  globalFilter: string
  sorting: Array<ColumnSort>
}

export const OTELLogs: FC<IOTELLogs> = ({
  data,
  columns,
  columnFilters,
  globalFilter,
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
          key={group.id}
          className="grid grid-cols-12 items-center justify-start gap-6 py-2 w-full"
        >
          {group.headers.map((header, i) => (
            <Text
              key={header.id}
              className={classNames(
                '!font-medium text-cool-grey-600 dark:text-cool-grey-500',
                {
                  'col-span-1': i === 0,
                  'col-span-2': i === 1 || i === 2,
                  'col-span-7': i === 3,
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
          className="grid grid-cols-12 items-center justify-start gap-6 py-2 w-full"
        >
          {row
            .getVisibleCells()
            .map((cell) =>
              flexRender(cell.column.columnDef.cell, cell.getContext())
            )}
        </span>
      ))}
    </div>
  )
}

export interface IRunnerLogs {
  heading: React.ReactNode
  logs: Array<TOTELLog>
}

export const RunnerLogs: FC<IRunnerLogs> = ({ heading, logs }) => {
  const [isDetailsOpen, setIsDetailsOpen] = useState<boolean>(false)
  const [data, _] = useState(logs)
  const [columnFilters, setColumnFilters] = useState([])
  const [globalFilter, setGlobalFilter] = useState('')
  const [columnSort, setColumnSort] = useState([
    { id: 'timestamp', desc: false },
  ])
  const lineStyle =
    'tracking-wider text-sm font-mono leading-loose text-cool-grey-600 dark:text-cool-grey-500'

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
        filterFn: (row, columnId, filterValue) => {
          const severityText = row.getValue<string>(columnId)
          return severityText === filterValue
        },
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
            <Time className="!text-sm" time={props.getValue<string>()} />
          </span>
        ),
      },
      {
        header: 'Service',
        accessorKey: 'service_name',
        cell: (props) => (
          <span
            className={classNames(lineStyle, {
              'col-span-2': true,
            })}
          >
            {props.getValue<string>()}
          </span>
        ),
      },
      {
        header: 'Content',
        accessorKey: 'body',
        cell: (props) => (
          <span
            className={classNames(lineStyle, {
              'col-span-7': true,
            })}
          >
            {props.getValue<string>()}
          </span>
        ),
      },
    ],
    []
  )

  const handleStatusFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    setColumnFilters(() => [{ id: 'severity_text', value: value }])
  }

  const handleGlobleFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    setGlobalFilter(e.target.value)
  }

  const handleColumnSort = (e: React.ChangeEvent<HTMLInputElement>) => {
    setColumnSort([
      { id: 'timestamp', desc: Boolean(e.target.value === 'true') },
    ])
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
            handleColumnSort={handleColumnSort}
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
          data={data}
          columns={columns}
          columnFilters={columnFilters}
          globalFilter={globalFilter}
          sorting={columnSort}
        />
      </Modal>
      <Section
        className="border-r"
        actions={
          <div className="flex items-center divide-x">
            {logs?.length > 0 ? (
              <>
                <RunnerLogsActions
                  columnSort={columnSort}
                  columnFilters={columnFilters}
                  globalFilter={globalFilter}
                  handleGlobalFilter={handleGlobleFilter}
                  handleStatusFilter={handleStatusFilter}
                  handleColumnSort={handleColumnSort}
                  id="preview"
                  shouldHideFilter
                />
                <div className="ml-4 pl-4">
                  <Button
                    className="flex items-center gap-2 text-base !font-medium"
                    onClick={() => {
                      setIsDetailsOpen(true)
                    }}
                  >
                    <ArrowsOutSimple />
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
            data={data}
            columnFilters={columnFilters}
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

interface IRunnerLogsActions {
  columnFilters: any
  columnSort: any
  globalFilter: string
  handleStatusFilter: any
  handleGlobalFilter: any
  handleColumnSort: any
  id: string
  shouldHideFilter?: boolean
}

const RunnerLogsActions: FC<IRunnerLogsActions> = ({
  columnSort,
  globalFilter,
  handleGlobalFilter,
  handleStatusFilter,
  handleColumnSort,
  id,
  shouldHideFilter = false,
}) => {
  return (
    <div className="flex items-center gap-4">
      <Dropdown
        alignment="right"
        className="text-base !font-medium !p-2 w-[32px] h-[32px]"
        variant="ghost"
        id="logs-search"
        text={<MagnifyingGlass />}
      >
        <div>
          <label className="relative">
            <MagnifyingGlass className="text-cool-grey-600 dark:text-cool-grey-500 absolute top-0.5 left-2" />
            <input
              className="rounded-md pl-8 pr-3.5 py-1.5 text-base border bg-white dark:bg-dark-grey-100 placeholder:text-cool-grey-600 dark:placeholder:text-cool-grey-500 md:min-w-80"
              type="search"
              placeholder="Search..."
              value={globalFilter}
              onChange={handleGlobalFilter}
            />
          </label>
        </div>
      </Dropdown>

      <Dropdown
        alignment="right"
        className="text-base !font-medium !p-2 w-[32px] h-[32px]"
        variant="ghost"
        id="sort-logs"
        text={<FunnelSimple />}
      >
        <div>
          <RadioInput
            name={`${id}-column-sort`}
            checked={columnSort?.[0]?.desc}
            onChange={handleColumnSort}
            value="true"
            labelText="Newest"
          />

          <RadioInput
            name={`${id}-column-sort`}
            checked={!columnSort?.[0]?.desc}
            onChange={handleColumnSort}
            value="false"
            labelText="Oldest"
          />
        </div>
      </Dropdown>

      {shouldHideFilter ? null : (
        <Dropdown
          alignment="right"
          className="text-base !font-medium !p-2 w-[32px] h-[32px]"
          variant="ghost"
          id="logs-filter"
          text={<Funnel />}
        >
          <div>
            <RadioInput
              name={`${id}-status-filter`}
              onChange={handleStatusFilter}
              value="Trace"
              labelText="Trace"
            />

            <RadioInput
              name={`${id}-status-filter`}
              onChange={handleStatusFilter}
              value="Debug"
              labelText="Debug"
            />

            <RadioInput
              name={`${id}-status-filter`}
              onChange={handleStatusFilter}
              value="Info"
              labelText="Info"
            />

            <RadioInput
              name={`${id}-status-filter`}
              onChange={handleStatusFilter}
              value="Warn"
              labelText="Warning"
            />

            <RadioInput
              name={`${id}-status-filter`}
              onChange={handleStatusFilter}
              value="Error"
              labelText="Error"
            />

            <RadioInput
              name={`${id}-status-filter`}
              onChange={handleStatusFilter}
              value="Fatal"
              labelText="Fatal"
            />
          </div>
        </Dropdown>
      )}
    </div>
  )
}
