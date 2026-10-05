import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { CommitLink, PullRequestLink } from './GitReferenceLink'

afterEach(cleanup)

const sha = 'e5aef07c91b24d0a8f3310c0e5aef07c91b24d0a'

test('links a short commit label to the full commit', () => {
  render(<CommitLink sha={sha} repo="acme/platform" />)
  const link = screen.getByRole('link', { name: `Commit ${sha}` })
  expect(link.getAttribute('href')).toBe(
    `https://github.com/acme/platform/commit/${sha}`
  )
  expect(link.textContent).toContain(sha.slice(0, 7))
  expect(link.getAttribute('target')).toBe('_blank')
})

test('renders an unlinked commit when the repository is missing', () => {
  render(<CommitLink sha={sha} />)
  expect(screen.queryByRole('link')).toBeNull()
  expect(screen.getByLabelText(`Commit ${sha}`).textContent).toContain(
    sha.slice(0, 7)
  )
})

test('prefers an explicit pull request URL', () => {
  render(
    <PullRequestLink
      number={104}
      repo="acme/platform"
      href="https://gitlab.com/acme/platform/-/merge_requests/104"
    />
  )
  expect(screen.getByRole('link', { name: 'Pull request 104' }).getAttribute('href')).toBe(
    'https://gitlab.com/acme/platform/-/merge_requests/104'
  )
})
