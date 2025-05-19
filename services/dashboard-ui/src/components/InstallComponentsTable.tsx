'use client'

import React, { type FC, useEffect, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { CaretRight, Minus } from '@phosphor-icons/react'
import { ComponentDependencies } from '@/components/Components'
import {
  StaticComponentConfigType,
  getComponentConfigType,
} from '@/components/ComponentConfig'
import {
  DeleteComponentsModal,
  DeployComponentsModal,
} from '@/components/InstallComponents'
import { Link } from '@/components/Link'
import { StatusBadge } from '@/components/Status'
import { DataTableSearch, Table } from '@/components/DataTable'
import { ID, Text } from '@/components/Typography'
// eslint-disable-next-line import/no-cycle
import type { TComponentConfig, TInstallComponentSummary } from '@/types'

export type TTableInstallComponent = TInstallComponentSummary & {
  config?: TComponentConfig
}

export interface IInstallComponentsTable {
  installComponents: Array<TTableInstallComponent>
  installId: string
  orgId: string
}

export const InstallComponentsTable: FC<IInstallComponentsTable> = ({
  installComponents,
  installId,
  orgId,
}) => {
  const [data, updateData] = useState(installComponents)
  const [columnFilters, __] = useState([])
  const [globalFilter, setGlobalFilter] = useState('')

  useEffect(() => {
    updateData(installComponents)
  }, [installComponents])

  const columns: Array<ColumnDef<TTableInstallComponent>> = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'component_name',
        cell: (props) => (
          <div className="flex flex-col gap-2">
            <Link
              href={`/${orgId}/installs/${installId}/components/${props.row.original.component_id}`}
            >
              <Text variant="med-14">{props.getValue<string>()}</Text>
            </Link>

            <ID id={props.row.original.component_id} />
          </div>
        ),
      },
      {
        header: 'Type',
        accessorKey: 'config',
        cell: (props) =>
          props.getValue<TComponentConfig>() ? (
            <Text className="gap-4">
              <StaticComponentConfigType
                configType={getComponentConfigType(
                  props.getValue<TComponentConfig>()
                )}
              />
            </Text>
          ) : (
            <Minus />
          ),
      },
      {
        header: 'Deployment',
        accessorKey: 'deploy_status',
        cell: (props) =>
          props.getValue<string>() ? (
            <StatusBadge
              status={props.getValue<string>()}
              description={props.row?.original?.deploy_status_description}
            />
          ) : (
            <Minus />
          ),
      },
      {
        header: 'Dependencies',
        accessorKey: 'dependencies',
        enableSorting: false,
        cell: (props) => (
          <div className="flex flex-wrap items-center gap-4">
            {props.getValue<number>() ? (
              <div className="flex items-center gap-4 flex-wrap w-full">
                <ComponentDependencies
                  deps={props?.row?.original?.dependencies}
                  name={props.row.original?.component_name}
                  installId={installId}
                />
              </div>
            ) : (
              <Text>None</Text>
            )}
          </div>
        ),
      },
      {
        header: 'Build',
        accessorKey: 'build_status',
        cell: (props) =>
          props.getValue<string>() ? (
            <StatusBadge
              status={props.getValue<string>()}
              description={props.row?.original?.build_status_description}
              descriptionAlignment="right"
            />
          ) : (
            <Minus />
          ),
      },
      /* {
       *   header: 'Config',
       *   accessorKey: 'configVersion',
       *   cell: (props) => <Text>{props.getValue<number>()}</Text>,
       * }, */
      {
        id: 'test',
        enableSorting: false,
        cell: (props) => (
          <Link
            href={`/${orgId}/installs/${installId}/components/${props.row.original.component_id}`}
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
        <div className="w-full flex items-start justify-between">
          <DataTableSearch
            handleOnChange={handleGlobleFilter}
            value={globalFilter}
          />
          <div className="flex items-center gap-4">
            <DeployComponentsModal installId={installId} orgId={orgId} />
            <DeleteComponentsModal installId={installId} orgId={orgId} />
          </div>
        </div>
      }
      data={data}
      columns={columns}
      columnFilters={columnFilters}
      emptyMessage="Reset your search and try again."
      emptyTitle="No components found"
      globalFilter={globalFilter}
    />
  )
}
