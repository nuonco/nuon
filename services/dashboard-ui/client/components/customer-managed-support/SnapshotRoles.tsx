import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel } from '@/components/surfaces/Panel'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import type { TCustomerManagedCapturedRole } from '@/lib/ctl-api/installs/customer-managed-support-snapshots'
import { CustomerManagedSnapshotContent } from './SnapshotEmpty'

const panelLinkClass =
  '!p-0 !h-auto !border-none !rounded-none !bg-transparent hover:!bg-transparent active:!bg-transparent focus:!shadow-none text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600'

const RoleDetail = ({
  role,
  observedAt,
}: {
  role: TCustomerManagedCapturedRole
  observedAt: string
}) => (
  <div className="flex flex-col gap-4">
    <Card>
      <Text weight="strong">Summary</Text>
      <div className="grid grid-cols-2 gap-6">
        <LabeledValue label="Name">{role.name}</LabeledValue>
        <LabeledValue label="Type">
          <Badge variant="code" size="sm">
            {role.type}
          </Badge>
        </LabeledValue>
        <LabeledValue label="Status">
          <Status status={role.provisioned ? 'active' : 'inactive'}>
            {role.provisioned ? 'Provisioned' : 'Not provisioned'}
          </Status>
        </LabeledValue>
        <LabeledValue label="Observed">
          <Time time={observedAt} format="long-datetime" />
        </LabeledValue>
      </div>
      <LabeledValue label="Cloud identifier">
        <ID>{role.cloud_id}</ID>
      </LabeledValue>
    </Card>
    <Card>
      <HeadingGroup>
        <Text weight="strong">Policies and usage</Text>
        <Text variant="subtext" theme="neutral">
          Policies and runtime usage were not included in this support snapshot.
        </Text>
      </HeadingGroup>
    </Card>
  </div>
)

export const CustomerManagedSnapshotRoles = () => {
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const captured = snapshot?.snapshot?.roles
  const roles = captured?.roles ?? []
  const columns = useMemo<ColumnDef<TCustomerManagedCapturedRole, unknown>[]>(
    () => [
      {
        accessorKey: 'name',
        header: 'Name',
        cell: ({ row }) => (
          <Panel
            size="3/4"
            panelKey={`captured-role-${row.original.type}-${row.original.name}`}
            heading={
              <div className="flex flex-col">
                <Text variant="h3">{row.original.name}</Text>
                <Text variant="subtext" theme="neutral" weight="normal">
                  Captured {row.original.type} role
                </Text>
              </div>
            }
            triggerButton={{
              variant: 'ghost',
              className: panelLinkClass,
              children: row.original.name,
            }}
          >
            <RoleDetail
              role={row.original}
              observedAt={captured?.observed_at ?? ''}
            />
          </Panel>
        ),
      },
      {
        accessorKey: 'type',
        header: 'Type',
        cell: ({ row }) => (
          <Badge variant="code" theme="neutral">
            {row.original.type}
          </Badge>
        ),
      },
      {
        accessorKey: 'provisioned',
        header: 'Status',
        cell: ({ row }) => (
          <Status status={row.original.provisioned ? 'active' : 'inactive'}>
            {row.original.provisioned ? 'Provisioned' : 'Not provisioned'}
          </Status>
        ),
      },
      {
        accessorKey: 'cloud_id',
        header: 'Cloud identifier',
        cell: ({ row }) => <ID>{row.original.cloud_id}</ID>,
      },
      {
        id: 'actions',
        header: '',
        enableSorting: false,
        cell: ({ row }) => (
          <Panel
            size="3/4"
            panelKey={`captured-role-${row.original.type}-${row.original.name}-action`}
            heading={<Text variant="h3">{row.original.name}</Text>}
            triggerButton={{
              variant: 'ghost',
              className: panelLinkClass,
              children: (
                <span className="flex items-center gap-1.5">
                  View <Icon variant="CaretRightIcon" />
                </span>
              ),
            }}
          >
            <RoleDetail
              role={row.original}
              observedAt={captured?.observed_at ?? ''}
            />
          </Panel>
        ),
      },
    ],
    [captured?.observed_at]
  )

  return (
    <CustomerManagedSnapshotContent>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Install roles
          </Text>
          <Text variant="subtext" theme="neutral">
            Cloud role identifiers captured from the customer stack outputs.
          </Text>
        </HeadingGroup>
        {captured?.observed_at ? (
          <span className="flex items-center gap-2">
            <Text variant="subtext" theme="neutral">
              Observed
            </Text>
            <Time
              time={captured.observed_at}
              format="relative"
              variant="subtext"
            />
          </span>
        ) : null}
      </div>

      {!captured ? (
        <EmptyState
          variant="table"
          emptyTitle="No roles captured"
          emptyMessage="Roles will appear after the customer captures a snapshot with compatible stack outputs."
        />
      ) : (
        <Table
          columns={columns}
          data={roles}
          searchPlaceholder="Search by name or cloud identifier..."
          emptyMessage="No roles found"
        />
      )}
    </CustomerManagedSnapshotContent>
  )
}
