import { afterEach, expect, mock, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { useState } from 'react'
import { SearchInput } from './SearchInput'

afterEach(cleanup)

const Example = ({ onKeyDown = mock() }) => {
  const [value, setValue] = useState('payments')
  return (
    <SearchInput
      value={value}
      onValueChange={setValue}
      aria-label="Search resources"
      onKeyDown={onKeyDown}
    />
  )
}

test('clear button empties the query', () => {
  render(<Example />)

  fireEvent.click(screen.getByRole('button', { name: 'Clear search' }))

  expect((screen.getByRole('searchbox') as HTMLInputElement).value).toBe('')
  expect(screen.queryByRole('button', { name: 'Clear search' })).toBeNull()
})

test('Escape empties the query before delegating the key', () => {
  const onKeyDown = mock()
  render(<Example onKeyDown={onKeyDown} />)

  fireEvent.keyDown(screen.getByRole('searchbox'), { key: 'Escape' })

  expect((screen.getByRole('searchbox') as HTMLInputElement).value).toBe('')
  expect(onKeyDown).not.toHaveBeenCalled()
})

test('delegates other keys to the caller', () => {
  const onKeyDown = mock()
  render(<Example onKeyDown={onKeyDown} />)

  fireEvent.keyDown(screen.getByRole('searchbox'), { key: 'ArrowDown' })

  expect(onKeyDown).toHaveBeenCalledTimes(1)
})
