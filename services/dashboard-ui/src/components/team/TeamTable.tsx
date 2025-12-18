'use client'

import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import { useEffect, useMemo, useState } from 'react'
import { createPortal } from 'react-dom'
import type { ColumnDef } from '@tanstack/react-table'
import { DotsThree, UserMinus } from '@phosphor-icons/react'
import { removeUser } from '@/actions/orgs/remove-user'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Menu } from '@/components/common/Menu'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Modal as OldModal } from '@/components/old/Modal'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import type { TAccount } from '@/types'
import type { IPagination } from '@/components/common/Pagination'

export type TTeamMemberRow = {
  id: string
  name: string
  email: string
  role: string
  status: string
  account: TAccount
}

function parseAccountToTableData(members: TAccount[]): TTeamMemberRow[] {
  return members.map((member) => ({
    id: member.id || '',
    name: member.email?.split('@')[0] || 'Unknown',
    email: member.email || '',
    role: 'All permissions',
    status: 'active', // Maps to green/success theme
    account: member,
  }))
}

const ActionCell = ({
  account,
  currentMemberCount,
  limit,
}: {
  account: TAccount
  currentMemberCount: number
  limit: number
}) => {
  const path = usePathname()
  const router = useRouter()
  const searchParams = useSearchParams()
  const { org } = useOrg()

  const [isModalOpen, setIsModalOpen] = useState(false)
  const {
    data: removedAccount,
    error,
    execute,
    isLoading,
  } = useServerAction({
    action: removeUser,
  })

  useEffect(() => {
    if (removedAccount) {
      setIsModalOpen(false)

      const currentOffset = parseInt(searchParams.get('offset') || '0', 10)

      if (currentMemberCount === 1 && currentOffset > 0) {
        const previousOffset = Math.max(0, currentOffset - limit)

        const params = new URLSearchParams(searchParams.toString())
        if (previousOffset === 0) {
          params.delete('offset')
        } else {
          params.set('offset', previousOffset.toString())
        }

        const newUrl = `${path}?${params.toString()}`
        router.push(newUrl)
      }
    }
  }, [removedAccount, currentMemberCount, searchParams, path, router, limit])

  return (
    <>
      <Dropdown
        id={`action-${account.id}`}
        buttonText={<DotsThree size={20} weight="bold" />}
        hideIcon
        variant="ghost"
        buttonClassName="!p-1"
        alignment="right"
      >
        <Menu>
          <Button variant="danger" onClick={() => setIsModalOpen(true)}>
            <UserMinus size={16} />
            Remove user
          </Button>
        </Menu>
      </Dropdown>

      {isModalOpen
        ? createPortal(
            <OldModal
              className="!max-w-lg"
              contentClassName="!p-0"
              heading="Remove team member"
              isOpen={isModalOpen}
              onClose={() => setIsModalOpen(false)}
            >
              <div className="p-6 flex flex-col gap-4">
                <Text>
                  Are you sure you want to remove{' '}
                  <strong>{account.email}</strong> from your organization?
                </Text>
                {error && (
                  <Text theme="error">
                    {error?.error || 'Unable to remove user from organization'}
                  </Text>
                )}
              </div>
              <div className="p-6 border-t flex gap-3 justify-end">
                <Button variant="secondary" onClick={() => setIsModalOpen(false)}>
                  Cancel
                </Button>
                <Button
                  variant="danger"
                  onClick={() => {
                    execute({
                      body: { user_id: account.id },
                      orgId: org.id,
                      path,
                    })
                  }}
                  disabled={isLoading}
                >
                  {isLoading ? 'Removing...' : 'Remove user'}
                </Button>
              </div>
            </OldModal>,
            document.body
          )
        : null}
    </>
  )
}

export const TeamTable = ({
  members,
  pagination,
}: {
  members: TAccount[]
  pagination: IPagination
}) => {
  const columns: ColumnDef<TTeamMemberRow>[] = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'name',
        cell: (props) => (
          <Text variant="body" weight="strong">
            {props.getValue<string>()}
          </Text>
        ),
      },
      {
        header: 'Email',
        accessorKey: 'email',
        cell: (props) => (
          <Text variant="body" className="text-primary-600 dark:text-primary-400">
            {props.getValue<string>()}
          </Text>
        ),
      },
      {
        header: 'Role',
        accessorKey: 'role',
        cell: (props) => (
          <Text variant="body" className="text-primary-600 dark:text-primary-400">
            {props.getValue<string>()}
          </Text>
        ),
      },
      {
        header: 'Status',
        accessorKey: 'status',
        cell: (props) => (
          <Status status={props.getValue<string>()} variant="badge">
            Joined
          </Status>
        ),
      },
      {
        id: 'action',
        header: 'Action',
        cell: (props) => (
          <ActionCell
            account={props.row.original.account}
            currentMemberCount={members.length}
            limit={pagination?.limit || 10}
          />
        ),
      },
    ],
    [members.length, pagination?.limit]
  )

  return (
    <Table<TTeamMemberRow>
      columns={columns}
      data={parseAccountToTableData(members)}
      pagination={pagination}
      searchPlaceholder="Search..."
      emptyStateProps={{
        emptyTitle: 'No team members',
        emptyMessage: 'No team members found.',
      }}
    />
  )
}

const skeletonColumns: ColumnDef<TTeamMemberRow>[] = [
  { header: 'Name', accessorKey: 'name' },
  { header: 'Email', accessorKey: 'email' },
  { header: 'Role', accessorKey: 'role' },
  { header: 'Status', accessorKey: 'status' },
  { header: 'Action', id: 'action' },
]

export const TeamTableSkeleton = () => (
  <TableSkeleton<TTeamMemberRow> columns={skeletonColumns} skeletonRows={5} />
)
