import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import type { TApp } from '@/types/ctl-api.types'
import { UserPreferencesProvider } from '../../../providers/user-preferences-provider'
import { AppsTable } from './AppsTable'

const APP: TApp = {
  id: 'app_payments',
  name: 'Payments API',
  status_v2: { status: 'active' },
  runner_config: { cloud_platform: 'aws' },
  config_repo: 'example/payments',
  updated_at: '2026-09-05T20:00:00Z',
}

const renderTable = (
  overrides: Partial<React.ComponentProps<typeof AppsTable>> = {}
) => {
  const onOffsetChange = () => {}
  render(
    <MemoryRouter>
      <UserPreferencesProvider>
        <AppsTable
          apps={[APP]}
          orgId="org_example"
          search=""
          onSearchChange={() => {}}
          offset={0}
          pageSize={20}
          hasNext={false}
          onOffsetChange={onOffsetChange}
          {...overrides}
        />
      </UserPreferencesProvider>
    </MemoryRouter>
  )
}

afterEach(() => {
  cleanup()
  window.localStorage.clear()
})

describe('AppsTable', () => {
  test('renders API app data as a linked table row', () => {
    renderTable()

    expect(screen.getByRole('link', { name: 'Payments API' })).toHaveAttribute(
      'href',
      '/org_example/apps/app_payments'
    )
    expect(screen.getByText('Active')).not.toBeNull()
    expect(screen.getByText('example/payments')).not.toBeNull()
  })

  test('moves to the next API offset', () => {
    let offset = 0
    renderTable({
      hasNext: true,
      onOffsetChange: (next) => {
        offset = next
      },
    })

    fireEvent.click(screen.getByRole('button', { name: 'Next' }))
    expect(offset).toBe(20)
  })

  test('renders the request failure in the empty collection', () => {
    renderTable({ apps: [], error: new globalThis.Error('Request failed') })
    expect(screen.getByText('Apps failed to load')).not.toBeNull()
  })
})
