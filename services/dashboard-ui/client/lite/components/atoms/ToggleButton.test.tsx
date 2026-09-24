import { afterEach, describe, expect, mock, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { ToggleButton } from './ToggleButton'

afterEach(cleanup)

describe('ToggleButton', () => {
  test('renders a labelled group with one pressed option', () => {
    render(
      <ToggleButton
        label="Time range"
        value="week"
        onValueChange={() => {}}
        options={[
          { value: 'day', label: 'Day' },
          { value: 'week', label: 'Week' },
          { value: 'month', label: 'Month' },
        ]}
      />
    )

    expect(screen.getByRole('group', { name: 'Time range' })).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Day' })).toHaveAttribute(
      'aria-pressed',
      'false'
    )
    expect(screen.getByRole('button', { name: 'Week' })).toHaveAttribute(
      'aria-pressed',
      'true'
    )
  })

  test('reports selection changes', () => {
    const onValueChange = mock(() => {})
    render(
      <ToggleButton
        label="View"
        value="table"
        onValueChange={onValueChange}
        options={[
          { value: 'table', label: 'Table' },
          { value: 'cards', label: 'Cards' },
        ]}
      />
    )

    fireEvent.click(screen.getByRole('button', { name: 'Cards' }))
    expect(onValueChange).toHaveBeenCalledWith('cards')
  })

  test('does not select a disabled option', () => {
    const onValueChange = mock(() => {})
    render(
      <ToggleButton
        label="View"
        value="table"
        onValueChange={onValueChange}
        options={[
          { value: 'table', label: 'Table' },
          { value: 'cards', label: 'Cards', disabled: true },
        ]}
      />
    )

    fireEvent.click(screen.getByRole('button', { name: 'Cards' }))
    expect(onValueChange).not.toHaveBeenCalled()
  })
})
