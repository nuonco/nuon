import { expect, test, mock } from 'bun:test'
import { render, screen, fireEvent } from '@testing-library/react'
import { Button } from './Button'

test('calls onClick when clicked', () => {
  const onClick = mock()
  render(<Button onClick={onClick}>Click me</Button>)
  fireEvent.click(screen.getByText(/click me/i))
  expect(onClick).toHaveBeenCalled()
})

test('tab labels are neutral and the underline stays primary', () => {
  const { rerender } = render(<Button variant="tab">Logs</Button>)
  const inactive = screen.getByRole('button', { name: 'Logs' }).className.split(/\s+/)
  expect(inactive).toContain('text-cool-grey-600')
  expect(inactive).toContain('dark:text-white/70')
  expect(inactive).toContain('hover:text-foreground')
  expect(inactive).toContain('hover:!border-primary-600/50')
  expect(inactive).not.toContain('text-primary-600')
  expect(inactive).not.toContain('dark:text-primary-400')

  rerender(
    <Button variant="tab" isActive>
      Logs
    </Button>
  )
  const active = screen.getByRole('button', { name: 'Logs' }).className.split(/\s+/)
  expect(active).toContain('!text-foreground')
  expect(active).toContain('!border-primary-600')
  expect(active).not.toContain('text-primary-600')
  expect(active).not.toContain('dark:text-primary-400')
})
