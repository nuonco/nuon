'use client'

import React, { type FC, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { DotsThreeVertical } from '@phosphor-icons/react'
import {
  DataTableSearch,
  Heading,
  Link,
  StaticComponentConfigType,
  StatusBadge,
  Table,
  Text,
  Time,
  getComponentConfigType,
} from '@/components'
import type { TBuild, TComponentConfig, TInstallComponent } from '@/types'

export type TDataInstallComponent = {
  build: TBuild
  deps: Array<TInstallComponent>
  config: TComponentConfig
} & TInstallComponent

type TData = {
  buildStatus: string
  componentType: string
  configVersion: number
  installComponentId: string
  deployDate: string
  dependencies: number
  deps: Array<TInstallComponent>
  name: string
}

function parseInstallComponentsToTableData(
  installComponents: Array<TDataInstallComponent>
): Array<TData> {
  return installComponents.map((comp) => ({
    buildStatus: comp.build?.status || 'noop',
    componentType: getComponentConfigType(comp.config),
    configVersion: comp.config?.version,
    installComponentId: comp.id,
    deployDate: comp.install_deploys?.[0]?.created_at,
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
  const [data, _] = useState(
    parseInstallComponentsToTableData(installComponents)
  )
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
              href={`/beta/${orgId}/installs/${installId}/components/${props.row.original.installComponentId}`}
            >
              <Heading variant="subheading">{props.getValue<string>()}</Heading>
            </Link>
            <Text variant="id">{props.row.original.installComponentId}</Text>
          </div>
        ),
      },
      {
        header: 'Type',
        accessorKey: 'componentType',
        cell: (props) => (
          <Text className="gap-4">
            <StaticComponentConfigType configType={props.getValue<string>()} />
          </Text>
        ),
      },
      {
        header: 'Deployment',
        accessorKey: 'deployDate',
        cell: (props) => (
          <Time
            time={props.getValue<string>()}
            format="relative"
            variant="caption"
          />
        ),
      },
      {
        header: 'Dependencies',
        accessorKey: 'dependencies',
        cell: (props) => (
          <div className="flex flex-wrap items-center gap-4">
            {props.getValue<number>() ? (
              props.row.original.deps.map((dep, i) => (
                <Text
                  key={`${dep.id}-${i}`}
                  className="bg-gray-500/10 px-2 py-1 rounded-lg border w-fit"
                  variant="caption"
                >
                  <Link
                    href={`/beta/${orgId}/installs/${installId}/components/${dep.id}`}
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
            href={`/beta/${orgId}/installs/${installId}/components/${props.row.original.installComponentId}`}
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
