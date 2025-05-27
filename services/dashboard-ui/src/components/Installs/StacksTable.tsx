'use client'

import React, { type FC, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { CaretRight, CalendarBlank, Minus } from '@phosphor-icons/react'
import { ClickToCopyButton } from '@/components/ClickToCopy'
import { DataTableSearch, Table } from '@/components/DataTable'
import { Link } from '@/components/Link'
import { StatusBadge } from '@/components/Status'
import { Time } from '@/components/Time'
import { ID, Text } from '@/components/Typography'
// eslint-disable-next-line import/no-cycle
import type { TInstallStack } from '@/types'

export interface IStacksTable {
  installId: string
  orgId: string
  stack: TInstallStack
}

export const StacksTable: FC<IStacksTable> = ({ installId, orgId, stack }) => {
  const [data, _] = useState(stack?.versions)
  const [columnFilters, __] = useState([])
  const [globalFilter, setGlobalFilter] = useState('')

  const columns: Array<ColumnDef<TInstallStack['versions'] & { id: string }>> =
    useMemo(
      () => [
        {
          header: 'Stack version',
          accessorKey: 'id',
          cell: (props) => (
            <div className="flex flex-col gap-2">
              <ID className="mt-1" id={props.getValue<string>()} />
            </div>
          ),
        },
        {
          header: 'App version',
          accessorKey: 'app_config_id',
          cell: (props) => (
            <ID className="mt-1" id={props.getValue<string>()} />
          ),
        },
        {
          header: 'Status',
          accessorKey: 'composite_status.status',
          cell: (props) => {
            const status = props.getValue<string>()
            return status ? <StatusBadge status={status} /> : <Minus />
          },
        },
        {
          header: 'Runs',
          accessorKey: 'runs',
          cell: (props) => {
            const runs = props.getValue<Array<any>>()
            return <Text>{runs?.length}</Text>
          },
        },
        {
          header: 'Created',
          accessorKey: 'created_at',
          cell: (props) => (
            <Text className="!items-center">
              <CalendarBlank size="18" />
              <Time format="relative" time={props.getValue<string>()} />
            </Text>
          ),
        },
        {
          accessorKey: 'quick_link_url',
          header: 'Quick link',
          cell: (props) => (
            <Text className="flex flex-nowrap max-w-[160px] items-center gap-2">
              <span className="truncate">{props.getValue<string>()}</span>{' '}
              <ClickToCopyButton textToCopy={props.getValue<string>()} />
            </Text>
          ),
        },
        {
          accessorKey: 'template_url',
          header: 'Template link',
          cell: (props) => (
            <Text className="flex flex-nowrap max-w-[160px] items-center gap-2">
              <span className="truncate">{props.getValue<string>()}</span>{' '}
              <ClickToCopyButton textToCopy={props.getValue<string>()} />
            </Text>
          ),
        },
        /* {
         *   id: 'test',
         *   enableSorting: false,
         *   cell: (props) => (
         *     <Link
         *       href={`/${orgId}/installs/${installId}/stacks/${props.row.original.id}`}
         *       variant="ghost"
         *     >
         *       <CaretRight />
         *     </Link>
         *   ),
         * }, */
      ],
      []
    )

  const handleGlobleFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    setGlobalFilter(e.target.value || '')
  }

  return (
    <Table
      header={
        <>
          <DataTableSearch
            handleOnChange={handleGlobleFilter}
            value={globalFilter}
          />
        </>
      }
      data={data}
      columns={columns}
      columnFilters={columnFilters}
      emptyMessage="Reset your search and try again."
      emptyTitle="No stacks found"
      globalFilter={globalFilter}
    />
  )
}
