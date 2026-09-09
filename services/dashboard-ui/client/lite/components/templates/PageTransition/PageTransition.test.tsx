import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { PageTransition } from './PageTransition'

afterEach(cleanup)

test('renders page content inside a single transition boundary', () => {
  const { container } = render(
    <PageTransition>
      <h1>Install overview</h1>
    </PageTransition>
  )

  const boundary = container.querySelector('[data-page-transition]')

  expect(container.querySelectorAll('[data-page-transition]')).toHaveLength(1)
  expect(boundary?.contains(screen.getByRole('heading'))).toBe(true)
})

test('stays layout neutral for the shells it renders into', () => {
  const { container } = render(
    <PageTransition className="gap-6">
      <span>Page</span>
    </PageTransition>
  )

  const boundary = container.querySelector('[data-page-transition]')

  expect(boundary?.className).toBe('flex w-full min-w-0 flex-col gap-6')
})
