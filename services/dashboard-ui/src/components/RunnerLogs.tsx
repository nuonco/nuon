'use client'

import classNames from 'classnames'
import React, { type FC, useMemo, useState } from 'react'
import { DateTime } from 'luxon'
import {
  ArrowDown,
  ArrowUp,
  ArrowsOutSimple,
  ArrowsInLineVertical,
  ArrowsOutLineVertical,
  MagnifyingGlass,
  Funnel,
  SortAscending,
  SortDescending,
  X,
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
import { Button } from '@/components/Button'
import { Dropdown } from '@/components/Dropdown'
import { Expand } from '@/components/Expand'
import { LogsPreview } from '@/components/RunnerLogsPreview'
import { LogLineSeverity } from '@/components/RunnerLogLineSeverity'
import { Modal } from '@/components/Modal'
import { Section } from '@/components/Card'
import { RadioInput } from '@/components/Input'
import { Time } from '@/components/Time'
import { Text } from '@/components/Typography'
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

      {table.getRowModel().rows.map((row) => {
        const logAttributes = row.original.log_attributes
        const resourceAttributes = row.original.resource_attributes

        return (
          <Expand
            key={row.id}
            id={row.id}
            className="grid grid-cols-12 items-center justify-start gap-6 py-2 w-full"
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
                <Expand
                  id={`${row.id}-log-attr`}
                  heading={
                    <Text className="text-base !font-medium leading-normal p-4">
                      Log attributes
                    </Text>
                  }
                  expandContent={
                    <div className="divide-y p-4">
                      <div className="grid grid-cols-3 gap-4 pb-3">
                        <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
                          Key
                        </Text>
                        <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
                          Value
                        </Text>
                      </div>

                      {Object.keys(logAttributes).map((key, i) => (
                        <div
                          key={`${key}-${i}`}
                          className="grid grid-cols-3 gap-4 py-3"
                        >
                          <Text className="font-mono text-sm break-all !inline truncate max-w-[200px]">
                            {key}
                          </Text>

                          <Text className="text-sm font-mono text-pretty col-span-2 !inline">
                            {logAttributes[key]}
                          </Text>
                        </div>
                      ))}
                    </div>
                  }
                  isOpen
                />

                <Expand
                  id={`${row.id}-resource-attr`}
                  heading={
                    <Text className="text-base !font-medium leading-normal p-4">
                      Resource attributes
                    </Text>
                  }
                  expandContent={
                    <div className="divide-y p-4">
                      <div className="grid grid-cols-3 gap-4 pb-3">
                        <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
                          Key
                        </Text>
                        <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
                          Value
                        </Text>
                      </div>

                      {Object.keys(resourceAttributes).map((key, i) => (
                        <div
                          key={`${key}-${i}`}
                          className="grid grid-cols-3 gap-4 py-3"
                        >
                          <Text className="font-mono text-sm break-all !inline truncate max-w-[200px]">
                            {key}
                          </Text>
                          <Text className="text-sm font-mono text-pretty col-span-2 !inline">
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

function parseOTELLog(logs: Array<TOTELLog>) {
  return logs.map((l) => ({
    ...l,
    timestamp: DateTime.fromISO(l.timestamp).toMillis(),
  }))
}

export interface IRunnerLogs {
  heading: React.ReactNode
  logs: Array<TOTELLog>
}

export const RunnerLogs: FC<IRunnerLogs> = ({ heading, logs }) => {
  const [isDetailsOpen, setIsDetailsOpen] = useState<boolean>(false)
  const [data, _] = useState(parseOTELLog(logs))
  const [columnFilters, setColumnFilters] = useState([])
  const [globalFilter, setGlobalFilter] = useState('')
  const [columnSort, setColumnSort] = useState([
    { id: 'timestamp', desc: false },
  ])
  const [isAllExpanded, setIsAllExpanded] = useState(false)
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
              'col-span-7 flex items-center justify-between': true,
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
    const { value } = e.target
    setColumnFilters(() => [{ id: 'severity_text', value: value }])
  }

  const clearStatusFilter = () => {
    setColumnFilters(() => [])
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
          data={data}
          columns={columns}
          columnFilters={columnFilters}
          globalFilter={globalFilter}
          sorting={columnSort}
          isAllExpanded={isAllExpanded}
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
  handleExpandAll?: any
  clearStatusFilter?: any
  isAllExpanded?: boolean
  id: string
  shouldHideFilter?: boolean
  shouldShowExpandAll?: boolean
}

const RunnerLogsActions: FC<IRunnerLogsActions> = ({
  columnSort,
  globalFilter,
  handleGlobalFilter,
  handleStatusFilter,
  handleColumnSort,
  handleExpandAll,
  clearStatusFilter,
  isAllExpanded = false,
  id,
  shouldHideFilter = false,
  shouldShowExpandAll = false,
}) => {
  return (
    <div className="flex items-center gap-4">
      {shouldShowExpandAll && (
        <Button
          className="text-base !font-medium !p-2 w-[32px] h-[32px]"
          variant="ghost"
          title={
            isAllExpanded ? 'Collapse all log lines' : 'Expand all log lines'
          }
          onClick={handleExpandAll}
        >
          {isAllExpanded ? <ArrowsInLineVertical /> : <ArrowsOutLineVertical />}
        </Button>
      )}
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

      <Button
        className="text-base !font-medium !p-2 w-[32px] h-[32px]"
        variant="ghost"
        title={columnSort?.[0].desc ? 'Sort by oldest' : 'Sort by newest'}
        onClick={handleColumnSort}
      >
        {columnSort?.[0].desc ? <SortAscending /> : <SortDescending />}
      </Button>

      {shouldHideFilter ? null : (
        <Dropdown
          alignment="right"
          className="text-base !font-medium !p-2 w-[32px] h-[32px]"
          variant="ghost"
          id="logs-filter"
          text={<Funnel />}
        >
          <div>
            <form>
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
              <hr />
              <Button
                className="w-full !rounded-t-none !text-sm flex items-center gap-2"
                type="reset"
                onClick={clearStatusFilter}
                variant="ghost"
              >
                <X />
                Clear
              </Button>
            </form>
          </div>
        </Dropdown>
      )}
    </div>
  )
}
