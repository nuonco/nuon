import { describe, expect, test } from 'bun:test'
import { installCloudLocation, installStatusFacets } from './install-details'

describe('installStatusFacets', () => {
  test('uses sandbox health while an active sandbox is reporting health', () => {
    const facets = installStatusFacets({
      runner_status: 'active',
      sandbox_status: 'active',
      sandbox_health_status: 'unhealthy',
      sandbox_health_message: 'Health check failed',
      composite_component_status: 'deploying',
    })

    expect(facets.map(({ id, status }) => ({ id, status }))).toEqual([
      { id: 'runner', status: 'active' },
      { id: 'sandbox', status: 'unhealthy' },
      { id: 'components', status: 'deploying' },
    ])
    expect(facets[1]?.description).toBe('Health check failed')
  })

  test('uses lifecycle phase when resource statuses are stale', () => {
    const facets = installStatusFacets({
      lifecycle_phase: { phase: 'deprovisioning' },
      runner_status: 'active',
      sandbox_status: 'executing',
      composite_component_status: 'active',
    })

    expect(facets.slice(0, 3).map(({ status }) => status)).toEqual([
      'deprovisioning',
      'deprovisioning',
      'deprovisioning',
    ])
  })

  test('includes aggregate health and drift when reported', () => {
    const facets = installStatusFacets({
      composite_health_status: 'degraded',
      drifted_objects: [{}],
    })

    expect(facets.at(-2)?.id).toBe('health')
    expect(facets.at(-1)).toMatchObject({
      id: 'drift',
      status: 'warn',
      title: 'Drift detected',
    })
  })
})

describe('installCloudLocation', () => {
  test('resolves AWS and GCP regions and Azure locations', () => {
    expect(
      installCloudLocation({
        cloud_platform: 'aws',
        aws_account: { region: 'us-west-2' },
      })
    ).toEqual({
      platform: 'aws',
      region: 'us-west-2',
      location: undefined,
    })
    expect(
      installCloudLocation({
        cloud_platform: 'gcp',
        gcp_account: { region: 'us-central1' },
      })
    ).toEqual({
      platform: 'gcp',
      region: 'us-central1',
      location: undefined,
    })
    expect(
      installCloudLocation({
        cloud_platform: 'azure',
        azure_account: { location: 'eastus' },
      })
    ).toEqual({
      platform: 'azure',
      region: undefined,
      location: 'eastus',
    })
  })
})
