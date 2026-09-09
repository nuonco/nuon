import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { Time } from './Time'

afterEach(cleanup)

describe('Time', () => {
  test('renders a semantic time with its machine-readable value', () => {
    const { container } = render(
      <Time value="2026-09-05T21:58:12.345-04:00" format="precise" />
    )
    const time = container.querySelector('time')

    expect(time).not.toBeNull()
    expect(time?.getAttribute('datetime')).toContain('2026-09-06T01:58:12.345')
    expect(time?.textContent).toContain('.345')
  })

  test('accepts Unix seconds', () => {
    const { container } = render(<Time value={1640995200} format="date" />)
    expect(container.querySelector('time')?.getAttribute('datetime')).toContain(
      '2022-01-01'
    )
  })

  test('shows full timestamp context in a tooltip', () => {
    const { container } = render(
      <Time value="2026-09-05T21:58:12-04:00" format="date" />
    )
    fireEvent.mouseEnter(container.querySelector('time')!)
    expect(screen.getByRole('tooltip').textContent).toMatch(/2026/)
  })

  test('shows relative time when the display is already a full timestamp', () => {
    const { container } = render(
      <Time value="2026-09-05T21:58:12-04:00" format="full" />
    )
    fireEvent.mouseEnter(container.querySelector('time')!)
    expect(screen.getByRole('tooltip').textContent).toMatch(/ago|in |just now/i)
  })

  test('anchors the tooltip on the requested side', () => {
    const { container } = render(
      <Time
        value="2026-09-05T21:58:12.345-04:00"
        format="relative"
        tooltipSide="bottom"
      />
    )
    fireEvent.mouseEnter(container.querySelector('time')!)

    expect(screen.getByRole('tooltip').dataset.side).toBe('bottom')
  })

  test('renders fallback for missing and invalid values', () => {
    const { rerender } = render(<Time fallback="Not recorded" />)
    expect(screen.getByText('Not recorded')).toBeTruthy()

    rerender(<Time value="invalid" fallback="Invalid time" />)
    expect(screen.getByText('Invalid time')).toBeTruthy()
  })

  test('uses Text loading at the requested width', () => {
    const { container } = render(<Time loading loadingWidth={12} />)
    const skeleton = container.querySelector('.skeleton-text')
    expect(skeleton).not.toBeNull()
    expect((skeleton as HTMLElement).style.width).toBe('12ch')
  })
})
