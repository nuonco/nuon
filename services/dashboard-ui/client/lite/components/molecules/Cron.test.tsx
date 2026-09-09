import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { Cron } from './Cron'

afterEach(cleanup)

describe('Cron', () => {
  test('renders a schedule in plain language', () => {
    render(<Cron value="*/15 * * * *" tooltip={false} />)
    expect(screen.getByText(/Every 15 minutes/i)).toBeTruthy()
  })

  test('reveals the expression from human display', () => {
    render(<Cron value="0 */6 * * *" />)
    fireEvent.mouseEnter(screen.getByText(/Every 6 hours/i))
    expect(screen.getByRole('tooltip').textContent).toBe('0 */6 * * *')
  })

  test('reveals the meaning from expression display', () => {
    render(<Cron value="0 */6 * * *" format="expression" />)
    fireEvent.mouseEnter(screen.getByText('0 */6 * * *'))
    expect(screen.getByRole('tooltip').textContent).toMatch(/Every 6 hours/i)
  })

  test('renders both representations without a tooltip dependency', () => {
    const { container } = render(<Cron value="0 9 * * 1-5" format="both" />)
    expect(screen.getByText(/Monday through Friday/i)).toBeTruthy()
    expect(container.querySelector('code')?.textContent).toBe('0 9 * * 1-5')
  })

  test('makes invalid and missing values explicit', () => {
    const { rerender } = render(<Cron value="not a cron" />)
    expect(screen.getByText('Invalid schedule')).toBeTruthy()
    fireEvent.mouseEnter(screen.getByText('Invalid schedule'))
    expect(screen.getByRole('tooltip').textContent).toBe(
      'Invalid schedule: not a cron'
    )

    rerender(<Cron fallback="Not scheduled" />)
    expect(screen.getByText('Not scheduled')).toBeTruthy()
  })

  test('uses Text loading at the requested width', () => {
    const { container } = render(<Cron loading loadingWidth={11} />)
    expect(
      (container.querySelector('.skeleton-text') as HTMLElement).style.width
    ).toBe('11ch')
  })
})
