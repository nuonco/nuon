import { afterAll, afterEach, beforeAll, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, within } from '@testing-library/react'
import { ShowMore } from './ShowMore'

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

test('renders short children with no toggle', () => {
  const { container } = render(<ShowMore>short</ShowMore>)

  expect(within(container).getByText('short')).toBeTruthy()
  expect(within(container).queryByRole('button')).toBeNull()
})

test('clips overflowing children until expanded', () => {
  const { container } = render(<ShowMore>{'x'.repeat(200)}</ShowMore>)

  fireEvent.click(within(container).getByRole('button', { name: /Show more/ }))

  expect(within(container).getByRole('button', { name: /Show less/ })).toBeTruthy()
})

test('respects a controlled expanded prop', () => {
  const { container } = render(
    <ShowMore expanded onExpandedChange={() => {}}>
      {'x'.repeat(200)}
    </ShowMore>,
  )

  expect(within(container).getByRole('button', { name: /Show less/ })).toBeTruthy()
})

test('swaps the icon set via the icons prop', () => {
  const { container } = render(
    <ShowMore icons="plusMinus">{'x'.repeat(200)}</ShowMore>,
  )

  expect(container.querySelector('svg')).toBeTruthy()
})
