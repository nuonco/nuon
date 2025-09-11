'use client'

import React, { type FC, useMemo, useState } from 'react'
import { createPortal } from 'react-dom'
import { type ColumnDef } from '@tanstack/react-table'
import { UserMinus, TrashSimple } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Table } from '@/components/DataTable'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Time } from '@/components/Time'
import { Text } from '@/components/Typography'
import { removeUserFromOrg } from '@/components/org-actions'
import type { TAccount } from '@/types'
import { useOrg } from '@/hooks/use-org'

interface ITeamMembersTable {
  members: Array<TAccount>
}

export const TeamMembersTable: FC<ITeamMembersTable> = ({ members }) => {
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
        cell: (props) => <RemoveUserModal user={props?.row.original} />,
      },
    ],
    []
  )

  return <Table columns={columns} data={members} />
}

const RemoveUserModal = ({ user }: { user: TAccount }) => {
  const { org } = useOrg()
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string>()

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
                {error ? <Notice>{error}</Notice> : null}
                <Text>
                  Are you sure you want to remove {user?.email} from your org?
                </Text>
              </div>
              <div className="p-6 border-t flex gap-3 justify-end">
                <Button
                  className="text-sm"
                  onClick={() => {
                    setError(undefined)
                    setIsLoading(false)
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
                    setIsLoading(true)
                    removeUserFromOrg({ user_id: user?.id }, org.id)
                      .then(() => {
                        setIsLoading(false)
                        setIsOpen(false)
                      })
                      .catch((err) => {
                        console.error(err)
                        setIsLoading(false)
                        setError(
                          'Unable to invite user, refresh page and try again.'
                        )
                      })
                  }}
                >
                  {isLoading ? (
                    <>
                      <SpinnerSVG /> Removing user...
                    </>
                  ) : (
                    <>
                      <UserMinus size="18" /> remove user
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
        <TrashSimple size="18" />
      </Button>
    </>
  )
}
