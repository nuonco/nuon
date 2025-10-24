import { expect, test } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Link } from './Link'

test('renders a link', () => {
  render(<Link href="/test">Test</Link>)
  const link = screen.getByRole('link')

  expect(link).toBeInTheDocument()
  expect(link).toHaveAttribute('href', '/test')
  expect(link).toHaveTextContent('Test')
})
