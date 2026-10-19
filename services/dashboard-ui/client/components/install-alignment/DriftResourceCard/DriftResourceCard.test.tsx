import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import {
  inputUpdateResource,
  unappliedResource,
} from '@/components/__fixtures__/install-alignment'
import { DriftResourceCard } from './DriftResourceCard'

afterEach(cleanup)

test('cancelled input changes keep details and their workflow link even when the last attempt is the evidence', () => {
  render(
    <MemoryRouter>
      <DriftResourceCard
        resource={inputUpdateResource()}
        explanationHref="/workflows/input-update?panel=deploy"
        lastAttemptHref="/workflows/input-update"
      />
    </MemoryRouter>
  )

  expect(
    screen.getByText('Saved input changes haven’t taken effect.')
  ).toBeTruthy()
  expect(screen.getByText('Workflow cancelled')).toBeTruthy()
  const details = screen.getByRole('button', {
    name: 'Saved input changes haven’t taken effect.',
  })
  expect(details.getAttribute('aria-expanded')).toBe('false')
  fireEvent.click(details)
  expect(details.getAttribute('aria-expanded')).toBe('true')
  expect(
    document.getElementById(details.getAttribute('aria-controls')!)
  ).toBeTruthy()
  expect(
    screen.getAllByText('Saved input changes haven’t taken effect.')
  ).toHaveLength(1)
  expect(screen.queryByText('What hasn’t been applied')).toBeNull()
  expect(
    screen.queryByRole('button', { name: /View details|Hide details/ })
  ).toBeNull()
  expect(
    screen.getByText(
      'Applying the input changes was cancelled. The saved values haven’t taken effect.'
    )
  ).toBeTruthy()
  expect(
    screen.getByRole('link', { name: 'View workflow' }).getAttribute('href')
  ).toBe('/workflows/input-update?panel=deploy')
  expect(screen.getByText(/Started/)).toBeTruthy()
  fireEvent.click(details)
  expect(screen.queryByRole('link', { name: 'View workflow' })).toBeNull()
})

test('unattributed changes explain the missing evidence without inventing a workflow', () => {
  render(
    <DriftResourceCard
      resource={unappliedResource({
        explanation: undefined,
        last_attempt: undefined,
        plan: undefined,
        newer_branch_runs: null,
      })}
    />
  )
  fireEvent.click(
    screen.getByRole('button', {
      name: 'This component hasn’t been deployed yet.',
    })
  )
  expect(
    screen.getByText(
      'We couldn’t determine why these changes haven’t taken effect.'
    )
  ).toBeTruthy()
  expect(screen.queryByRole('link')).toBeNull()
  expect(screen.queryByText('Workflow cancelled')).toBeNull()
})
