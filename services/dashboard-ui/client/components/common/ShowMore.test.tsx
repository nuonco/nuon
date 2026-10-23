import { afterAll, afterEach, beforeAll, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
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
  render(<ShowMore>short</ShowMore>)

  expect(screen.getByText('short')).toBeTruthy()
  expect(screen.queryByRole('button')).toBeNull()
})

test('clips overflowing children until expanded', () => {
  render(<ShowMore>{'x'.repeat(200)}</ShowMore>)

  expect(screen.getByRole('button', { name: /Show more/ })).toBeTruthy()

  fireEvent.click(screen.getByRole('button', { name: /Show more/ }))

  expect(screen.getByRole('button', { name: /Show less/ })).toBeTruthy()
})

test('respects a controlled expanded prop', () => {
  const onExpandedChange = () => {}

  render(
    <ShowMore expanded onExpandedChange={onExpandedChange}>
      {'x'.repeat(200)}
    </ShowMore>,
  )

  expect(screen.getByRole('button', { name: /Show less/ })).toBeTruthy()
})

test('swaps the icon set via the icons prop', () => {
  render(<ShowMore icons="plusMinus">{'x'.repeat(200)}</ShowMore>)

  expect(document.querySelector('svg')).toBeTruthy()
})
