import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { Duration } from './Duration'

afterEach(cleanup)

describe('Duration', () => {
  test('formats nanoseconds without discarding a valid zero', () => {
    const { rerender } = render(<Duration nanoseconds={65_000_000_000} />)
    expect(screen.getByText('1m 5s')).toBeTruthy()

    rerender(<Duration nanoseconds={0} />)
    expect(screen.getByText('0s')).toBeTruthy()
  })

  test('formats a fixed timestamp range', () => {
    const { container } = render(
      <Duration
        start="2026-09-05T20:00:00Z"
        end="2026-09-05T21:32:18Z"
        format="long"
      />
    )
    expect(container.querySelector('time')?.textContent).toBe(
      '1 hour, 32 minutes, 18 seconds'
    )
  })

  test('renders clock-style output in timer format', () => {
    render(<Duration nanoseconds={93_784_000_000_000} format="timer" />)
    expect(screen.getByText('1d 02:03:04')).toBeTruthy()
  })

  test('uses semantic duration markup and full tooltip context', () => {
    const { container } = render(
      <Duration nanoseconds={93_784_000_000_000} />
    )
    const time = container.querySelector('time')
    expect(time?.getAttribute('datetime')).toMatch(/^P/)

    fireEvent.mouseEnter(screen.getByText('1d 2h'))
    expect(screen.getByRole('tooltip').textContent).toBe(
      '1 day, 2 hours, 3 minutes, 4 seconds'
    )
  })

  test('renders fallback for missing, negative, and invalid input', () => {
    const { rerender } = render(<Duration fallback="Not recorded" />)
    expect(screen.getByText('Not recorded')).toBeTruthy()

    rerender(<Duration nanoseconds={-1} fallback="Invalid duration" />)
    expect(screen.getByText('Invalid duration')).toBeTruthy()

    rerender(
      <Duration
        start="invalid"
        end="also-invalid"
        fallback="Invalid range"
      />
    )
    expect(screen.getByText('Invalid range')).toBeTruthy()
  })

  test('uses Text loading at the requested width', () => {
    const { container } = render(<Duration loading loadingWidth={9} />)
    expect(
      (container.querySelector('.skeleton-text') as HTMLElement).style.width
    ).toBe('9ch')
  })
})
