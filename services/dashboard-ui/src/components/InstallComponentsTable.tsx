'use client'

import React, { type FC, useEffect, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { CaretRight } from '@phosphor-icons/react'
import {
  StaticComponentConfigType,
  getComponentConfigType,
} from '@/components/ComponentConfig'
import { InstallComponentsManagementDropdown } from '@/components/Installs'
import { Link } from '@/components/Link'
import { StatusBadge } from '@/components/Status'
import { DataTableSearch, Table } from '@/components/DataTable'
import { ID, Text } from '@/components/Typography'
// eslint-disable-next-line import/no-cycle
import type { TBuild, TComponentConfig, TInstallComponent } from '@/types'

export type TDataInstallComponent = {
  build: TBuild
  deps: Array<TInstallComponent>
  config: TComponentConfig
} & TInstallComponent

type TData = {
  buildStatus: string
  componentId: string
  componentType: string
  configVersion: number
  installComponentId: string
  deployStatus: string
  dependencies: number
  deps: Array<TInstallComponent>
  name: string
}

function parseInstallComponentsToTableData(
  installComponents: Array<TDataInstallComponent>
): Array<TData> {
  return installComponents.map((comp) => ({
    buildStatus: comp.build?.status || 'No build',
    componentId: comp.component_id,
    componentType: comp?.config ?  getComponentConfigType(comp?.config) : undefined,
    configVersion: comp.config?.version,
    installComponentId: comp.id,
    deployStatus: comp.install_deploys?.[0]?.status || 'Not deployed',
    dependencies: comp.deps?.length || 0,
    deps: comp.deps,
    name: comp.component?.name,
  }))
}

export interface IInstallComponentsTable {
  installComponents: Array<TDataInstallComponent>
  installId: string
  orgId: string
}

export const InstallComponentsTable: FC<IInstallComponentsTable> = ({
  installComponents,
  installId,
  orgId,
}) => {
  const [data, updateData] = useState(
    parseInstallComponentsToTableData(installComponents)
  )
  const [columnFilters, __] = useState([])
  const [globalFilter, setGlobalFilter] = useState('')

  useEffect(() => {
    updateData(parseInstallComponentsToTableData(installComponents))
  }, [installComponents])

  const columns: Array<ColumnDef<TData>> = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'name',
        cell: (props) => (
          <div className="flex flex-col gap-2">
            <Link
              href={`/${orgId}/installs/${installId}/components/${props.row.original.componentId}`}
            >
              <Text variant="med-14">{props.getValue<string>()}</Text>
            </Link>

            <ID id={props.row.original.componentId} />
          </div>
        ),
      },
      {
        header: 'Type',
        accessorKey: 'componentType',
        cell: (props) => props.getValue<string>() ? (
          <Text className="gap-4">
            <StaticComponentConfigType configType={props.getValue<string>()} />
          </Text>
        ) : "-", 
      },
      {
        header: 'Deployment',
        accessorKey: 'deployStatus',
        cell: (props) => <StatusBadge status={props.getValue<string>()} />,
      },
      {
        header: 'Dependencies',
        accessorKey: 'dependencies',
        enableSorting: false,
        cell: (props) => (
          <div className="flex flex-wrap items-center gap-4">
            {props.getValue<number>() ? (
              props.row.original.deps.map((dep, i) => (
                <Text
                  key={`${dep.id}-${i}`}
                  className="bg-gray-500/10 px-2 py-1 rounded-lg border w-fit"
                >
                  <Link
                    href={`/${orgId}/installs/${installId}/components/${dep.component_id}`}
                  >
                    {dep?.component?.name}
                  </Link>
                </Text>
              ))
            ) : (
              <Text>None</Text>
            )}
          </div>
        ),
      },
      {
        header: 'Build',
        accessorKey: 'buildStatus',
        cell: (props) => <StatusBadge status={props.getValue<string>()} />,
      },
      {
        header: 'Config',
        accessorKey: 'configVersion',
        cell: (props) => <Text>{props.getValue<number>()}</Text>,
      },
      {
        id: 'test',
        enableSorting: false,
        cell: (props) => (
          <Link
            href={`/${orgId}/installs/${installId}/components/${props.row.original.componentId}`}
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
          <InstallComponentsManagementDropdown />
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
