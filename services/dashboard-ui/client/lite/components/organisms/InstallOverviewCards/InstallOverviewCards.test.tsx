import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { InstallOverviewCards } from './InstallOverviewCards'

afterEach(cleanup)

describe('InstallOverviewCards', () => {
  test('renders health, drift, branch, and commit summaries', () => {
    render(
      <InstallOverviewCards
        install={{
          id: 'inst_prod',
          composite_health_status: 'healthy',
          drifted_objects: [{ target_id: 'cmp_api' }],
          app_branch: {
            id: 'br_main',
            name: 'main',
            configs: [{ config_number: 14 }],
          },
        }}
        latestSync={{
          id: 'sync_latest',
          created_at: '2026-09-08T13:45:00Z',
          app_branch_id: 'br_main',
          triggered_by: 'git',
          vcs_connection_commit: {
            sha: 'a1b2c3d4e5f6',
            message: 'Update component versions',
          },
        }}
      />
    )

    expect(screen.getByText('Healthy')).toBeTruthy()
    expect(screen.getByText('Drift detected')).toBeTruthy()
    expect(screen.getByText('1 resource drifted')).toBeTruthy()
    expect(screen.getByText('main')).toBeTruthy()
    expect(screen.getByText('a1b2c3d')).toBeTruthy()
  })

  test('distinguishes missing evaluations from healthy state', () => {
    render(
      <InstallOverviewCards
        install={{ id: 'inst_prod', app_branch: { id: 'br_main', name: 'main' } }}
      />
    )

    expect(screen.getByText('Not reported')).toBeTruthy()
    expect(screen.getByText('Not scanned')).toBeTruthy()
  })
})
