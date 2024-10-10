'use client'

import classNames from 'classnames'
import React, { type FC, useMemo, useState } from 'react'
import {
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
} from '@tanstack/react-table'
import {
  Button,
  Dropdown,
  Text,
  Time,
  Section,
  Modal,
  LogsPreview,
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

export const OTELLogs: FC<{ logs?: Array<TOTELLog> }> = ({ logs = [] }) => {
  const [data, _] = useState(logs)
  const [columnFilters, __] = useState([])
  const [globalFilter, setGlobalFilter] = useState('')
  const lineStyle =
    'tracking-wider text-sm font-mono leading-loose text-cool-grey-600 dark:text-cool-grey-500'

  const columns = useMemo(
    () => [
      {
        header: 'Severity',
        accessorKey: 'severity_number',
        cell: (props) => (
          <span className={classNames('flex items-center gap-2')}>
            <LogLineSeverity severity_number={props.getValue()} />
            <span className={lineStyle + ' font-semibold uppercase'}>
              {props.row.original?.severity_text || 'UNKOWN'}
            </span>
          </span>
        ),
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
            <Time className="!text-sm" time={props.getValue()} />
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
            {props.getValue()}
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
            {props.getValue()}
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
    state: { columnFilters, globalFilter },
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
                }
              )}
              onClick={(e) => {
                header.column.getToggleSortingHandler()(e)
              }}
            >
              {header.column.columnDef.header as React.ReactNode}
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

  return (
    <>
      <Modal
        actions={<RunnerLogsActions />}
        heading={heading}
        isOpen={isDetailsOpen}
        onClose={() => {
          setIsDetailsOpen(false)
        }}
      >
        <OTELLogs logs={logs} />
      </Modal>
      <Section
        className="border-r"
        actions={
          <div className="flex items-center divide-x">
            <div className="pl-4">
              <Button
                className="flex items-center gap-2 text-base !font-medium"
                onClick={() => {
                  setIsDetailsOpen(true)
                }}
              >
                <ArrowsOutSimple />
                Open logs
              </Button>
            </div>
          </div>
        }
        heading={heading}
      >
        {logs?.length ? (
          <LogsPreview logs={logs} />
        ) : (
          <Text className="text-base">No logs found</Text>
        )}
      </Section>
    </>
  )
}

const RunnerLogsActions: FC = () => {
  return (
    <div className="flex items-center gap-4">
      <Button
        className="text-base !font-medium !p-2 w-[32px] h-[32px]"
        variant="ghost"
      >
        <MagnifyingGlass />
      </Button>

      <Button
        className="text-base !font-medium !p-2 w-[32px] h-[32px]"
        variant="ghost"
      >
        <FunnelSimple />
      </Button>

      <Button
        className="text-base !font-medium !p-2 w-[32px] h-[32px]"
        variant="ghost"
      >
        <Funnel />
      </Button>
    </div>
  )
}
