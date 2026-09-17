import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { Status } from './Status'

afterEach(cleanup)

describe('Status', () => {
  test('names a dot in a tooltip', () => {
    render(<Status status="deploying" variant="dot" />)
    fireEvent.mouseEnter(screen.getByLabelText('Deploying'))
    expect(screen.getByRole('tooltip').textContent).toBe('Deploying')
  })

  test('shows a composite status and description in a dot tooltip', () => {
    render(
      <Status
        status={{
          status: 'error',
          status_human_description: 'Terraform apply exited with status 1.',
        }}
        variant="dot"
        label="Apply failed"
      />
    )
    fireEvent.mouseEnter(screen.getByLabelText('Apply failed'))
    const tooltip = screen.getByRole('tooltip')
    expect(tooltip.textContent).toBe(
      'Apply failedTerraform apply exited with status 1.'
    )
    expect(screen.getByText('Apply failed').className).toContain(
      'font-semibold'
    )
  })

  test('wraps very long descriptions without widening the tooltip', () => {
    const longWord = 'AccessDeniedException'.repeat(20)
    render(
      <Status
        status="failed"
        description={`${longWord} ${'Unable to assume the execution role. '.repeat(20)}`}
        variant="dot"
      />
    )
    fireEvent.mouseEnter(screen.getByLabelText('Failed'))
    const tooltip = screen.getByRole('tooltip')
    const description = screen.getByText((content) =>
      content.startsWith(longWord)
    )

    expect(tooltip.className).toContain('max-w-sm')
    expect(description.className).toContain('[overflow-wrap:anywhere]')
    expect(description.className).toContain('line-clamp-10')
  })

  test('pairs the theme icon with the label in a dot tooltip', () => {
    const { container } = render(<Status status="failed" variant="dot" />)
    fireEvent.mouseEnter(screen.getByLabelText('Failed'))
    const tooltip = screen.getByRole('tooltip')

    expect(tooltip.querySelector('svg')).not.toBeNull()
    expect(container.querySelector('svg')).toBeNull()
  })

  test('an in-progress dot tooltip uses a static icon, not the spinner', () => {
    render(<Status status="deploying" variant="dot" />)
    fireEvent.mouseEnter(screen.getByLabelText('Deploying'))
    const tooltip = screen.getByRole('tooltip')

    expect(tooltip.querySelector('svg')).not.toBeNull()
    expect(tooltip.querySelector('.animate-spin')).toBeNull()
  })

  test('the icon variant is a tinted disc that names itself in a tooltip', () => {
    const { container } = render(<Status status="failed" variant="icon" />)
    const disc = container.querySelector('.status-tint')!

    expect(disc.className).toContain('size-6')
    expect(disc.className).toContain('rounded-full')
    expect(disc.querySelector('svg')).not.toBeNull()

    fireEvent.mouseEnter(screen.getByLabelText('Failed'))
    expect(screen.getByRole('tooltip').textContent).toBe('Failed')
  })

  test('an in-progress icon disc spins', () => {
    const { container } = render(<Status status="deploying" variant="icon" />)
    expect(container.querySelector('.status-tint .animate-spin')).not.toBeNull()
  })

  test('uses the status-specific icon independently of its theme', () => {
    const { container, rerender } = render(
      <Status status="cancelled" variant="icon" />
    )
    const cancelledIcon = container.querySelector('.status-tint svg')?.innerHTML

    rerender(<Status status="pending" variant="icon" />)
    const pendingIcon = container.querySelector('.status-tint svg')?.innerHTML

    expect(cancelledIcon).not.toBe(pendingIcon)
  })

  test('allows an icon override without changing the status theme', () => {
    const { container, rerender } = render(
      <Status status="pending" variant="icon" icon="ProhibitIcon" />
    )
    const overriddenIcon =
      container.querySelector('.status-tint svg')?.innerHTML
    const statusStyle = screen.getByLabelText('Pending').getAttribute('style')

    rerender(<Status status="disabled" variant="icon" />)
    const disabledIcon = container.querySelector('.status-tint svg')?.innerHTML

    expect(overriddenIcon).toBe(disabledIcon)
    expect(statusStyle).toContain('--status-color: var(--status-neutral)')
  })

  test('a loading icon variant is a circular skeleton, not a text band', () => {
    const { container } = render(<Status variant="icon" loading />)
    const skeleton = container.firstElementChild!

    expect(skeleton.className).toContain('skeleton')
    expect(skeleton.className).toContain('size-6')
    expect(skeleton.className).toContain('rounded-full')
  })

  test('chip and inline tooltips also title the description', () => {
    const { rerender } = render(
      <Status
        status="degraded"
        description="Runner health checks are delayed."
      />
    )
    fireEvent.mouseEnter(screen.getAllByText('Degraded')[0])
    expect(screen.getByRole('tooltip').textContent).toBe(
      'DegradedRunner health checks are delayed.'
    )

    rerender(
      <Status
        status="degraded"
        description="Runner health checks are delayed."
        variant="inline"
      />
    )
    fireEvent.mouseEnter(screen.getAllByText('Degraded')[0])
    expect(screen.getByRole('tooltip').textContent).toBe(
      'DegradedRunner health checks are delayed.'
    )
  })
})
