import { useMemo, type ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { Dropdown } from '@/components/common/Dropdown'
import type { IEmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Menu } from '@/components/common/Menu'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type {
  TServiceAccount,
  TServiceAccountManagement,
  TServiceAccountOwnership,
} from '@/types'
import { humanize } from '@/utils/string-utils'
import { ChangeServiceAccountRoleButton } from '@/components/service-accounts/ChangeServiceAccountRole'
import { RenameServiceAccountButton } from '@/components/service-accounts/RenameServiceAccount'
import { CreateServiceAccountTokenButton } from '@/components/service-accounts/ServiceAccountToken'
import { DeleteServiceAccountButton } from '@/components/service-accounts/DeleteServiceAccount'

export type TServiceAccountOwner = {
  name: string
  kind: string
  href?: string
}

export type TServiceAccountRoleLabel = {
  key: string
  title: string
}

export type TServiceAccountRow = {
  id: string
  name: string
  identity: string
  hasName: boolean
  purposes: string[]
  isLegacy: boolean
  owner: TServiceAccountOwner | null
  roles: TServiceAccountRoleLabel[]
  createdAt: string
  account: TServiceAccount
}

const OWNER_KINDS: Record<string, string> = {
  orgs: 'Org',
  installs: 'Install',
  install_stacks: 'Install stack',
}

export function serviceAccountOwner(
  ownership: TServiceAccountOwnership | null | undefined,
  orgId: string
): TServiceAccountOwner | null {
  if (!ownership) return null

  const ownerType = ownership.owner_type ?? ''
  const ownerId = ownership.owner_id ?? ''
  const kind =
    OWNER_KINDS[ownerType] ?? (ownerType ? humanize(ownerType) : 'Resource')
  const name = ownership.owner_name || ownerId
  const installHref = (installId: string, suffix = '') =>
    `/${orgId}/installs/${installId}${suffix}`

  switch (ownerType) {
    case 'orgs':
      return { kind, name, href: `/${orgId}/settings/general` }
    case 'installs':
      return { kind, name, href: ownerId ? installHref(ownerId) : undefined }
    case 'install_stacks':
      return {
        kind,
        name,
        href: ownership.install_id
          ? installHref(ownership.install_id, '/stacks')
          : undefined,
      }
    default:
      return {
        kind,
        name,
        href: ownership.install_id
          ? installHref(ownership.install_id)
          : undefined,
      }
  }
}

export function serviceAccountRoleLabels(
  account: TServiceAccount
): TServiceAccountRoleLabel[] {
  const seen = new Set<string>()

  return (account.roles ?? []).reduce<TServiceAccountRoleLabel[]>(
    (labels, role, idx) => {
      const roleType = role.role_type ?? ''
      const key = role.id || `${roleType}-${idx}`
      if (seen.has(key)) return labels
      seen.add(key)

      const title =
        role.title || (roleType ? humanize(roleType) : 'Custom role')

      labels.push({ key, title })
      return labels
    },
    []
  )
}

export function parseServiceAccountsToTableData(
  accounts: TServiceAccount[],
  orgId: string
): TServiceAccountRow[] {
  return accounts.map((account) => {
    const identity = account.email || account.id || ''
    return {
      id: account.id || '',
      name: account.name || identity,
      identity,
      hasName: !!account.name && account.name !== identity,
      purposes: account.purposes ?? [],
      isLegacy: !!account.system_account && !account.managed_service_account,
      owner: serviceAccountOwner(account.managed_service_account, orgId),
      roles: serviceAccountRoleLabels(account),
      createdAt: account.created_at || '',
      account,
    }
  })
}

const EmptyValue = () => (
  <Text variant="body" theme="neutral">
    —
  </Text>
)

const AccountCell = ({ row }: { row: TServiceAccountRow }) =>
  row.hasName ? (
    <div className="flex min-w-0 max-w-96 flex-col gap-1">
      <Text
        family="mono"
        variant="body"
        weight="strong"
        className="break-words"
      >
        {row.name}
      </Text>
      <Text
        family="mono"
        variant="subtext"
        theme="neutral"
        className="break-all"
      >
        <ClickToCopy>{row.identity}</ClickToCopy>
      </Text>
    </div>
  ) : (
    <Text family="mono" variant="body" className="max-w-96 break-all">
      <ClickToCopy>{row.identity}</ClickToCopy>
    </Text>
  )

const PurposeCell = ({ row }: { row: TServiceAccountRow }) =>
  row.purposes.length || row.isLegacy ? (
    <div className="flex flex-wrap items-center gap-1.5">
      {row.purposes.map((purpose) => (
        <Badge key={purpose} theme="default" size="sm">
          {humanize(purpose)}
        </Badge>
      ))}
      {row.isLegacy ? (
        <Badge theme="default" size="sm">
          Legacy
        </Badge>
      ) : null}
    </div>
  ) : (
    <EmptyValue />
  )

const ResourceCell = ({ owner }: { owner: TServiceAccountOwner | null }) =>
  owner ? (
    <div className="flex min-w-0 flex-col gap-1">
      {owner.href ? (
        <Link
          href={owner.href}
          variant="inline"
          className="font-mono break-all"
        >
          {owner.name}
        </Link>
      ) : (
        <Text family="mono" variant="body" className="break-all">
          {owner.name}
        </Text>
      )}
      <Text variant="subtext" theme="neutral">
        {owner.kind}
      </Text>
    </div>
  ) : (
    <EmptyValue />
  )

const RolesCell = ({ roles }: { roles: TServiceAccountRoleLabel[] }) =>
  roles.length ? (
    <div className="flex max-w-72 flex-wrap items-center gap-1.5">
      {roles.map((role) => (
        <Badge key={role.key} theme="default" size="sm">
          {role.title}
        </Badge>
      ))}
    </div>
  ) : (
    <EmptyValue />
  )

const disabledMenuProps = (reason: string) => ({
  disabled: true,
  className: '!p-2 w-full justify-between',
  tooltipProps: {
    className: 'block !w-full',
    position: 'left' as const,
    tipContent: reason,
  },
})

const ActionCell = ({ row }: { row: TServiceAccountRow }) => {
  const { account, owner } = row
  const ownerKind = owner?.kind.toLowerCase()
  const lockedProps = owner
    ? disabledMenuProps(`Managed by Nuon for its ${ownerKind}`)
    : {}

  return (
    <Dropdown
      id={`action-${account.id}`}
      buttonText={<Icon variant="DotsThreeIcon" size={20} weight="bold" />}
      aria-label={`Actions for ${row.name}`}
      hideIcon
      variant="ghost"
      buttonClassName="!p-1"
      alignment="right"
    >
      <Menu className="w-64">
        {owner ? (
          <div className="px-1.5 py-1">
            <Text variant="subtext" theme="neutral">
              Managed by Nuon for {ownerKind} {owner.name}. You can still create
              tokens.
            </Text>
          </div>
        ) : null}
        {owner ? <hr /> : null}
        <span>
          <RenameServiceAccountButton
            account={account}
            isMenuButton
            {...lockedProps}
          />
        </span>
        <span>
          <ChangeServiceAccountRoleButton
            account={account}
            isMenuButton
            {...lockedProps}
          />
        </span>
        <span>
          <CreateServiceAccountTokenButton account={account} isMenuButton />
        </span>
        <span>
          <DeleteServiceAccountButton
            account={account}
            isMenuButton
            {...lockedProps}
          />
        </span>
      </Menu>
    </Dropdown>
  )
}

export const SERVICE_ACCOUNTS_TABLE_LIMIT = 20

function serviceAccountsEmptyState(
  management: TServiceAccountManagement,
  hasActiveFilters: boolean
): IEmptyState {
  if (hasActiveFilters) {
    return {
      emptyTitle: 'No service accounts found',
      emptyMessage:
        'Change the search or filters to see more service accounts.',
    }
  }
  if (management === 'system') {
    return {
      emptyTitle: 'No system service accounts yet',
      emptyMessage:
        'System accounts will appear here when Nuon creates them for runners, stacks, and other resources.',
    }
  }
  return {
    emptyTitle: 'No service accounts yet',
    emptyMessage:
      'Create a service account to automate access to the Nuon API.',
  }
}

export interface IServiceAccountsTable {
  data: TServiceAccount[]
  orgId: string
  isLoading: boolean
  error?: string
  pagination: { hasNext: boolean; offset: number; limit: number }
  management?: TServiceAccountManagement
  hasActiveFilters?: boolean
  filterActions?: ReactNode
}

export const ServiceAccountsTable = ({
  data,
  orgId,
  isLoading,
  error,
  pagination,
  management = 'user',
  hasActiveFilters = false,
  filterActions,
}: IServiceAccountsTable) => {
  const columns: ColumnDef<TServiceAccountRow>[] = useMemo(
    () => [
      {
        header: 'Account',
        accessorKey: 'name',
        cell: (props) => <AccountCell row={props.row.original} />,
      },
      {
        id: 'purpose',
        header: 'Purpose',
        enableSorting: false,
        cell: (props) => <PurposeCell row={props.row.original} />,
      },
      {
        id: 'resource',
        header: 'Resource',
        enableSorting: false,
        cell: (props) => <ResourceCell owner={props.row.original.owner} />,
      },
      {
        id: 'roles',
        header: 'Roles',
        enableSorting: false,
        cell: (props) => <RolesCell roles={props.row.original.roles} />,
      },
      {
        header: 'Created',
        accessorKey: 'createdAt',
        cell: (props) => {
          const value = props.getValue<string>()
          return value ? (
            <Time time={value} format="relative" />
          ) : (
            <EmptyValue />
          )
        },
      },
      {
        id: 'action',
        header: 'Action',
        enableSorting: false,
        cell: (props) => <ActionCell row={props.row.original} />,
      },
    ],
    []
  )

  const rows = useMemo(
    () => parseServiceAccountsToTableData(data, orgId),
    [data, orgId]
  )

  if (error) {
    return (
      <div className="flex flex-col gap-4">
        {filterActions}
        <Banner theme="error" role="alert">
          <Text role="paragraph" weight="strong">
            Unable to load service accounts
          </Text>
          <Text role="paragraph">{error}</Text>
        </Banner>
      </div>
    )
  }

  return (
    <Table<TServiceAccountRow>
      columns={
        management === 'user'
          ? columns.filter(({ id }) => id !== 'purpose' && id !== 'resource')
          : columns
      }
      data={rows}
      isLoading={isLoading}
      pagination={pagination}
      enableSearch={false}
      filterActions={filterActions}
      emptyStateProps={serviceAccountsEmptyState(management, hasActiveFilters)}
    />
  )
}
