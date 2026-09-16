import { describe, expect, test } from 'bun:test'
import { cloudRegionsFor } from './cloud-regions'

describe('cloudRegionsFor', () => {
  test('returns the AWS catalog', () => {
    const regions = cloudRegionsFor('aws')

    expect(regions.some((region) => region.value === 'us-west-2')).toBe(true)
    expect(regions.length).toBeGreaterThan(1)
  })

  test('returns the Azure catalog', () => {
    expect(
      cloudRegionsFor('azure').some((region) => region.value === 'westeurope')
    ).toBe(true)
  })

  test('returns the GCP catalog', () => {
    expect(
      cloudRegionsFor('gcp').some((region) => region.value === 'us-central1')
    ).toBe(true)
  })

  test('returns no regions for an unknown platform', () => {
    expect(cloudRegionsFor('unknown')).toEqual([])
  })
})
