import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { BranchCard, type TBranchCardData } from './BranchCard'

afterEach(cleanup)

const card: TBranchCardData = {
  branchId: 'branch-main',
  name: 'main',
  href: '/org-1/apps/app-1/branches/branch-main',
  latestRun: {
    href: '/org-1/apps/app-1/branches/branch-main/runs/run-1',
    status: 'success',
    sha: 'abc1234',
    commitMessage: 'latest update',
  },
  planGroups: [{ name: 'enterprise', installs: 1, hasSelector: false }],
  latestRunInstalls: [
    {
      id: 'install-1',
      name: 'production-acme',
      group: 'enterprise',
      runStatus: 'success',
      health: 'healthy',
      rolledOut: true,
    },
  ],
}

test('shows the latest run deployment bars expanded by default', () => {
  render(
    <MemoryRouter>
      <BranchCard card={card} />
    </MemoryRouter>
  )

  expect(screen.getByText('latest update')).toBeTruthy()
  expect(screen.getByRole('link', { name: 'View run' })).toBeTruthy()
  expect(screen.getByRole('button', { expanded: true })).toBeTruthy()
  expect(screen.getByLabelText('production-acme: Success')).toBeTruthy()
})

test('shows pull request metadata for the latest run', () => {
  render(
    <MemoryRouter>
      <BranchCard
        card={{
          ...card,
          latestRun: {
            ...card.latestRun!,
            trigger: 'pull_request',
            prNumber: 142,
          },
        }}
      />
    </MemoryRouter>
  )

  expect(screen.getByText('PR #142')).toBeTruthy()
})

test('shows the tag when a run was triggered by a tag', () => {
  render(
    <MemoryRouter>
      <BranchCard
        card={{
          ...card,
          latestRun: { ...card.latestRun!, trigger: 'tag', tag: 'v1.42.0' },
        }}
      />
    </MemoryRouter>
  )

  expect(screen.getByText('v1.42.0')).toBeTruthy()
})

test('marks a preview plan group', () => {
  render(
    <MemoryRouter>
      <BranchCard
        card={{
          ...card,
          planGroups: [
            {
              name: 'pr-previews',
              installs: 1,
              hasSelector: false,
              isPreview: true,
            },
          ],
          latestRunInstalls: [
            {
              id: 'install-preview-142',
              name: 'preview-142',
              group: 'pr-previews',
              runStatus: 'success',
              rolledOut: true,
            },
          ],
        }}
      />
    </MemoryRouter>
  )

  expect(screen.getAllByText('Preview').length).toBeGreaterThan(0)
})
