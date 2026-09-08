import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { AppBranchOverviewCards } from './AppBranchOverviewCards'

afterEach(cleanup)

const BRANCH = {
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
}

describe('AppBranchOverviewCards', () => {
  test('surfaces the branch name, repository, and config number', () => {
    render(<AppBranchOverviewCards branch={BRANCH} installCount={12} />)

    expect(screen.getByText('main')).toBeTruthy()
    expect(screen.getByText('acme/payments')).toBeTruthy()
    expect(screen.getByText(/\bv14\b/)).toBeTruthy()
  })

  test('reads the config from the highest config number', () => {
    render(
      <AppBranchOverviewCards
        branch={{
          ...BRANCH,
          configs: [
            { config_number: 9, connected_github_vcs_config: { repo: 'acme/old' } },
            { config_number: 21, connected_github_vcs_config: { repo: 'acme/new' } },
          ],
        }}
        installCount={0}
      />
    )

    expect(screen.getByText('acme/new')).toBeTruthy()
    expect(screen.queryByText('acme/old')).toBeNull()
  })

  test('abbreviates the commit sha', () => {
    render(<AppBranchOverviewCards branch={BRANCH} installCount={12} />)

    expect(screen.getByText('a1b2c3d')).toBeTruthy()
  })

  test('omits the commit when the branch has never run', () => {
    render(
      <AppBranchOverviewCards
        branch={{ id: 'br_main', name: 'main' }}
        installCount={0}
      />
    )

    expect(screen.queryByText('a1b2c3d')).toBeNull()
  })

  test('marks a capped install count as a lower bound', () => {
    const view = render(
      <AppBranchOverviewCards branch={BRANCH} installCount={12} />
    )

    expect(screen.getByText('12')).toBeTruthy()

    view.rerender(
      <AppBranchOverviewCards branch={BRANCH} installCount={12} hasMoreInstalls />
    )

    expect(screen.getByText('12+')).toBeTruthy()
  })
})
