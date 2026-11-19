export default {
  title: 'Playground / Table v2 / Team',
}

import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { StoryFrame, useMockPage } from './story-utils'
import { TableV2 } from '@/components/common/TableV2'

type TMockMember = {
  email: string
  id: string
  isSelf: boolean
  name: string
  role: string
}

const FIRST_NAMES = [
  'avery',
  'blake',
  'casey',
  'devon',
  'emery',
  'finley',
  'harper',
  'jordan',
  'kendall',
  'logan',
  'morgan',
  'parker',
  'quinn',
]
const ROLES = ['Admin', 'Developer', 'Viewer', 'Developer', 'Admin']

const mockMembers: TMockMember[] = Array.from({ length: 26 }, (_, i) => {
  const name =
    i < FIRST_NAMES.length
      ? FIRST_NAMES[i]
      : `${FIRST_NAMES[i - FIRST_NAMES.length]}.${i}`
  return {
    email: `${name}@example.com`,
    id: `acct${String(i + 1).padStart(3, '0')}example`,
    isSelf: i === 0,
    name,
    role: i === 0 ? 'Admin' : ROLES[i % ROLES.length],
  }
})

const ActionCell = ({ member }: { member: TMockMember }) => (
  <Dropdown
    id={`action-${member.id}`}
    buttonText={<Icon variant="DotsThreeIcon" size={20} weight="bold" />}
    hideIcon
    variant="ghost"
    buttonClassName="!p-1"
    alignment="right"
  >
    <Menu>
      {!member.isSelf ? (
        <Button isMenuButton type="button">
          Change role
        </Button>
      ) : null}
      <Button isMenuButton type="button" variant="danger">
        Remove user
      </Button>
    </Menu>
  </Dropdown>
)

const columns: ColumnDef<TMockMember>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    enableHiding: false,
    cell: (props) => (
      <Text variant="body" weight="strong">
        {props.getValue<string>()}
      </Text>
    ),
  },
  {
    accessorKey: 'email',
    header: 'Email',
    cell: (props) => (
      <Text variant="body" className="text-primary-600 dark:text-primary-400">
        {props.getValue<string>()}
      </Text>
    ),
  },
  {
    accessorKey: 'role',
    header: 'Role',
    cell: (props) => (
      <Text variant="body" className="text-primary-600 dark:text-primary-400">
        {props.getValue<string>()}
      </Text>
    ),
  },
  {
    id: 'status',
    header: 'Status',
    enableSorting: false,
    meta: { card: 'status' },
    cell: () => (
      <Status status="active" variant="badge">
        Joined
      </Status>
    ),
  },
  {
    id: 'action',
    header: 'Action',
    enableSorting: false,
    enableHiding: false,
    meta: { card: 'actions' },
    cell: (props) => <ActionCell member={props.row.original} />,
  },
]

const PAGE_SIZE = 20

const emptyStateProps = {
  emptyTitle: 'No team members',
  emptyMessage: 'No team members found.',
}

export const Default = () => {
  const { data, pagination } = useMockPage(mockMembers, PAGE_SIZE)
  return (
    <StoryFrame note="Mirrors the org team table, which has no search box. Status is pinned to the top of each card and the row menu sits beside it. The first member is the current user, so their menu has no change-role option.">
      <TableV2
        tableId="playground-team"
        columns={columns}
        data={data}
        enableSearch={false}
        pagination={pagination}
        emptyStateProps={emptyStateProps}
      />
    </StoryFrame>
  )
}

export const Loading = () => (
  <StoryFrame>
    <TableV2
      tableId="playground-team"
      columns={columns}
      data={[]}
      enableSearch={false}
      isLoading
      pagination={{ limit: PAGE_SIZE, offset: 0 }}
    />
  </StoryFrame>
)

export const Empty = () => (
  <StoryFrame>
    <TableV2
      tableId="playground-team"
      columns={columns}
      data={[]}
      enableSearch={false}
      emptyStateProps={emptyStateProps}
    />
  </StoryFrame>
)
