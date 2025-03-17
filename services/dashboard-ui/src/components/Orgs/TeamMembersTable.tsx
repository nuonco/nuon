'use client'

import React, { type FC, useMemo } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { Table } from '@/components/DataTable'
import { Time } from '@/components/Time'
import { Text } from '@/components/Typography'
import type { TAccount } from '@/types'

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
    ],
    []
  )

  return <Table columns={columns} data={members} />
}
