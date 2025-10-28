'use client'

import { useSearchParams } from 'next/navigation'
import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Menu } from '@/components/common/Menu'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { type IPagination } from '@/components/common/Pagination'
import { Modal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TInstallStack } from '@/types'
import { StackLinks } from './StackLinks'
import { StackOutputs } from './StackOutputs'

export type TInstallStackRow = {
  versionId: string
  appConfigId: string
  appStackConfigHref: string
  status: ReactNode
  runs: string
  createdAt: string
  more: ReactNode
}

function parseInstallStackSummaryToTableData(
  stack: TInstallStack,
  orgId: string,
  appId: string
): TInstallStackRow[] {
  return stack?.versions.map((version) => {
    return {
      versionId: version?.id,
      appConfigId: version?.app_config_id,
      appStackConfigHref: `/${orgId}/apps/${appId}/configs/${version?.app_config_id}/stack`,
      status: (
        <Status variant="badge" status={version.composite_status?.status} />
      ),
      runs: version?.runs?.length?.toString() || '-',
      createdAt: version?.created_at,
      more: (
        <Dropdown
          id={`stack-${version.id}`}
          icon=""
          alignment="right"
          buttonClassName="!p-1"
          buttonText={<Icon variant="DotsThree" />}
        >
          <Menu>
            <Modal
              size="3/4"
              heading="View stack links"
              triggerButton={{
                children: (
                  <>
                    View links <Icon variant="Link" />
                  </>
                ),
                isMenuButton: true,
                variant: 'ghost',
              }}
            >
              <StackLinks
                quick_link_url={version?.quick_link_url}
                template_url={version?.template_url}
              />
            </Modal>
            <Modal
              size="3/4"
              heading="View stack outputs"
              triggerButton={{
                children: (
                  <>
                    View outputs <Icon variant="CodeBlock" />
                  </>
                ),
                isMenuButton: true,
                variant: 'ghost',
              }}
            >
              <StackOutputs runs={version?.runs} />
            </Modal>
          </Menu>
        </Dropdown>
      ),
    }
  })
}

const columns: ColumnDef<TInstallStackRow>[] = [
  {
    accessorKey: 'versionId',
    header: 'Version',
    cell: (info) => <ID>{info.getValue<string>()}</ID>,
    enableSorting: true,
  },
  {
    accessorKey: 'appConfigId',
    header: 'App version',
    cell: (info) => (
      <Text variant="subtext">
        <Link href={info?.row?.original?.appStackConfigHref}>
          {info.getValue<string>()}
        </Link>
      </Text>
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'status',
    header: 'Status',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    enableSorting: true,
    accessorKey: 'runs',
    header: 'Runs',
    cell: (info) => info.getValue<string>(),
  },
  {
    accessorKey: 'createdAt',
    header: 'Created',
    cell: (info) => (
      <Time
        time={info.getValue() as string}
        variant="subtext"
        format="relative"
      />
    ),
    enableSorting: true,
  },

  {
    accessorKey: 'more',
    header: '',
    id: 'more-options',
    cell: (info) => info.getValue() as string,
    enableSorting: true,
  },
]

export const InstallStacksTable = ({
  stack: initStack,
  pagination,
  pollInterval = 20000,
  shouldPoll,
}: {
  stack: TInstallStack
  pagination: IPagination
} & IPollingProps) => {
  const searchParams = useSearchParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const queryParams = useQueryParams({
    offset: pagination?.offset,
    limit: pagination?.limit,
    q: searchParams.get('q'),
  })
  const { data: stack } = usePolling<TInstallStack>({
    initData: initStack,
    path: `/api/orgs/${org.id}/installs/${install.id}/stack${queryParams}`,
    pollInterval,
    shouldPoll,
  })
  return (
    <Table<TInstallStackRow>
      columns={columns}
      data={parseInstallStackSummaryToTableData(stack, org.id, install.app_id)}
      emptyMessage="No stack found"
      pagination={pagination}
      searchPlaceholder="Search stack version..."
    />
  )
}

export const InstallStacksTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
