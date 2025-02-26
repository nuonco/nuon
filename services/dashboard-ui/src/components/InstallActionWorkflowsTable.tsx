'use client'

import React, { type FC, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { CaretRight, Timer, CalendarBlank, Minus } from '@phosphor-icons/react'
import { Badge } from '@/components/Badge'
import { DataTableSearch, Table } from '@/components/DataTable'
import { Link } from '@/components/Link'
import { StatusBadge } from '@/components/Status'
import { Time, Duration } from '@/components/Time'
import { EventStatus } from '@/components/Timeline'
import { ID, Text } from '@/components/Typography'
// eslint-disable-next-line import/no-cycle
import type {
  TActionWorkflow,
  TInstallActionWorkflow,
  TInstallActionWorkflowRun,
} from '@/types'

type TData = {
  action_workflow: TActionWorkflow
  latest_run: TInstallActionWorkflowRun
}

function parseActionData(actions: Array<TInstallActionWorkflow>): Array<TData> {
  return actions?.map((a) => ({
    action_workflow: a?.action_workflow,
    latest_run: a?.runs?.at(0),
  }))
}

export interface IInstallActionWorkflowsTable {
  installId: string
  actions: Array<TInstallActionWorkflow>
  orgId: string
}

export const InstallActionWorkflowsTable: FC<IInstallActionWorkflowsTable> = ({
  installId,
  actions,
  orgId,
}) => {
  const [data, _] = useState(parseActionData(actions))
  const [columnFilters, __] = useState([])
  const [globalFilter, setGlobalFilter] = useState('')

  const columns: Array<ColumnDef<TData>> = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'action_workflow.name',
        cell: (props) => (
          <div className="flex flex-col gap-2">
            <Link
              href={`/${orgId}/installs/${installId}/actions/${props.row.original.action_workflow.id}`}
            >
              <Text variant="med-14">{props.getValue<string>()}</Text>
            </Link>

            <ID id={props.row.original.action_workflow.id} />
          </div>
        ),
      },
      {
        header: 'Time since last run',
        accessorKey: 'latest_run.updated_at',
        cell: (props) =>
          props.row.original?.latest_run ? (
            <Text>
              <CalendarBlank size={18} />
              <Time time={props.getValue<string>()} format="relative" />
            </Text>
          ) : (
            <Minus />
          ),
      },
      {
        header: 'Run duration',
        accessorKey: 'latest_run.execution_time',
        cell: (props) =>
          props.row.original?.latest_run ? (
            <Text>
              <Timer size={18} />
              <Duration nanoseconds={props.getValue<number>()} />
            </Text>
          ) : (
            <Minus />
          ),
      },
      {
        header: 'Recent trigger',
        accessorKey: 'latest_run.trigger_type',
        cell: (props) =>
          props.row.original?.latest_run ? (
            <Badge variant="code">{props.getValue<string>()}</Badge>
          ) : (
            <Minus />
          ),
      },
      {
        id: 'status',
        header: 'Status',
        cell: (props) => (
          <div className="inline-flex h-12 w-full">
            {props.row.original?.latest_run ? (
              <StatusBadge status={props.row.original?.latest_run?.status} />
            ) : (
              <Minus />
            )}
          </div>
        ),
      },
      {
        id: 'test',
        enableSorting: false,
        cell: (props) => (
          <Link
            href={`/${orgId}/installs/${installId}/actions/${props.row.original.action_workflow.id}`}
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
