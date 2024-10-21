'use client'

import classNames from 'classnames'
import React, { type FC, useMemo, useState } from 'react'
import {
  ArrowDown,
  ArrowUp,
  ArrowsOutSimple,
  ArrowsInLineVertical,
  ArrowsOutLineVertical,
  CaretUpDown,
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
  CodeViewer,
  Dropdown,
  Expand,
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
            className="grid grid-cols-12 items-center justify-start gap-6 py-2 w-full border-t"
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
              <div className="flex flex-col gap-6 p-4 bg-black/5 dark:bg-white/5">
                <div className="flex flex-col gap-4">
                  <Text className="text-base !font-medium leading-normal">
                    Log attributes
                  </Text>
                  <div className="divide-y">
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
                        <Text className="text-sm font-mono break-all col-span-2 !inline truncate max-w-[200px]">
                          {logAttributes[key]}
                        </Text>
                      </div>
                    ))}
                  </div>
                </div>
                <div className="flex flex-col gap-4">
                  <Text className="text-base !font-medium leading-normal">
                    Resource attributes
                  </Text>
                  <div className="divide-y">
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
                        <Text className="text-sm font-mono break-all col-span-2 !inline truncate max-w-[200px]">
                          {resourceAttributes[key]}
                        </Text>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            }
            isOpen={isAllExpanded}
          />
        )
      })}
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
              'col-span-7 flex items-center justify-between': true,
            })}
          >
            <span>{props.getValue<string>()}</span>
            <CaretUpDown className="mr-2" />
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
            data={logs}
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
  handleExpandAll?: any
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
  isAllExpanded = false,
  id,
  shouldHideFilter = false,
  shouldShowExpandAll = false,
}) => {
  return (
    <div className="flex items-center gap-4">
      {false && (
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
