import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { InstallOverviewCards } from './InstallOverviewCards'

afterEach(cleanup)

const INSTALL = {
  id: 'inst_prod',
  composite_health_status: 'healthy',
  drifted_objects: [{ target_id: 'cmp_api' }],
  app_branch: {
    id: 'br_main',
    name: 'main',
    configs: [{ config_number: 14 }],
  },
}

const SYNC = {
  id: 'sync_latest',
  created_at: '2026-09-08T13:45:00Z',
  app_branch_id: 'br_main',
  triggered_by: 'git',
  vcs_connection_commit: {
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

  test('reads the commit from the latest config sync', () => {
    const view = render(<InstallOverviewCards install={INSTALL} />)

    expect(screen.queryByText('a1b2c3d')).toBeNull()

    view.rerender(<InstallOverviewCards install={INSTALL} latestSync={SYNC} />)

    expect(screen.getByText('a1b2c3d')).toBeTruthy()
  })

  test('surfaces the branch name and config number', () => {
    render(<InstallOverviewCards install={INSTALL} latestSync={SYNC} />)

    expect(screen.getByText('main')).toBeTruthy()
    expect(screen.getByText(/\bv14\b/)).toBeTruthy()
  })
})
