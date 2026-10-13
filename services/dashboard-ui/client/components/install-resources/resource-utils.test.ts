import { describe, expect, test } from 'bun:test'
import type { TInstallResource } from '@/types'
import {
  isHealthCheckResource,
  resourceIdentity,
  resourceOwner,
  resourceSummary,
} from './resource-utils'

describe('health resource identity and classification', () => {
  test('summarizes reported readiness, not current replica counts or an invented desired count', () => {
    expect(
      resourceSummary({
        kind: 'Deployment',
        details: JSON.stringify({
          spec: { replicas: 5 },
          status: { replicas: 4, readyReplicas: 0 },
        }),
      })
    ).toBe('0 / 5 ready')
    expect(
      resourceSummary({
        kind: 'Deployment',
        details: JSON.stringify({ status: { readyReplicas: 2 } }),
      })
    ).toBe('2 ready')
    expect(resourceSummary({ kind: 'Deployment' })).toBeUndefined()
    expect(
      resourceSummary({
        kind: 'Pod',
        details: JSON.stringify({
          status: {
            phase: 'Running',
            containerStatuses: [
              {
                name: 'app',
                ready: false,
                restartCount: 17,
                state: { waiting: { reason: 'CrashLoopBackOff' } },
              },
            ],
          },
        }),
      })
    ).toBe('0 / 1 ready · CrashLoopBackOff · 17 restarts')
    expect(
      resourceSummary({
        kind: 'Pod',
        details: JSON.stringify({
          status: {
            phase: 'Running',
            containerStatuses: [{ name: 'app', state: { running: {} } }],
          },
        }),
      })
    ).toBe('Running')
  })

  test('classifies checks by provider, not names that custom Kubernetes kinds can share', () => {
    expect(
      isHealthCheckResource({ provider: 'probe', kind: 'ExecProbe' })
    ).toBe(true)
    expect(
      isHealthCheckResource({ provider: 'custom', kind: 'CustomCheck' })
    ).toBe(true)
    expect(
      isHealthCheckResource({ provider: 'kubernetes', kind: 'CustomCheck' })
    ).toBe(false)
    expect(
      isHealthCheckResource({ provider: 'kubernetes', kind: 'HTTPProbe' })
    ).toBe(false)
    expect(
      isHealthCheckResource({ provider: 'aws', kind: 'aws_db_instance' })
    ).toBe(false)
  })

  test('keeps overlapping resources distinct by install, owner, provider, API group and namespace', () => {
    const resource: TInstallResource = {
      install_id: 'inst-1',
      source: 'component',
      install_component_id: 'instcmp-api',
      provider: 'kubernetes',
      api_group: 'databases.example.com',
      kind: 'Database',
      namespace: 'acme',
      name: 'primary',
    }
    const variants: TInstallResource[] = [
      resource,
      { ...resource, install_id: 'inst-2' },
      { ...resource, install_component_id: 'instcmp-worker' },
      { ...resource, source: 'sandbox', owner_name: 'api' },
      { ...resource, provider: 'custom' },
      { ...resource, api_group: 'storage.example.com' },
      { ...resource, namespace: 'jobs' },
    ]
    expect(new Set(variants.map(resourceIdentity)).size).toBe(7)
    expect(
      resourceIdentity({
        ...resource,
        health: 'unhealthy',
        observed_at: '2026-10-08T09:30:00Z',
      })
    ).toBe(resourceIdentity(resource))
    expect(resourceIdentity({ kind: 'Pod', name: 'api' })).toBe(
      resourceIdentity({
        kind: 'Pod',
        name: 'api',
        namespace: '',
        api_group: '',
      })
    )
  })

  test('falls back to component IDs without merging sandbox releases or colliding with named components', () => {
    expect(resourceOwner({ install_component_id: 'instcmp-api' }, {})).toEqual({
      key: 'component:instcmp-api',
      label: 'instcmp-api',
    })
    expect(
      resourceOwner(
        { install_component_id: 'instcmp-api' },
        { 'instcmp-api': 'api' }
      ).label
    ).toBe('api')
    expect(resourceOwner({ source: 'sandbox', owner_name: 'api' }, {})).toEqual(
      { key: 'sandbox:api', label: 'Sandbox · api' }
    )
    expect(
      resourceOwner({ source: 'sandbox', owner_name: 'api' }, {}).key
    ).not.toBe(resourceOwner({ install_component_id: 'api' }, {}).key)
  })
})
