import { expect, test } from 'bun:test'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { StepErrorLog } from './StepErrorLog'

const renderErrorLog = (props: Parameters<typeof StepErrorLog>[0]) =>
  render(
    <MemoryRouter>
      <StepErrorLog {...props} />
    </MemoryRouter>
  )

test('renders the default heading and logs link', () => {
  renderErrorLog({ text: 'exit status 1', logsHref: '/logs/1' })

  expect(screen.getByText('Last error log')).toBeDefined()
  expect(screen.getByText('exit status 1')).toBeDefined()
  expect(
    screen.getByRole('link', { name: 'View logs' }).getAttribute('href')
  ).toBe('/logs/1')
})

test('renders a custom heading, body, and link label', () => {
  renderErrorLog({
    text: 'exit status 1',
    logsHref: '/logs/1',
    heading: 'Custom heading',
    linkLabel: 'Custom link',
    children: 'Custom body text',
  })

  expect(screen.getByText('Custom heading')).toBeDefined()
  expect(screen.getByText('Custom body text')).toBeDefined()
  expect(screen.getByText('exit status 1')).toBeDefined()
  expect(
    screen.getByRole('link', { name: 'Custom link' }).getAttribute('href')
  ).toBe('/logs/1')
})
