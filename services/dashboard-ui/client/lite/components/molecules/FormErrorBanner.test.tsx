import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { FormErrorBanner } from './FormErrorBanner'

afterEach(cleanup)

test('renders nothing without an error', () => {
  const { container } = render(
    <FormErrorBanner error={null} fallback="Unable to save changes" />
  )

  expect(container.innerHTML).toBe('')
})

test('prefers the API error over the fallback', () => {
  render(
    <FormErrorBanner
      error={{
        error: 'Configuration update failed',
        description: '',
        user_error: true,
      }}
      fallback="Unable to save changes"
    />
  )

  expect(screen.getByText('Configuration update failed')).toBeTruthy()
  expect(screen.queryByText('Unable to save changes')).toBeNull()
})

test('falls back to a thrown error message', () => {
  render(
    <FormErrorBanner
      error={new Error('Network connection lost')}
      fallback="Unable to save changes"
    />
  )

  expect(screen.getByText('Network connection lost')).toBeTruthy()
})

test('uses the fallback when the error carries no message', () => {
  render(
    <FormErrorBanner error={new Error()} fallback="Unable to save changes" />
  )

  expect(screen.getByText('Unable to save changes')).toBeTruthy()
})

test('renders the API description alongside the heading', () => {
  render(
    <FormErrorBanner
      error={{
        error: 'Configuration update failed',
        description: 'The configuration changed after this form was opened.',
        user_error: true,
      }}
      fallback="Unable to save changes"
    />
  )

  expect(
    screen.getByText('The configuration changed after this form was opened.')
  ).toBeTruthy()
})

test('announces assertively through the error theme', () => {
  render(
    <FormErrorBanner
      error={new Error('Network connection lost')}
      fallback="Unable to save changes"
    />
  )

  const alert = screen.getByRole('alert')

  expect(alert.getAttribute('aria-live')).toBe('assertive')
  expect(alert.getAttribute('data-banner-theme')).toBe('error')
})
