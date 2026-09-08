import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { AppBranchOverviewCards } from './AppBranchOverviewCards'

afterEach(cleanup)

describe('AppBranchOverviewCards', () => {
  test('renders branch, commit, and install summaries', () => {
    render(
      <AppBranchOverviewCards
        branch={{
          id: 'br_main',
          name: 'main',
          configs: [
            {
              config_number: 14,
              connected_github_vcs_config: { repo: 'acme/payments' },
            },
          ],
          latest_run: {
            vcs_connection_commit: {
              sha: 'a1b2c3d4e5f6',
              message: 'Update component versions',
            },
          },
        }}
        installCount={12}
      />
    )

    expect(screen.getByText('Branch')).toBeTruthy()
    expect(screen.getByText('main')).toBeTruthy()
    expect(screen.getByText('Config v14')).toBeTruthy()
    expect(screen.getByText('a1b2c3d')).toBeTruthy()
    expect(screen.getByText('12')).toBeTruthy()
  })

  test('renders empty commit state', () => {
    render(
      <AppBranchOverviewCards
        branch={{ id: 'br_main', name: 'main' }}
        installCount={0}
      />
    )

    expect(screen.getByText('No commits yet')).toBeTruthy()
    expect(screen.getByText('0')).toBeTruthy()
  })
})
