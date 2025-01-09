'use client'

import classNames from 'classnames'
import React, { type FC, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import {
  CaretRight,
  Timer,
  CalendarBlank,
  CheckCircle,
  ClockCountdown,
  Minus,
} from '@phosphor-icons/react'
import { Badge } from '@/components/Badge'
import { DataTableSearch, Table } from '@/components/DataTable'
import { Link } from '@/components/Link'
import { Time, Duration } from '@/components/Time'
import { EventStatus } from '@/components/Timeline'
import { ID, Text } from '@/components/Typography'
// eslint-disable-next-line import/no-cycle
import type { TActionWorkflow, TInstallActionWorkflowRun } from '@/types'

const WorkflowRunStatus: FC<{ status: string }> = ({ status }) => {
  const statusColor = {
    'text-green-800 dark:text-green-500':
      status === 'finished' || status === 'active',
    'text-red-600 dark:text-red-500': status === 'failed' || status === 'error',
    'text-cool-grey-600 dark:text-cool-grey-500': status === 'noop',
    'text-orange-800 dark:text-orange-500':
      status === 'waiting' ||
      status === 'started' ||
      status === 'in-progress' ||
      status === 'building' ||
      status === 'queued' ||
      status === 'planning' ||
      status === 'deploying',
  }

  return (
    <span className={classNames('w-4 h-4 rounded-full', statusColor)}>
      {status === 'active' ? <CheckCircle /> : <ClockCountdown />}
    </span>
  )
}

type TData = {
  action_workflow: TActionWorkflow
  install_action_workflow_run: TInstallActionWorkflowRun
}

export interface IInstallActionWorkflowsTable {
  installId: string
  actions: Array<TData>
  orgId: string
}

export const InstallActionWorkflowsTable: FC<IInstallActionWorkflowsTable> = ({
  installId,
  actions,
  orgId,
}) => {
  const [data, _] = useState(actions)
  const [columnFilters, __] = useState([])
  const [globalFilter, setGlobalFilter] = useState('')

  const columns: Array<ColumnDef<TData>> = useMemo(
    () => [
      {
        id: 'status',
        cell: (props) => (
          <div className="inline-flex h-12 w-full items-center justify-center">
            {props.row.original.install_action_workflow_run ? (
              <EventStatus
                status={props.row.original.install_action_workflow_run.status}
              />
            ) : (
              <EventStatus status="noop" />
            )}
          </div>
        ),
      },
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
        accessorKey: 'install_action_workflow_run.updated_at',
        cell: (props) =>
          props.row.original.install_action_workflow_run ? (
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
        accessorKey: 'install_action_workflow_run.execution_time',
        cell: (props) =>
          props.row.original.install_action_workflow_run ? (
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
        accessorKey: 'install_action_workflow_run.trigger_type',
        cell: (props) =>
          props.row.original.install_action_workflow_run ? (
            <Badge variant="code">{props.getValue<string>()}</Badge>
          ) : (
            <Minus />
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
