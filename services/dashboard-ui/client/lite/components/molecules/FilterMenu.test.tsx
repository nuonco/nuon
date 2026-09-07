import { afterEach, expect, mock, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { FilterOption } from './FilterMenu'

afterEach(cleanup)

const renderOption = ({
  checked = true,
  isolated = false,
}: {
  checked?: boolean
  isolated?: boolean
} = {}) => {
  const onToggle = mock()
  const onIsolate = mock()

  render(
    <FilterOption
      checked={checked}
      isolated={isolated}
      onToggle={onToggle}
      onIsolate={onIsolate}
      label="Update"
      textValue="Update"
      tabIndex={0}
    />
  )

  return {
    checkbox: screen.getByRole('checkbox', { name: 'Include Update' }),
    onIsolate,
    onToggle,
    row: screen.getByRole('group', { name: /Update, included/ }),
  }
}

test('checkbox click toggles without isolating', () => {
  const { checkbox, onIsolate, onToggle } = renderOption()

  fireEvent.click(checkbox)

  expect(onToggle).toHaveBeenCalledTimes(1)
  expect(onIsolate).not.toHaveBeenCalled()
})

test('row click isolates without toggling', () => {
  const { onIsolate, onToggle, row } = renderOption()

  fireEvent.click(row)

  expect(onIsolate).toHaveBeenCalledTimes(1)
  expect(onToggle).not.toHaveBeenCalled()
})

test('Space toggles and Enter isolates', () => {
  const { onIsolate, onToggle, row } = renderOption()

  fireEvent.keyDown(row, { key: ' ' })
  fireEvent.keyDown(row, { key: 'Enter' })

  expect(onToggle).toHaveBeenCalledTimes(1)
  expect(onIsolate).toHaveBeenCalledTimes(1)
})

test('isolated row exposes reset behavior', () => {
  renderOption({ isolated: true })

  expect(
    screen.getByRole('group', { name: /Enter resets filters/ })
  ).toBeInTheDocument()
  expect(screen.getByText('Reset')).toBeInTheDocument()
})
