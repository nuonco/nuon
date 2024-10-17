'use client'

import React, { type FC, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { DotsThreeVertical } from '@phosphor-icons/react'
import {
  ClickToCopy,
  DataTableSearch,
  Heading,
  Link,
  StaticComponentConfigType,
  StatusBadge,
  Table,
  Text,
  getComponentConfigType,
} from '@/components'
import type { TBuild, TComponent, TComponentConfig } from '@/types'

type TDataComponent = {
  deps: Array<TComponent>
  config: TComponentConfig
  latestBuild: TBuild
} & TComponent

type TData = {
  build: string
  componentId: string
  componentType: string
  configVersion: number
  dependencies: number
  deps: Array<TComponent>
  name: string
}

function parseComponentsToTableData(
  components: Array<TDataComponent>
): Array<TData> {
  return components.map((component) => ({
    build: component?.latestBuild?.status || 'noop',
    componentId: component.id,
    componentType: getComponentConfigType(component.config),
    configVersion: component.config_versions,
    dependencies: component.dependencies?.length || 0,
    deps: component.deps,
    name: component.name,
  }))
}

export interface IAppComponentsTable {
  appId: string
  components: Array<TDataComponent>
  orgId: string
}

export const AppComponentsTable: FC<IAppComponentsTable> = ({
  appId,
  components,
  orgId,
}) => {
  const [data, _] = useState(parseComponentsToTableData(components))
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
              href={`/${orgId}/apps/${appId}/components/${props.row.original.componentId}`}
            >
              <Heading variant="subheading">{props.getValue<string>()}</Heading>
            </Link>
            <ClickToCopy>
              <Text variant="id">{props.row.original.componentId}</Text>
            </ClickToCopy>
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
                  <Link href={`/${orgId}/apps/${appId}/components/${dep.id}`}>
                    {dep?.name}
                  </Link>
                </Text>
              ))
            ) : (
              <Text className="text-sm">None</Text>
            )}
          </div>
        ),
      },
      {
        header: 'Build',
        accessorKey: 'build',
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
            href={`/${orgId}/apps/${appId}/components/${props.row.original.componentId}`}
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
