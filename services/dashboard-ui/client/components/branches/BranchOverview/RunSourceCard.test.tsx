import { afterAll, afterEach, beforeAll, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { RunCommitSummary, RunSourceCard } from './RunSourceCard'

const originalScrollHeight = Object.getOwnPropertyDescriptor(
  HTMLElement.prototype,
  'scrollHeight',
)
const originalClientHeight = Object.getOwnPropertyDescriptor(
  HTMLElement.prototype,
  'clientHeight',
)

beforeAll(() => {
  Object.defineProperty(HTMLElement.prototype, 'scrollHeight', {
    configurable: true,
    get(this: HTMLElement) {
      return this.textContent?.length ?? 0
    },
  })
  Object.defineProperty(HTMLElement.prototype, 'clientHeight', {
    configurable: true,
    get() {
      return 50
    },
  })
})

afterAll(() => {
  if (originalScrollHeight) {
    Object.defineProperty(HTMLElement.prototype, 'scrollHeight', originalScrollHeight)
  }
  if (originalClientHeight) {
    Object.defineProperty(HTMLElement.prototype, 'clientHeight', originalClientHeight)
  }
})

afterEach(cleanup)

const source = { kind: 'manual' as const }

test('labels the branch run time instead of leaving it beside the commit author', () => {
  render(
    <RunCommitSummary
      source={source}
      title="Manual run"
      commit={{
        message: 'feat: add cache',
        author: 'Casey',
        createdAt: '2026-11-01T12:00:00Z',
      }}
      runAt="2026-11-08T15:00:00Z"
    />,
  )

  expect(screen.getByText('Branch run was')).toBeTruthy()
  expect(screen.getByText(/ago$/)).toBeTruthy()
  expect(screen.queryByText('Commit was')).toBeNull()
})

test('shows the git commit time beside the branch run time', () => {
  render(
    <RunCommitSummary
      source={source}
      title="Manual run"
      commit={{
        message: 'feat: add cache',
        committedAt: '2026-11-01T12:00:00Z',
      }}
      runAt="2026-11-08T15:00:00Z"
    />,
  )

  expect(screen.getByText('Commit was')).toBeTruthy()
  expect(screen.getByText('Branch run was')).toBeTruthy()
  expect(screen.getAllByText(/ago$/)).toHaveLength(2)
})

test('shows a short commit message without a toggle', () => {
  render(
    <RunSourceCard
      source={source}
      title="Manual run"
      status="success"
      commit={{ message: 'feat: add cache component' }}
    />,
  )

  expect(screen.getByText('feat: add cache component')).toBeTruthy()
  expect(screen.queryByRole('button', { name: 'Show more' })).toBeNull()
})

test('collapses a long commit message until it is expanded', () => {
  const message = [
    'feat: map preview install configs onto branch groups',
    '',
    '* rename the first config to Customer A',
    '* rename the second config to Customer B',
    '* add Customer C install config',
    '* add Customer D install config',
    '* move prod install configs under install-configs/prod',
    '* pin AWS install configs to their app branches',
  ].join('\n')

  render(
    <RunSourceCard source={source} title="Manual run" status="error" commit={{ message }} />,
  )

  const clip = screen.getByText(/map preview install configs/).parentElement
  expect(clip?.style.maxHeight).toBeTruthy()
  expect(clip?.className).toContain('overflow-hidden')
  expect(clip?.className).toContain('mask-image')

  fireEvent.click(screen.getByRole('button', { name: 'Show more' }))

  expect(clip?.style.maxHeight).toBeFalsy()
  expect(clip?.className).not.toContain('overflow-hidden')
  expect(screen.getByRole('button', { name: 'Show less' })).toBeTruthy()

  fireEvent.click(screen.getByRole('button', { name: 'Show less' }))

  expect(clip?.style.maxHeight).toBeTruthy()
})

test('collapses a short commit body with a co-author line', () => {
  const message = [
    'chore: bump dependencies',
    '',
    '* chore: update lockfile',
    'Co-authored-by: Example Dev <dev@example.com>',
  ].join('\n')

  render(
    <RunSourceCard source={source} title="Manual run" status="success" commit={{ message }} />,
  )

  const clip = screen.getByText(/chore: bump dependencies/).parentElement
  expect(clip?.style.maxHeight).toBeTruthy()
  expect(screen.getByRole('button', { name: 'Show more' })).toBeTruthy()
})
