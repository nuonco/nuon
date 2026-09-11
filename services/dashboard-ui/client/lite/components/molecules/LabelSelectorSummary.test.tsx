import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { LabelSelectorSummary } from './LabelSelectorSummary'

afterEach(cleanup)

describe('LabelSelectorSummary', () => {
  test('renders match_labels as key/value chips', () => {
    render(
      <LabelSelectorSummary
        selector={{ match_labels: { env: 'prod', tier: 'a' } }}
      />
    )

    expect(screen.getByText('env')).toBeTruthy()
    expect(screen.getByText('prod')).toBeTruthy()
    expect(screen.getByText('tier')).toBeTruthy()
    expect(screen.getByText('a')).toBeTruthy()
  })

  test('renders not_match_labels with a visible not marker', () => {
    render(
      <LabelSelectorSummary
        selector={{ not_match_labels: { env: 'stage' } }}
      />
    )

    expect(screen.getByLabelText('not env=stage')).toBeTruthy()
    expect(screen.getByText('not')).toBeTruthy()
    expect(screen.getByText('env')).toBeTruthy()
    expect(screen.getByText('stage')).toBeTruthy()
  })

  test('renders a * value as the key plus an any marker', () => {
    render(
      <LabelSelectorSummary selector={{ match_labels: { env: '*' } }} />
    )

    expect(screen.getByText('env')).toBeTruthy()
    expect(screen.getByText('any')).toBeTruthy()
    expect(screen.queryByText('*')).toBeNull()
  })

  test('renders nothing for an empty selector', () => {
    const { container } = render(<LabelSelectorSummary selector={{}} />)
    expect(container.textContent).toBe('')
  })

  test('shows skeleton chips while loading', () => {
    const { container } = render(<LabelSelectorSummary loading />)
    expect(container.querySelectorAll('.skeleton').length).toBe(2)
  })

  test('applies per-key colours to the value half', () => {
    const { container } = render(
      <LabelSelectorSummary
        selector={{ match_labels: { env: 'prod' } }}
        labelColors={{ env: '#4cc9f0' }}
      />
    )

    const value = [...container.querySelectorAll('span')].find((node) =>
      node.className.includes('badge-custom')
    )
    expect(value?.style.getPropertyValue('--badge-color')).toBe('#4cc9f0')
  })
})
