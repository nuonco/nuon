'use client'

import React, { type FC, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { DotsThreeVertical } from '@phosphor-icons/react'
import { Link } from '@/components/Link'
import { DataTableSearch, Table } from '@/components/DataTable'
import { ID, Text } from '@/components/Typography'
import type { TWorkflow } from '@/types'

type TData = {
  id: string
  name: string
  on: string
  jobCount: number
}

function parseWorkflowsToTableData(workflows: Array<TWorkflow>): Array<TData> {
  return workflows.map(({ jobs, ...wf }) => ({ ...wf, jobCount: jobs?.length }))
}

export interface IAppWorkflowsTable {
  appId: string
  orgId: string
  workflows: Array<TWorkflow>
}

export const AppWorkflowsTable: FC<IAppWorkflowsTable> = ({
  appId,
  orgId,
  workflows,
}) => {
  const [data, _] = useState(parseWorkflowsToTableData(workflows))
  const [columnFilters, __] = useState([])
  const [globalFilter, setGlobalFilter] = useState('')

  const columns: Array<ColumnDef<TData>> = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'name',
        cell: (props) => (
          <div className="flex flex-col gap-2">
            <Link
              href={`/${orgId}/apps/${appId}/workflows/${props.row.original.id}`}
            >
              <Text variant="med-14">{props.getValue<string>()}</Text>
            </Link>

            <ID id={props.row.original.id} />
          </div>
        ),
      },
      {
        header: 'On',
        accessorKey: 'on',
        cell: (props) => (
          <Text className="gap-4">{props.getValue<string>()}</Text>
        ),
      },
      {
        header: 'Jobs',
        accessorKey: 'jobCount',
        cell: (props) => <Text>{props.getValue<number>()}</Text>,
      },
      {
        id: 'test',
        enableSorting: false,
        cell: (props) => (
          <Link
            href={`/${orgId}/apps/${appId}/workflows/${props.row.original.id}`}
            variant="ghost"
          >
            <DotsThreeVertical />
          </Link>
        ),
      },
    ],
    []
  )

  const handleGlobleFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    setGlobalFilter(e.target.value)
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
      globalFilter={globalFilter}
    />
  )
}
