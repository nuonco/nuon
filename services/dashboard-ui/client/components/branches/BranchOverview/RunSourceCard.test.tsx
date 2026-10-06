import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { RunSourceCard } from './RunSourceCard'

afterEach(cleanup)

const source = { kind: 'manual' as const }

test('shows a short commit message without a toggle', () => {
  render(
    <RunSourceCard
      source={source}
      title="Manual run"
      status="success"
      commit={{ message: 'feat: add cache component' }}
    />
  )

  expect(screen.getByText('feat: add cache component')).toBeTruthy()
  expect(screen.queryByRole('button', { name: 'Show more' })).toBeNull()
})

test('collapses a long commit message until it is expanded', () => {
  const message = [
    'feat: map preview install configs onto branch groups',
    '',
    '* rename the ramp config to Customer A',
    '* rename the sony config to Customer B',
    '* add Customer C install config',
    '* add Customer D install config',
    '* move prod install configs under install-configs/prod',
    '* pin AWS install configs to their app branches',
  ].join('\n')

  render(
    <RunSourceCard
      source={source}
      title="Manual run"
      status="error"
      commit={{ message }}
    />
  )

  const clip = screen.getByText(/map preview install configs/).parentElement
  expect(clip?.className).toContain('max-h-32')

  fireEvent.click(screen.getByRole('button', { name: 'Show more' }))

  expect(clip?.className).not.toContain('max-h-32')
  expect(screen.getByRole('button', { name: 'Show less' })).toBeTruthy()

  fireEvent.click(screen.getByRole('button', { name: 'Show less' }))

  expect(clip?.className).toContain('max-h-32')
})
