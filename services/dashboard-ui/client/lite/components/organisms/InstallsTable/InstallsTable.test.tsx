import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import type { TInstall } from '@/types/ctl-api.types'
import { UserPreferencesProvider } from '../../../providers/user-preferences-provider'
import { columnsFor, InstallsTable, type IInstallFilter } from './InstallsTable'

const INSTALL: TInstall = {
  id: 'inst_production',
  name: 'Production',
  app_id: 'app_payments',
  app: { id: 'app_payments', name: 'Payments API' },
  runner_status: 'offline',
  sandbox_status: 'queued',
  composite_component_status: 'active',
  cloud_platform: 'aws',
  aws_account: { region: 'us-west-2' },
  app_branch: { id: 'branch_main', name: 'main' },
  labels: { env: 'production' },
  updated_at: '2026-09-05T20:00:00Z',
}

const filter = (overrides: Partial<IInstallFilter> = {}): IInstallFilter => ({
  label: 'Labels',
  options: [{ value: 'env:production', label: 'env:production' }],
  selected: new Set(),
  onToggle: () => {},
  onIsolate: () => {},
  onReset: () => {},
  constrained: false,
  ...overrides,
})

const renderTable = (
  overrides: Partial<React.ComponentProps<typeof InstallsTable>> = {}
) => {
  render(
    <MemoryRouter>
      <UserPreferencesProvider>
        <InstallsTable
          installs={[INSTALL]}
          orgId="org_example"
          search=""
          onSearchChange={() => {}}
          offset={0}
          pageSize={20}
          hasNext={false}
          onOffsetChange={() => {}}
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

const facetTooltip = (label: string) => {
  const marker = screen.getByLabelText(label)
  fireEvent.mouseEnter(marker)
  const id = marker.getAttribute('aria-describedby')!
  return document.getElementById(id)?.textContent
}

const NARROWEST_DESKTOP_VIEWPORT = 1280
const DESKTOP_SIDEBAR_WIDTH = 224
const SHELL_HORIZONTAL_PADDING = 32
const NARROWEST_DESKTOP_TABLE_WIDTH =
  NARROWEST_DESKTOP_VIEWPORT - DESKTOP_SIDEBAR_WIDTH - SHELL_HORIZONTAL_PADDING

describe('InstallsTable', () => {
  test('fits the desktop shell so table view stays available', () => {
    const declaredWidth = columnsFor('org_example').reduce(
      (total, column) => total + (column.size ?? 150),
      0
    )

    expect(declaredWidth).toBeLessThanOrEqual(NARROWEST_DESKTOP_TABLE_WIDTH)
  })

  test('renders API install data and linked entities', () => {
    renderTable()

    expect(screen.getByRole('link', { name: 'Production' })).toHaveAttribute(
      'href',
      '/org_example/installs/inst_production'
    )
    expect(screen.getByRole('link', { name: 'Payments API' })).toHaveAttribute(
      'href',
      '/org_example/apps/app_payments'
    )
    expect(screen.getByText('US West (Oregon)')).not.toBeNull()
    expect(screen.getByText('main')).not.toBeNull()
  })

  test('reports each status axis separately instead of one rollup', () => {
    renderTable()

    expect(screen.getByLabelText('Runner offline')).not.toBeNull()
    expect(screen.getByLabelText('Sandbox queued')).not.toBeNull()
    expect(screen.getByLabelText('Components active')).not.toBeNull()
  })

  test('omits health and drift until the API reports them', () => {
    renderTable()

    expect(screen.queryByLabelText(/^Health/)).toBeNull()
    expect(screen.queryByLabelText(/drift/i)).toBeNull()
  })

  test('renders health and drift axes once present', () => {
    renderTable({
      installs: [
        {
          ...INSTALL,
          composite_health_status: 'degraded',
          drifted_objects: [{ target_id: 'cmp_api' }],
        },
      ],
    })

    expect(screen.getByLabelText('Health degraded')).not.toBeNull()
    expect(screen.getByLabelText('Drift detected')).not.toBeNull()
  })

  test('reports a deprovisioned install rather than a stale active axis', () => {
    renderTable({
      installs: [
        {
          ...INSTALL,
          runner_status: 'active',
          sandbox_status: 'active',
          composite_component_status: 'active',
          lifecycle_phase: { phase: 'deprovisioned' },
        },
      ],
    })

    expect(screen.getByLabelText('Runner deprovisioned')).not.toBeNull()
    expect(screen.getByLabelText('Sandbox deprovisioned')).not.toBeNull()
    expect(screen.getByLabelText('Components deprovisioned')).not.toBeNull()
  })

  test('describes the sandbox axis from its own status description', () => {
    renderTable({
      installs: [
        {
          ...INSTALL,
          sandbox_status_description: 'Waiting on the runner to come online.',
        },
      ],
    })

    expect(facetTooltip('Sandbox queued')).toContain(
      'Waiting on the runner to come online.'
    )
  })

  test('prefers the health message once sandbox health overrides', () => {
    renderTable({
      installs: [
        {
          ...INSTALL,
          sandbox_status: 'active',
          sandbox_status_description: 'Sandbox is running.',
          sandbox_health_status: 'unhealthy',
          sandbox_health_message: 'Nodes stopped reporting health.',
        },
      ],
    })

    expect(facetTooltip('Sandbox unhealthy')).toContain(
      'Nodes stopped reporting health.'
    )
  })

  test('prefers sandbox health over a plain active sandbox', () => {
    renderTable({
      installs: [
        {
          ...INSTALL,
          sandbox_status: 'active',
          sandbox_health_status: 'unhealthy',
        },
      ],
    })

    expect(screen.getByLabelText('Sandbox unhealthy')).not.toBeNull()
  })

  test('renders install labels as badges in the table row', () => {
    renderTable()

    const badge = screen.getByText('env').parentElement!.parentElement!
    expect(badge.textContent).toBe('envproduction')
  })

  test('tints a label badge with its app color', () => {
    renderTable({ labelColors: { app_payments: { env: '#2563eb' } } })

    const value = screen.getByText('production').parentElement!
    expect(value.className).toContain('badge-custom')
    expect(value.getAttribute('style')).toContain('#2563eb')
  })

  test('binds filter controls to the supplied callbacks', () => {
    let selected = ''
    renderTable({
      labelFilter: filter({
        onToggle: (value) => {
          selected = value
        },
      }),
    })

    fireEvent.click(screen.getByRole('button', { name: 'Labels' }))
    fireEvent.click(
      screen.getByRole('checkbox', { name: 'Include env:production' })
    )
    expect(selected).toBe('env:production')
  })

  test('marks constrained filters in the toolbar', () => {
    renderTable({
      labelFilter: filter({
        selected: new Set(['env:production']),
        constrained: true,
      }),
    })

    expect(screen.getByRole('button', { name: 'Labels (1)' })).not.toBeNull()
  })
})
