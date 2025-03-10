'use client'

import React, { type FC, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { CaretRight } from '@phosphor-icons/react'
import { ActionTriggerType } from '@/components/ActionTriggerType'
import { Link } from '@/components/Link'
import { DataTableSearch, Table } from '@/components/DataTable'
import { ID, Text } from '@/components/Typography'
import type {
  TActionWorkflow,
  TActionConfigTriggerType,
  TActionConfigStep,
} from '@/types'

type TData = {
  id?: string
  name?: string
  config_count?: number
  steps?: TActionConfigStep[]
  triggers?: TActionConfigTriggerType[] | string[]
}

function parseWorkflowsToTableData(
  workflows: Array<TActionWorkflow>
): Array<TData> {
  return workflows.map((wf) => ({
    id: wf.id,
    name: wf.name,
    config_count: wf.config_count,
    steps: wf?.configs?.[0]?.steps?.map((s) => s),
    triggers: wf?.configs?.[0]?.triggers?.map((t) => t?.type),
  }))
}

export interface IAppWorkflowsTable {
  appId: string
  orgId: string
  workflows: Array<TActionWorkflow>
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
              href={`/${orgId}/apps/${appId}/actions/${props.row.original.id}`}
            >
              <Text variant="med-14">{props.getValue<string>()}</Text>
            </Link>

            <ID id={props.row.original.id} />
          </div>
        ),
      },
      {
        header: 'Version',
        accessorKey: 'config_count',
        cell: (props) => (
          <Text className="gap-4">{props.getValue<string>()}</Text>
        ),
      },
      {
        header: 'Triggers',
        accessorKey: 'triggers',
        cell: (props) => (
          <Text className="gap-4">
            {props
              .getValue<TActionConfigTriggerType[]>()
              ?.map((t) => <ActionTriggerType key={t} triggerType={t} />)}
          </Text>
        ),
      },
      {
        header: 'Steps',
        accessorKey: 'steps',
        cell: (props) => (
          <ol className="flex flex-col gap-1 list-decimal">
            {props
              .getValue<TActionConfigStep[]>()
              ?.sort((a, b) => b?.idx - a?.idx)
              ?.reverse()
              ?.map((s) => (
                <li key={s?.id} className="text-sm">
                  <Text className="!leading-none self-start">{s?.name}</Text>
                </li>
              ))}
          </ol>
        ),
      },
      {
        id: 'test',
        enableSorting: false,
        cell: (props) => (
          <Link
            href={`/${orgId}/apps/${appId}/actions/${props.row.original.id}`}
            variant="ghost"
          >
            <CaretRight />
          </Link>
        ),
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
      emptyTitle="No actions found"
      globalFilter={globalFilter}
    />
  )
}
