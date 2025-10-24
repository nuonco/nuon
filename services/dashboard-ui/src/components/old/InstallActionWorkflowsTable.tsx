'use client'

import React, { type FC, useEffect, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { CaretRightIcon, TimerIcon, CalendarBlankIcon, MinusIcon } from '@phosphor-icons/react'
import { ActionTriggerType } from '@/components/old/ActionTriggerType'
import { Table } from '@/components/old/DataTable'
import { DebouncedSearchInput } from '@/components/old/DebouncedSearchInput'
import { Link } from '@/components/old/Link'
import { StatusBadge } from '@/components/old/Status'
import { Time, Duration } from '@/components/old/Time'
import { ID, Text } from '@/components/old/Typography'
import { InstallActionTriggerFilter } from '@/components/old/InstallActionTriggerFilter'
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
  const [data, updateData] = useState(parseActionData(actions))
  const [columnFilters] = useState([])
  const [globalFilter] = useState('')

  useEffect(() => {
    updateData(parseActionData(actions))
  }, [actions])

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
        header: 'Last run',
        accessorKey: 'latest_run.updated_at',
        cell: (props) =>
          props.row.original?.latest_run ? (
            <Text>
              <CalendarBlankIcon size={18} />
              <Time time={props.getValue<string>()} format="relative" />
            </Text>
          ) : (
            <MinusIcon />
          ),
      },
      {
        header: 'Duration',
        accessorKey: 'latest_run.execution_time',
        cell: (props) =>
          props.row.original?.latest_run &&
          (props.row.original?.latest_run?.status_v2?.status as string) !==
            'queued' &&
          (props.row.original?.latest_run?.status as string) !== 'queued' ? (
            <Text>
              <TimerIcon size={18} />
              <Duration nanoseconds={props.getValue<number>()} />
            </Text>
          ) : (
            <MinusIcon />
          ),
      },
      {
        header: 'Last trigger',
        id: 'trigger_by_type',
        accessorKey: 'latest_run.triggered_by_type',
        cell: (props) =>
          props.row.original?.latest_run ? (
            <ActionTriggerType
              triggerType={props.getValue<string>()}
              componentName={
                props?.row?.original?.latest_run?.run_env_vars?.COMPONENT_NAME
              }
              componentPath={`/${orgId}/installs/${installId}/components/${props?.row?.original?.latest_run?.run_env_vars?.COMPONENT_ID}`}
            />
          ) : (
            <MinusIcon />
          ),
      },
      {
        id: 'status',
        header: 'Status',
        cell: (props) => (
          <div className="inline-flex h-12 w-full">
            {props.row.original?.latest_run ? (
              <StatusBadge
                status={
                  props.row.original?.latest_run?.status_v2?.status ||
                  props.row.original?.latest_run?.status
                }
              />
            ) : (
              <MinusIcon />
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
            <CaretRightIcon />
          </Link>
        ),
      },
    ],
    []
  )

  return (
    <Table
      header={
        <>
          <DebouncedSearchInput placeholder="Search action name" />
          <InstallActionTriggerFilter />
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
