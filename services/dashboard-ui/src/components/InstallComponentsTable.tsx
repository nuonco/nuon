'use client'

import { useEffect, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { CaretRightIcon, MinusIcon } from '@phosphor-icons/react'
import { AppConfigGraph } from '@/components/Apps'
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
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TInstallComponentSummary, TPaginationParams } from '@/types'

export interface IInstallComponentsTable
  extends IPollingProps,
    TPaginationParams {
  initInstallComponents: Array<TInstallComponentSummary>

  q?: string
  types?: string
}

export const InstallComponentsTable = ({
  initInstallComponents,
  pollInterval = 10000,
  shouldPoll = false,
  offset,
  limit,
  q,
  types,
}: IInstallComponentsTable) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const params = useQueryParams({ q, offset, limit, types })
  const { data: installComponents } = usePolling<TInstallComponentSummary[]>({
    dependencies: [params],
    initData: initInstallComponents,
    path: `/api/orgs/${org.id}/installs/${install.id}/components/summary${params}`,
    pollInterval,
    shouldPoll,
  })

  const [data, updateData] = useState(installComponents)
  const [columnFilters] = useState([
    {
      id: 'component_config.type',
      value: [
        'docker_build',
        'external_image',
        'helm_chart',
        'terraform_module',
        'kubernetes_manifest',
      ],
    },
  ])
  const [globalFilter] = useState('')

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
              href={`/${org.id}/installs/${install?.id}/components/${props.row.original.component_id}`}
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
            <MinusIcon />
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
            <MinusIcon />
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
                  installId={install.id}
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
            <MinusIcon />
          ),
      },
      {
        id: 'test',
        enableSorting: false,
        cell: (props) => (
          <Link
            href={`/${org.id}/installs/${install.id}/components/${props.row.original.component_id}`}
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
        <div className="flex-auto flex flex-col gap-2">
          <div className="w-full flex items-start justify-between">
            <DebouncedSearchInput placeholder="Search component name" />

            <div className="flex items-center gap-4">
              <AppConfigGraph
                appId={install?.app_id}
                configId={install?.app_config_id}
              />
              <DeployComponentsModal installId={install.id} orgId={org.id} />
              <DeleteComponentsModal installId={install.id} orgId={org.id} />
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
