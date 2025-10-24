'use client'

import React, { type FC, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import {
  CalendarBlankIcon,
  MinusIcon,
  DotsThreeVerticalIcon,
} from '@phosphor-icons/react'
import { DataTableSearch, Table } from '@/components/old/DataTable'
import { Dropdown } from '@/components/old/Dropdown'
import { StatusBadge } from '@/components/old/Status'
import { Time } from '@/components/old/Time'
import { ID, Text } from '@/components/old/Typography'
import { StackLinksModal } from './StackLinksModal'
import { StackOutputsModal } from './StackOutputsModal'
// eslint-disable-next-line import/no-cycle
import type { TInstallStack, TInstallStackVersion } from '@/types'

export interface IStacksTable {
  stack: TInstallStack
}

export const StacksTable: FC<IStacksTable> = ({ stack }) => {
  const [data, _] = useState(stack?.versions)
  const [columnFilters, __] = useState([])
  const [globalFilter, setGlobalFilter] = useState('')

  const columns: Array<ColumnDef<TInstallStackVersion>> = useMemo(
    () => [
      {
        header: 'Stack version',
        accessorKey: 'id',
        cell: (props) => <ID className="mt-1" id={props.getValue<string>()} />,
      },
      {
        header: 'App version',
        accessorKey: 'app_config_id',
        cell: (props) => <ID className="mt-1" id={props.getValue<string>()} />,
      },
      {
        header: 'Status',
        accessorKey: 'composite_status.status',
        cell: (props) => {
          const status = props.getValue<string>()
          return status ? <StatusBadge status={status} /> : <MinusIcon />
        },
      },
      {
        header: 'Runs',
        accessorKey: 'runs',
        id: 'run_count',
        cell: (props) => {
          const runs = props.getValue<Array<any>>()
          return <Text>{runs?.length}</Text>
        },
      },
      {
        header: 'Created',
        accessorKey: 'created_at',
        cell: (props) => (
          <Text className="!flex truncate">
            <CalendarBlankIcon size="16" />
            <Time format="relative" time={props.getValue<string>()} />
          </Text>
        ),
      },
      {
        id: 'more',
        enableSorting: false,
        cell: (props) => {
          return (
            <div className="flex justify-start">
              <Dropdown
                className="!p-1"
                id="more-stack"
                text={<DotsThreeVerticalIcon size="14" />}
                noIcon
                alignment="right"
              >
                <StackLinksModal
                  template_url={props?.row.original.template_url}
                  quick_link_url={props.row.original.quick_link_url}
                />
                <StackOutputsModal runs={props.row.original.runs} />
              </Dropdown>
            </div>
          )
        },
      },
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
