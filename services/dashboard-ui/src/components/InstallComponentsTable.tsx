'use client'

import React, { type FC, useEffect, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { CaretRight, Minus } from '@phosphor-icons/react'
import {
  ComponentDependencies,
  ComponentConfigType,
  type TComponentConfigType,
} from '@/components/Components'
import { ComponentTypeFilterDropdown } from '@/components/Components/NewComponentTypeFilter'
import {
  DeleteComponentsModal,
  DeployComponentsModal,
} from '@/components/InstallComponents'
import { Link } from '@/components/Link'
import { StatusBadge } from '@/components/Status'
import { Table } from '@/components/DataTable'
import { DebouncedSearchInput } from '@/components/DebouncedSearchInput'
import { ID, Text } from '@/components/Typography'
// eslint-disable-next-line import/no-cycle
import type { TInstallComponentSummary } from '@/types'

export interface IInstallComponentsTable {
  installComponents: Array<TInstallComponentSummary>
  installId: string
  orgId: string
}

export const InstallComponentsTable: FC<IInstallComponentsTable> = ({
  installComponents,
  installId,
  orgId,
}) => {
  const [data, updateData] = useState(installComponents)
  const [columnFilters, setColumnFilters] = useState([
    {
      id: 'component_config.type',
      value: [
        'docker_build',
        'external_image',
        'helm_chart',
        'terraform_module',
      ],
    },
  ])
  const [globalFilter, setGlobalFilter] = useState('')

  useEffect(() => {
    updateData(installComponents)
  }, [installComponents])

  const columns: Array<ColumnDef<TInstallComponentSummary>> = useMemo(
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
        accessorKey: 'component_config.type',
        id: 'component_config.type',
        filterFn: 'arrIncludesSome',
        cell: (props) =>
          props.getValue<TComponentConfigType>() ? (
            <Text className="gap-4">
              <ComponentConfigType
                configType={props.getValue<TComponentConfigType>()}
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
              shouldPoll={props?.row.index === 0}
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

      return [{ id: 'component_config.type', value: values }]
    })
  }

  const handleTypeOnlyFilter = (e: React.MouseEvent<HTMLButtonElement>) => {
    setColumnFilters([
      { id: 'component_config.type', value: [e?.currentTarget?.value] },
    ])
  }

  const clearTypeFilter = () => {
    setColumnFilters([
      {
        id: 'component_config.type',
        value: [
          'docker_build',
          'external_image',
          'helm_chart',
          'terraform_module',
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
              <DeployComponentsModal installId={installId} orgId={orgId} />
              <DeleteComponentsModal installId={installId} orgId={orgId} />
            </div>
          </div>
          <ComponentTypeFilterDropdown isNotDropdown />
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
