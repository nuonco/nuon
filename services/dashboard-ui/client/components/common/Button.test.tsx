import { expect, test, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Button } from './Button'

test('calls onClick when clicked', () => {
  const onClick = vi.fn()
  render(<Button onClick={onClick}>Click me</Button>)
  fireEvent.click(screen.getByText(/click me/i))
  expect(onClick).toHaveBeenCalled()
})
