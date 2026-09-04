import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { Select } from './Select'

afterEach(cleanup)

const OPTIONS = [
  { value: 'payments', label: 'Payments' },
  { value: 'billing', label: 'Billing' },
  { value: 'search', label: 'Search' },
]

const openSelect = () => {
  render(
    <Select options={OPTIONS} searchable searchPlaceholder="Find option" />
  )
  const trigger = screen.getByRole('button', { name: /Select an option/ })
  fireEvent.click(trigger)

  return {
    search: screen.getByRole('searchbox', { name: 'Find option' }),
    trigger,
  }
}

test('filters options as the query changes', () => {
  const { search } = openSelect()

  fireEvent.change(search, { target: { value: 'bill' } })

  expect(!!screen.queryByRole('option', { name: 'Billing' })).toBe(true)
  expect(!!screen.queryByRole('option', { name: 'Payments' })).toBe(false)
})

test('Escape clears the query before it closes the dropdown', () => {
  const { search, trigger } = openSelect()
  fireEvent.change(search, { target: { value: 'bill' } })

  fireEvent.keyDown(search, { key: 'Escape' })

  expect((search as HTMLInputElement).value).toBe('')
  expect(trigger.getAttribute('aria-expanded')).toBe('true')

  fireEvent.keyDown(search, { key: 'Escape' })

  expect(trigger.getAttribute('aria-expanded')).toBe('false')
})

test('the clear button leaves the dropdown open', () => {
  const { search, trigger } = openSelect()
  fireEvent.change(search, { target: { value: 'bill' } })

  fireEvent.click(screen.getByRole('button', { name: 'Clear search' }))

  expect((search as HTMLInputElement).value).toBe('')
  expect(trigger.getAttribute('aria-expanded')).toBe('true')
  expect(!!screen.queryByRole('option', { name: 'Payments' })).toBe(true)
})
