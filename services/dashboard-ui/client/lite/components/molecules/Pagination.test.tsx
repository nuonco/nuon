import { afterEach, expect, mock, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { Pagination } from './Pagination'

afterEach(cleanup)

test('disables previous on the first page', () => {
  render(
    <Pagination offset={0} pageSize={20} hasNext onOffsetChange={mock()} />
  )

  expect(screen.getByText('Page 1')).toBeTruthy()
  expect(
    screen
      .getByRole('button', { name: 'Previous' })
      .getAttribute('aria-disabled')
  ).toBe('true')
  expect(
    screen.getByRole('button', { name: 'Next' }).hasAttribute('aria-disabled')
  ).toBe(false)
})

test('moves by one configured page', () => {
  const onOffsetChange = mock()
  render(
    <Pagination
      offset={20}
      pageSize={20}
      hasNext
      onOffsetChange={onOffsetChange}
    />
  )

  fireEvent.click(screen.getByRole('button', { name: 'Previous' }))
  fireEvent.click(screen.getByRole('button', { name: 'Next' }))

  expect(onOffsetChange).toHaveBeenNthCalledWith(1, 0)
  expect(onOffsetChange).toHaveBeenNthCalledWith(2, 40)
  expect(screen.getByText('Page 2')).toBeTruthy()
})

test('disables next on the final page', () => {
  render(
    <Pagination
      offset={40}
      pageSize={20}
      hasNext={false}
      onOffsetChange={mock()}
    />
  )

  expect(
    screen.getByRole('button', { name: 'Next' }).getAttribute('aria-disabled')
  ).toBe('true')
})

test('disables both actions while loading', () => {
  render(
    <Pagination
      offset={20}
      pageSize={20}
      hasNext
      loading
      onOffsetChange={mock()}
    />
  )

  expect(screen.getByRole('navigation').getAttribute('aria-busy')).toBe('true')
  expect(
    screen
      .getByRole('button', { name: 'Previous' })
      .getAttribute('aria-disabled')
  ).toBe('true')
  expect(
    screen.getByRole('button', { name: 'Next' }).getAttribute('aria-disabled')
  ).toBe('true')
})
