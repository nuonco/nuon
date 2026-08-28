import { describe, expect, test } from 'bun:test'
import type { TCustomerManagedBundleArtifact } from '@/types'
import { compareBundleArtifacts } from './BundleComparison'

const artifact = (
  kind: string,
  logicalName: string,
  digest: string
): TCustomerManagedBundleArtifact => ({
  kind,
  logical_name: logicalName,
  digest,
})

describe('compareBundleArtifacts', () => {
  test('classifies bundle inventory changes by kind and logical name', () => {
    const previous = [
      artifact('component', 'api', 'sha256:old'),
      artifact('component', 'worker', 'sha256:same'),
      artifact('action_step', 'retired/run', 'sha256:removed'),
    ]
    const current = [
      artifact('component', 'api', 'sha256:new'),
      artifact('component', 'worker', 'sha256:same'),
      artifact('sandbox', 'sandbox', 'sha256:added'),
    ]

    expect(
      compareBundleArtifacts(previous, current).map(
        ({ kind, name, change }) => ({ kind, name, change })
      )
    ).toEqual([
      { kind: 'component', name: 'api', change: 'changed' },
      { kind: 'action_step', name: 'retired/run', change: 'removed' },
      { kind: 'sandbox', name: 'sandbox', change: 'added' },
      { kind: 'component', name: 'worker', change: 'unchanged' },
    ])
  })
})
