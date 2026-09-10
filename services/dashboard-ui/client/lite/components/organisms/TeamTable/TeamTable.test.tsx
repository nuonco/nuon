import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import type { TOrgMember, TRoleInfo } from '@/types/ctl-api.types'
import { UserPreferencesProvider } from '../../../providers/user-preferences-provider'
import { TeamTable, type ITeamFilter } from './TeamTable'

const ACTIVE: TOrgMember = {
  id: 'acct_active',
  account_id: 'acct_active',
  email: 'active@example.com',
  name: 'Active Member',
  status: 'active',
  role_type: 'org_admin',
  joined_at: '2026-08-01T12:00:00Z',
  created_at: '2026-07-01T12:00:00Z',
}

const INVITED: TOrgMember = {
  id: 'invite_pending',
  invite_id: 'invite_pending',
  email: 'invited@example.com',
  status: 'invited',
  role_type: 'org_read_only',
  joined_at: '2026-06-01T12:00:00Z',
  created_at: '2026-09-01T12:00:00Z',
}

const ROLES: TRoleInfo[] = [
  { id: 'role_admin', role_type: 'org_admin', title: 'Org admin' },
  { id: 'role_read', role_type: 'org_read_only', title: 'Read only' },
]

const filter = (
  label: string,
  overrides: Partial<ITeamFilter> = {}
): ITeamFilter => ({
  label,
  options: [{ value: 'active', label: 'Active' }],
  selected: new Set(),
  onToggle: () => {},
  onIsolate: () => {},
  onReset: () => {},
  constrained: false,
  ...overrides,
})

const renderTable = (
  overrides: Partial<React.ComponentProps<typeof TeamTable>> = {}
) =>
  render(
    <UserPreferencesProvider>
      <TeamTable
        members={[ACTIVE, INVITED]}
        roles={ROLES}
        currentAccountId="acct_current"
        admin
        search=""
        onSearchChange={() => {}}
        offset={0}
        pageSize={20}
        hasNext={false}
        onOffsetChange={() => {}}
        statusFilter={filter('Status')}
        roleFilter={filter('Role')}
        onChangeRole={() => {}}
        onRemove={() => {}}
        onResend={() => {}}
        onRevoke={() => {}}
        {...overrides}
      />
    </UserPreferencesProvider>
  )

const openActions = (email: string) =>
  fireEvent.click(screen.getByRole('button', { name: `${email} actions` }))

afterEach(() => {
  cleanup()
  window.localStorage.clear()
})

describe('TeamTable', () => {
  test('renders one row per member keyed by member id', () => {
    renderTable()

    expect(
      screen.getByText('active@example.com').closest('tr')?.dataset.rowId
    ).toBe('acct_active')
    expect(
      screen.getByText('invited@example.com').closest('tr')?.dataset.rowId
    ).toBe('invite_pending')
  })

  test('shows active actions only on an active row', () => {
    renderTable({ members: [ACTIVE] })
    openActions('active@example.com')

    expect(screen.getByRole('menuitem', { name: 'Change role' })).not.toBeNull()
    expect(
      screen.getByRole('menuitem', { name: 'Remove from org' })
    ).not.toBeNull()
    expect(screen.queryByRole('menuitem', { name: 'Resend invite' })).toBeNull()
    expect(screen.queryByRole('menuitem', { name: 'Revoke invite' })).toBeNull()
  })

  test('shows invite actions only on an invited row', () => {
    renderTable({ members: [INVITED] })
    openActions('invited@example.com')

    expect(
      screen.getByRole('menuitem', { name: 'Resend invite' })
    ).not.toBeNull()
    expect(
      screen.getByRole('menuitem', { name: 'Revoke invite' })
    ).not.toBeNull()
    expect(screen.queryByRole('menuitem', { name: 'Change role' })).toBeNull()
    expect(
      screen.queryByRole('menuitem', { name: 'Remove from org' })
    ).toBeNull()
  })

  test('calls the selected row action with its member', () => {
    let selected = ''
    renderTable({
      members: [ACTIVE],
      onChangeRole: (member) => {
        selected = member.id ?? ''
      },
    })
    openActions('active@example.com')
    fireEvent.click(screen.getByRole('menuitem', { name: 'Change role' }))

    expect(selected).toBe('acct_active')
  })

  test('hides change role on the current account row', () => {
    renderTable({ members: [ACTIVE], currentAccountId: 'acct_active' })
    openActions('active@example.com')

    expect(screen.queryByRole('menuitem', { name: 'Change role' })).toBeNull()
    expect(
      screen.getByRole('menuitem', { name: 'Remove from org' })
    ).not.toBeNull()
  })

  test('hides only admin-gated actions from a non-admin', () => {
    renderTable({ members: [ACTIVE], admin: false })
    openActions('active@example.com')

    expect(screen.queryByRole('menuitem', { name: 'Change role' })).toBeNull()
    expect(
      screen.getByRole('menuitem', { name: 'Remove from org' })
    ).not.toBeNull()

    cleanup()
    renderTable({ members: [INVITED], admin: false })
    openActions('invited@example.com')

    expect(
      screen.getByRole('menuitem', { name: 'Resend invite' })
    ).not.toBeNull()
    expect(screen.queryByRole('menuitem', { name: 'Revoke invite' })).toBeNull()
  })

  test('shows role titles and falls back to role type', () => {
    renderTable({ members: [ACTIVE] })
    expect(screen.getByText('Org admin')).not.toBeNull()

    cleanup()
    renderTable({ members: [ACTIVE], roles: [] })
    expect(screen.getByText('org_admin')).not.toBeNull()
  })

  test('uses invite creation time instead of joined time', () => {
    renderTable({ members: [INVITED] })

    const time = document.querySelector('time')
    expect(time?.getAttribute('datetime')).toContain('2026-09-01T12:00:00')
    expect(time?.getAttribute('datetime')).not.toContain('2026-06-01T12:00:00')
    expect(screen.getAllByText('Invited')).toHaveLength(2)
  })

  test('renders different empty and failed messages', () => {
    renderTable({ members: [] })
    expect(screen.getByText('No team members yet')).not.toBeNull()

    cleanup()
    renderTable({ members: [], error: new Error('Request failed') })
    expect(screen.getByText('Team members failed to load')).not.toBeNull()
  })

  test('calls search and filter handlers', async () => {
    let search = ''
    let status = ''
    renderTable({
      onSearchChange: (value) => {
        search = value
      },
      statusFilter: filter('Status', {
        onToggle: (value) => {
          status = value
        },
      }),
    })

    fireEvent.change(
      screen.getByRole('searchbox', { name: 'Search team members' }),
      {
        target: { value: 'example' },
      }
    )
    await new Promise((resolve) => setTimeout(resolve, 350))
    expect(search).toBe('example')

    fireEvent.click(screen.getByRole('button', { name: 'Status' }))
    fireEvent.click(screen.getByRole('checkbox', { name: 'Include Active' }))
    expect(status).toBe('active')
  })
})
