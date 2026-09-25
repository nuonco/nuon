import { afterEach, expect, test } from 'bun:test'
import { cleanup, render } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { BranchCards } from './BranchCards'
import { mockBranchCards } from './BranchCards.fixtures'

afterEach(cleanup)

const renderCards = (layout?: 'grid' | 'stacked') =>
  render(
    <MemoryRouter>
      <BranchCards cards={mockBranchCards} layout={layout} />
    </MemoryRouter>
  )

test('defaults to the two column grid', () => {
  const { container } = renderCards()
  expect(container.querySelector('.lg\\:grid-cols-2')).toBeTruthy()
})

test('stacks cards full width when layout is stacked', () => {
  const { container } = renderCards('stacked')
  expect(container.querySelector('.lg\\:grid-cols-2')).toBeNull()
})
