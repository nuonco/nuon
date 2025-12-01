'use client'

import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import { useEffect, useMemo, useState } from 'react'
import { createPortal } from 'react-dom'
import { type ColumnDef } from '@tanstack/react-table'
import { UserMinusIcon, TrashSimpleIcon } from '@phosphor-icons/react'
import { removeUser } from '@/actions/orgs/remove-user'
import { Button } from '@/components/old/Button'
import { Table } from '@/components/old/DataTable'
import { SpinnerSVG } from '@/components/old/Loading'
import { Modal } from '@/components/old/Modal'
import { Notice } from '@/components/old/Notice'
import { Time } from '@/components/old/Time'
import { Text } from '@/components/old/Typography'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import type { TAccount } from '@/types'

export const TeamMembersTable = ({
  members,
  limit = 10,
}: {
  members: TAccount[]
  limit?: number
}) => {
  const columns: Array<ColumnDef<TAccount>> = useMemo(
    () => [
      {
        header: 'Member email',
        accessorKey: 'email',
        cell: (props) => <Text>{props.getValue<string>()}</Text>,
      },

      {
        header: 'Joined',
        accessorKey: 'created_at',
        cell: (props) => <Time time={props.getValue<string>()} />,
      },

      {
        id: 'remove',
        cell: (props) => (
          <RemoveUserModal
            user={props?.row.original}
            currentMemberCount={members.length}
            limit={limit}
          />
        ),
      },
    ],
    [members.length, limit]
  )

  return <Table columns={columns} data={members} />
}

const RemoveUserModal = ({
  user,
  currentMemberCount,
  limit,
}: {
  user: TAccount
  currentMemberCount: number
  limit: number
}) => {
  const path = usePathname()
  const router = useRouter()
  const searchParams = useSearchParams()
  const { org } = useOrg()

  const [isOpen, setIsOpen] = useState(false)
  const {
    data: account,
    error,
    execute,
    isLoading,
  } = useServerAction({
    action: removeUser,
  })

  useEffect(() => {
    if (error) {
    }

    if (account) {
      setIsOpen(false)

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
  }, [account, error, currentMemberCount, searchParams, path, router, limit])

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-2xl"
              contentClassName="!p-0"
              heading="Invite team member"
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="p-6 flex flex-col gap-4">
                {error ? (
                  <Notice>
                    {error?.error || 'Unable to remove user from organization'}
                  </Notice>
                ) : null}
                <Text>
                  Are you sure you want to remove {user?.email} from your org?
                </Text>
              </div>
              <div className="p-6 border-t flex gap-3 justify-end">
                <Button
                  className="text-sm"
                  onClick={() => {
                    setIsOpen(false)
                  }}
                  type="button"
                >
                  Cancel
                </Button>
                <Button
                  className="text-sm flex items-center gap-2 font-medium"
                  disabled={isLoading}
                  variant="danger"
                  onClick={() => {
                    execute({
                      body: { user_id: user.id },
                      orgId: org.id,
                      path,
                    })
                  }}
                >
                  {isLoading ? (
                    <>
                      <SpinnerSVG /> Removing user...
                    </>
                  ) : (
                    <>
                      <UserMinusIcon size="18" /> remove user
                    </>
                  )}
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm flex !p-2 self-end !border-none"
        onClick={() => {
          setIsOpen(true)
        }}
        variant="caution"
      >
        <TrashSimpleIcon size="18" />
      </Button>
    </>
  )
}
