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
import type { TInstallComponent, TPaginationParams } from '@/types'

export interface IInstallComponentsTable
  extends IPollingProps,
    TPaginationParams {
  initInstallComponents: Array<TInstallComponent>
  componentDeps: { id: string; component_id: string; dependencies: string[] }[]
  q?: string
  types?: string
}

export const InstallComponentsTable = ({
  initInstallComponents,
  componentDeps,
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
  const { data: installComponents } = usePolling<TInstallComponent[]>({
    dependencies: [params],
    initData: initInstallComponents,
    path: `/api/orgs/${org.id}/installs/${install.id}/components${params}`,
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

  const columns: Array<
    ColumnDef<TInstallComponent & { dependencies: string[] }>
  > = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'component.name',
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
        accessorKey: 'component.type',
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
        accessorKey: 'status',
        cell: (props) =>
          props.getValue<string>() ? (
            <StatusBadge
              status={
                props.row?.original?.status_v2?.status ||
                props.getValue<string>()
              }
              description={
                props.row?.original?.status_v2?.status_human_description
              }
            />
          ) : (
            <MinusIcon />
          ),
      },
      {
        header: 'Dependencies',
        id: 'dependencies',
        enableSorting: false,
        cell: (props) => {
          const depIndex = componentDeps?.findIndex(
            (dep) => dep?.id === props?.row?.original?.id
          )

          return (
            <div className="flex flex-wrap items-center gap-4">
              {componentDeps?.at(depIndex)?.dependencies?.length ? (
                <div className="flex items-center gap-4 flex-wrap w-full">
                  <ComponentDependencies
                    deps={install?.install_components
                      ?.map((ic) =>
                        componentDeps
                          ?.at(depIndex)
                          ?.dependencies?.includes(ic?.component_id)
                          ? ic?.component
                          : undefined
                      )
                      .filter(Boolean)}
                    name={props.row.original?.component?.name}
                    installId={install.id}
                  />
                </div>
              ) : (
                <Text>None</Text>
              )}
            </div>
          )
        },
      },
      /* {
       *   header: 'Build',
       *   accessorKey: 'build_status',
       *   cell: (props) =>
       *     props.getValue<string>() ? (
       *       <StatusBadge
       *         status={props.getValue<string>()}
       *         description={props.row?.original?.build_status_description}
       *         descriptionAlignment="right"
       *       />
       *     ) : (
       *       <MinusIcon />
       *     ),
       * }, */
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
              <DeployComponentsModal />
              <DeleteComponentsModal />
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
