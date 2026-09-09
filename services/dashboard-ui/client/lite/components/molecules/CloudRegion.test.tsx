import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { CloudRegion } from './CloudRegion'

afterEach(cleanup)

describe('CloudRegion', () => {
  test('resolves an AWS region and country flag', () => {
    render(<CloudRegion platform="aws" region="us-west-2" />)

    expect(screen.getByText('US West (Oregon)')).not.toBeNull()
    expect(screen.getByText('🇺🇸')).toHaveAttribute('aria-hidden', 'true')
  })

  test('uses location for Azure', () => {
    render(<CloudRegion platform="azure" location="westeurope" />)

    expect(screen.getByText('West Europe')).not.toBeNull()
    expect(screen.getByText('🇳🇱')).not.toBeNull()
  })

  test('uses region for GCP', () => {
    render(<CloudRegion platform="gcp" region="us-central1" />)

    expect(screen.getByText('Iowa (US)')).not.toBeNull()
  })

  test('renders Unknown for missing and unrecognized regions', () => {
    const { rerender } = render(<CloudRegion platform="unknown" />)
    expect(screen.getByText('Unknown')).not.toBeNull()

    rerender(<CloudRegion platform="aws" region="invalid-region" />)
    expect(screen.getByText('Unknown')).not.toBeNull()
  })

  test('inherits the Text loading state', () => {
    const { container } = render(
      <CloudRegion platform="aws" region="us-west-2" loading />
    )

    expect(container.querySelector('.skeleton-text')).not.toBeNull()
    expect(screen.queryByText('US West (Oregon)')).toBeNull()
  })
})
