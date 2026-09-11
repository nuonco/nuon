import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { InstallOverviewCards } from './InstallOverviewCards'

afterEach(cleanup)

const INSTALL = {
  id: 'inst_prod',
  composite_health_status: 'healthy',
  lifecycle_phase: { phase: 'active' },
  runner_status: 'active',
  sandbox_status: 'active',
  composite_component_status: 'active',
  drifted_objects: [{ target_id: 'cmp_api' }],
  app_branch: {
    id: 'br_main',
    name: 'main',
  },
}

const UPDATE = {
  runId: 'run_01k4m8f6a9',
  branchName: 'main',
  status: 'success',
  updatedAt: '2026-09-08T13:45:00Z',
  commit: {
    sha: 'a1b2c3d4e5f6',
    message: 'Update component versions',
  },
}

describe('InstallOverviewCards', () => {
  test('surfaces the health status reported by the install', () => {
    const view = render(<InstallOverviewCards install={INSTALL} />)

    expect(screen.getByText('Healthy')).toBeTruthy()

    view.rerender(
      <InstallOverviewCards
        install={{ ...INSTALL, composite_health_status: 'degraded' }}
      />
    )

    expect(screen.getByText('Degraded')).toBeTruthy()
    expect(screen.queryByText('Healthy')).toBeNull()
  })

  test('pluralizes the drifted resource count', () => {
    const view = render(<InstallOverviewCards install={INSTALL} />)

    expect(screen.getByText('1 resource drifted')).toBeTruthy()

    view.rerender(
      <InstallOverviewCards
        install={{
          ...INSTALL,
          drifted_objects: [{ target_id: 'cmp_api' }, { target_id: 'cmp_worker' }],
        }}
      />
    )

    expect(screen.getByText('2 resources drifted')).toBeTruthy()
  })

  test('claims neither health nor drift before the first evaluation', () => {
    render(
      <InstallOverviewCards
        install={{ id: 'inst_prod', app_branch: { id: 'br_main', name: 'main' } }}
      />
    )

    expect(screen.queryByText('Healthy')).toBeNull()
    expect(screen.queryByText(/drifted/)).toBeNull()
  })

  test('reads the expected commit from the latest branch update', () => {
    const view = render(<InstallOverviewCards install={INSTALL} />)

    expect(screen.queryByText('a1b2c3d')).toBeNull()

    view.rerender(
      <InstallOverviewCards install={INSTALL} lastBranchUpdate={UPDATE} />
    )

    expect(screen.getByText('a1b2c3d')).toBeTruthy()
    expect(screen.getByText('run_01k4m8f6a9')).toBeTruthy()
  })

  test('keeps current services as separate status facets', () => {
    render(<InstallOverviewCards install={INSTALL} lastBranchUpdate={UPDATE} />)

    expect(screen.getByText('Runner')).toBeTruthy()
    expect(screen.getByText('Sandbox')).toBeTruthy()
    expect(screen.getByText('Components')).toBeTruthy()
  })
})
