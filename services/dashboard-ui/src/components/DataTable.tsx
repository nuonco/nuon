'use client'

import classNames from 'classnames'
import {
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  flexRender,
  useReactTable,
  type ColumnDef,
  type ColumnFilter,
} from '@tanstack/react-table'
import { ArrowDown, ArrowUp, MagnifyingGlass } from '@phosphor-icons/react'
import React, { type FC } from 'react'
import { FiMoreVertical } from 'react-icons/fi'
import { Link } from '@/components'

export interface IDataTable {
  headers: Array<string>
  initData: Array<Array<React.ReactElement | string>>
}

export const DataTable: FC<IDataTable> = ({ headers, initData }) => {
  const data = initData

  return (
    <div className="flex flex-col gap-8">
      <table className="table-auto w-full">
        <thead>
          <tr className="border-b text-left">
            {headers.map((header, i) => (
              <th className="text-sm" key={`header-${i}`}>
                {header}
              </th>
            ))}
            <th></th>
          </tr>
        </thead>
        <tbody className="divide-y">
          {data.map((row, i) => (
            <tr key={`row-${i}`}>
              {row.map((td, i) => (
                <td className="py-2" key={`cell-${i}`}>
                  {i + 1 !== row.length ? (
                    <>{td}</>
                  ) : (
                    <Link
                      className="text-gray-950 dark:text-gray-50"
                      href={td as string}
                    >
                      <FiMoreVertical />
                    </Link>
                  )}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export interface ITable extends React.HTMLAttributes<HTMLTableElement> {
  data: Array<Record<string, any>>
  columns: Array<ColumnDef<any>>
  columnFilters: Array<ColumnFilter>
  globalFilter: string
  header?: React.ReactNode
}

export const Table: FC<ITable> = ({
  data,
  columns,
  columnFilters,
  globalFilter,
  header,
  ...props
}) => {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    state: { columnFilters, globalFilter },
  })

  return (
    <div className="flex flex-col gap-8">
      {header && (
        <div className="flex items-center justify-between w-full">{header}</div>
      )}

      <table className="table-auto w-full" {...props}>
        <thead>
          {table.getHeaderGroups().map((group) => (
            <tr className="border-b text-left" key={group.id}>
              {group.headers.map((header) => (
                <th
                  className={classNames(
                    'text-base font-medium leading-normal p-4 text-cool-grey-600 dark:text-cool-grey-500',
                    {
                      'cursor-pointer': header.column.getCanSort(),
                    }
                  )}
                  key={header.id}
                  onClick={(e) => {
                    header.column.getToggleSortingHandler()(e)
                  }}
                >
                  <div className="flex items-center gap-4">
                    <span>
                      {header.column.columnDef.header as React.ReactNode}
                    </span>
                    <span>
                      {header.column.getCanSort() &&
                        {
                          asc: <ArrowUp />,
                          desc: <ArrowDown />,
                        }[header.column.getIsSorted() as string]}
                    </span>
                  </div>
                </th>
              ))}
            </tr>
          ))}
        </thead>
        <tbody className="divide-y">
          {table.getRowModel().rows.map((row) => (
            <tr key={row.id}>
              {row.getVisibleCells().map((cell, i) => (
                <td
                  className={classNames('p-4', {
                    'align-top': row.getVisibleCells().length !== i + 1,
                    'align-center': row.getVisibleCells().length === i + 1,
                  })}
                  key={cell.id}
                >
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export const DataTableSearch: FC<{
  value?: string
  handleOnChange: any
}> = ({ value = '', handleOnChange }) => {
  return (
    <label className="relative">
      <MagnifyingGlass className="text-cool-grey-600 dark:text-cool-grey-500 absolute top-2.5 left-2" />
      <input
        className="rounded-md pl-8 pr-3.5 py-1.5 text-base border bg-white dark:bg-dark-grey-100 placeholder:text-cool-grey-600 dark:placeholder:text-cool-grey-500 md:min-w-80"
        type="search"
        placeholder="Search..."
        value={value}
        onChange={handleOnChange}
      />
    </label>
  )
}
