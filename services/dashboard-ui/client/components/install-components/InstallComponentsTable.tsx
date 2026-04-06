import { useSearchParams } from 'react-router'
import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import { ComponentTypeFilterDropdown } from '@/components/components/ComponentTypeFilter'
import { ComponentType } from '@/components/components/ComponentType'
import { InstallComponentDependencies } from '@/components/install-components/InstallComponentDependencies'
import { ManageAllDropdown } from '@/components/install-components/management/ManageAllDropdown'
import { QuickComponentManagementDropdown } from '@/components/install-components/management/QuickComponentManagementDropdown'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallComponents, getAppConfig } from '@/lib'
import type { TInstallComponent } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

const LIMIT = 10

type TComponentDeps = {
  id: string
  component_id: string
  dependencies: string[]
}

export type InstallComponentRow = {
  componentId: string
  componentName: string
  componentType: ReactNode
  deployStatus: ReactNode
  driftStatus: ReactNode
  href: string
  action: ReactNode
  dependencies: ReactNode
}

function parseInstallComponentSummaryToTableData(
  components: TInstallComponent[],
  deps: TComponentDeps[],
  orgId: string,
  installId: string
): InstallComponentRow[] {
  return components.map((component) => {
    const depIndex = deps?.findIndex((dep) => dep?.id === component?.id)

    return {
      componentId: component.component_id,
      componentName: component.component?.name,
      componentType: (
        <ComponentType
          type={component?.component?.type}
          variant="subtext"
          colorVariant="color"
        />
      ),
      deployStatus: (
        <Tooltip
          position="top"
          tipContentClassName="!p-0"
          tipContent={
            <div className="w-fit max-w-64">
              {component?.status_v2?.status ? (
                <>
                  <Time
                    className="!text-nowrap px-2 py-1"
                    variant="subtext"
                    seconds={component?.status_v2?.created_at_ts}
                    weight="strong"
                  />
                  <hr className="my-1" />
                  <Text as="div" className="flex px-2 pb-2" variant="subtext">
                    {toSentenceCase(
                      component?.status_v2?.status_human_description
                    )}
                  </Text>
                </>
              ) : (
                <Text flex nowrap className="p-2" variant="subtext">
                  Status unknown
                </Text>
              )}
            </div>
          }
        >
          <Status variant="badge" status={component.status_v2?.status} />
        </Tooltip>
      ),
      driftStatus: component?.drifted_object ? (
        <Status variant="badge" status="drifted" />
      ) : (
        <Icon variant="MinusIcon" />
      ),
      dependencies: (
        <InstallComponentDependencies deps={deps?.at(depIndex)?.dependencies} />
      ),
      href: `/${orgId}/installs/${installId}/components/${component.component_id}`,
      action: (
        <div className="hidden md:block">
          <QuickComponentManagementDropdown
            installComponent={component}
            orgId={orgId}
            installId={installId}
          />
        </div>
      ),
    }
  })
}

const columns: ColumnDef<InstallComponentRow>[] = [
  {
    accessorKey: 'componentName',
    header: 'Component name',
    cell: (info) => (
      <span>
        <Text variant="body">
          <Link href={info.row.original.href}>{info.getValue() as string}</Link>
        </Text>
        <ID>{info.row.original.componentId as string}</ID>
      </span>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'componentType',
    header: 'Type',
    cell: (info) => <Text>{info.getValue() as string}</Text>,
  },
  {
    enableSorting: true,
    accessorKey: 'dependencies',
    header: 'Dependencies',
    cell: (info) => (
      <Text as="div" className="flex">{info.getValue() as ReactNode}</Text>
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'deployStatus',
    header: 'Latest deploy',
    cell: (info) => (
      <Text className="flex">{info.getValue() as ReactNode}</Text>
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'driftStatus',
    header: 'Drifted',
    cell: (info) => (
      <Text as="div" className="flex">{info.getValue() as ReactNode}</Text>
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'action',
    id: 'action',
    header: '',
    cell: (info) => info.getValue<ReactNode>(),
  },
]

export const InstallComponentsTable = ({
  pollInterval = 20000,
  shouldPoll,
}: {
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: componentsResult, isLoading } = useQuery({
    queryKey: [
      'install-components',
      org?.id,
      install?.id,
      offset,
      searchParams.get('q'),
      searchParams.get('types'),
    ],
    queryFn: () =>
      getInstallComponents({
        orgId: org.id,
        installId: install.id,
        limit: LIMIT,
        offset,
        q: searchParams.get('q') || undefined,
        types: searchParams.get('types') || undefined,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  const { data: configResult } = useQuery({
    queryKey: [
      'app-config',
      org?.id,
      install?.app_id,
      install?.app_config_id,
      'recurse',
    ],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: install.app_config_id,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_config_id,
  })

  const components = componentsResult?.data ?? []
  const pagination = {
    hasNext: componentsResult?.pagination?.hasNext ?? false,
    offset,
    limit: LIMIT,
  }

  const deps: TComponentDeps[] = components.map((ic) => ({
    id: ic?.id,
    component_id: ic?.component_id,
    dependencies: configResult?.component_config_connections?.find(
      (c) => c?.component_id === ic?.component_id
    )?.component_dependency_ids,
  }))

  if (isLoading) return <InstallComponentsTableSkeleton />

  return (
    <Table<InstallComponentRow>
      columns={columns}
      data={parseInstallComponentSummaryToTableData(
        components,
        deps,
        org?.id ?? '',
        install?.id ?? ''
      )}
      filterActions={
        <div className="flex items-center gap-3">
          <ComponentTypeFilterDropdown />
          <ManageAllDropdown />
        </div>
      }
      emptyMessage="No components found"
      pagination={pagination}
      searchPlaceholder="Search component name..."
    />
  )
}

export const InstallComponentsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
