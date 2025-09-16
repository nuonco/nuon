'use client'

import React, { type FC, useEffect, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { CaretRightIcon } from '@phosphor-icons/react'
import { AppConfigGraph } from '@/components/Apps'
import {
  BuildAllComponentsButton,
  ComponentConfigType,
  ComponentDependencies,
  type TComponentConfigType,
} from '@/components/Components'
import { ComponentTypeFilterDropdown } from '@/components/Components/NewComponentTypeFilter'
import { Link } from '@/components/Link'
import { StatusBadge } from '@/components/Status'
import { Table } from '@/components/DataTable'
import { DebouncedSearchInput } from '@/components/DebouncedSearchInput'
import { ID, Text } from '@/components/Typography'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'

// eslint-disable-next-line import/no-cycle
import type { TBuild, TComponent, TPaginationParams } from '@/types'

type TDataComponent = {
  deps: Array<TComponent>
  latestBuild?: TBuild
} & TComponent

type TData = {
  build: string
  componentId: string
  type: string
  configVersion: number
  dependencies: number
  deps: Array<TComponent>
  name: string
}

function parseComponentsToTableData(
  components: Array<TDataComponent>
): Array<TData> {
  return components.map((component) => ({
    build:
      component?.latestBuild?.status_v2?.status ||
      component?.latestBuild?.status ||
      'noop',
    componentId: component.id,
    type: component?.type,
    configVersion: component.config_versions,
    dependencies: component.dependencies?.length || 0,
    deps: component.deps,
    name: component.name,
  }))
}

export interface IAppComponentsTable extends IPollingProps, TPaginationParams {
  appId: string
  initComponents: Array<TDataComponent>
  orgId: string
  configId: string
}

export const AppComponentsTable: FC<IAppComponentsTable> = ({
  appId,
  initComponents: components,
  configId,
  orgId,
  /* pollInterval = 5000,
   * shouldPoll = false,
   * offset,
   * limit, */
}) => {
  /* const { org } = useOrg()
   * const { app } = useApp()
   * const params = useQueryParams({ limit, offset })
   * const { data: components } = usePolling<TComponent[]>({
   *   dependencies: [params],
   *   initData: initComponents,
   *   path: `/api/orgs/${org.id}/apps/${app.id}/components${params}`,
   *   pollInterval,
   *   shouldPoll,
   * }) */

  const [data, updateData] = useState(parseComponentsToTableData(components))
  const [columnFilters, setColumnFilters] = useState([
    {
      id: 'type',
      value: [
        'docker_build',
        'external_image',
        'helm_chart',
        'terraform_module',
        'kubernetes_manifest',
      ],
    },
  ])
  const [globalFilter, setGlobalFilter] = useState('')

  useEffect(() => {
    updateData(parseComponentsToTableData(components))
  }, [components])

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
              <Text variant="med-14">{props.getValue<string>()}</Text>
            </Link>

            <ID id={props.row.original.componentId} />
          </div>
        ),
      },
      {
        header: 'Type',
        accessorKey: 'type',
        filterFn: 'arrIncludesSome',
        cell: (props) => (
          <Text className="gap-4">
            <ComponentConfigType
              configType={props.getValue<TComponentConfigType>()}
            />
          </Text>
        ),
      },
      {
        header: 'Dependencies',
        accessorKey: 'dependencies',
        cell: (props) => (
          <div className="flex flex-wrap items-center gap-4">
            {props.getValue<number>() ? (
              <ComponentDependencies
                deps={props.row.original.deps}
                name={props.row.original.name}
              />
            ) : (
              <Text>None</Text>
            )}
          </div>
        ),
      },
      {
        header: 'Build',
        accessorKey: 'build',
        cell: (props) => (
          <StatusBadge pollDuration={10000} status={props.getValue<string>()} />
        ),
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
            <CaretRightIcon />
          </Link>
        ),
      },
    ],
    []
  )

  const handleTypeFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { checked, value } = e.target
    setColumnFilters((state) => {
      const values = [...state?.at(0)?.value]
      const index = values?.indexOf(value)

      if (checked && index < 0) {
        values.push(value)
      } else if (index > -1) {
        values.splice(index, 1)
      }

      return [{ id: 'type', value: values }]
    })
  }

  const handleTypeOnlyFilter = (e: React.MouseEvent<HTMLButtonElement>) => {
    setColumnFilters([{ id: 'type', value: [e?.currentTarget?.value] }])
  }

  const clearTypeFilter = () => {
    setColumnFilters([
      {
        id: 'type',
        value: [
          'docker_build',
          'external_image',
          'helm_chart',
          'terraform_module',
          'kubernetes_manifest',
        ],
      },
    ])
  }

  const handleGlobleFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    setGlobalFilter(e.target.value || '')
  }

  return (
    <Table
      header={
        <div className="flex-auto flex flex-col gap-2">
          <div className="w-full flex items-start justify-between">
            <DebouncedSearchInput placeholder="Search component name" />

            <div className="flex items-center gap-4">
              <AppConfigGraph appId={appId} configId={configId} />
              <BuildAllComponentsButton components={components} />
            </div>
          </div>
          <ComponentTypeFilterDropdown
            {...{
              handleTypeFilter,
              handleTypeOnlyFilter,
              clearTypeFilter,
              columnFilters,
            }}
            isNotDropdown
          />
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
