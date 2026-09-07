import { describe, expect, test } from 'bun:test'
import {
  deploymentUpgradePlan,
  freshInstallPlan,
  largeScatteredChangesPlan,
  mixedOperationsPlan,
  mockPlan,
  withErrorsPlan,
} from '@/lib/fixtures/plan-diffs/kubernetes'
import type { TKubernetesPlan } from '@/types'
import { kubernetesPlanDiff } from './kubernetes'

describe('kubernetesPlanDiff', () => {
  test('normalizes path-based original and applied values', () => {
    const result = kubernetesPlanDiff(mockPlan)

    expect(result.sections.map(({ operation }) => operation)).toEqual([
      'update',
      'create',
    ])
    expect(result.sections[0]).toMatchObject({
      title: 'my-configmap',
      description: 'ConfigMap · v1 · default',
      before: 'data.DATABASE_URL: postgres://old-host:5432/db',
      after: 'data.DATABASE_URL: postgres://new-host:5432/db',
      filename: 'my-configmap.yaml',
    })
    expect(result.summary).toMatchObject({
      create: 1,
      update: 1,
      delete: 0,
    })
  })

  test('preserves unchanged lines and changed values in entry arrays', () => {
    const result = kubernetesPlanDiff(deploymentUpgradePlan)
    const deployment = result.sections[0]

    expect(deployment?.before).toContain('apiVersion: apps/v1')
    expect(deployment?.after).toContain('apiVersion: apps/v1')
    expect(deployment?.before).toContain('replicas: 2')
    expect(deployment?.after).toContain('replicas: 3')
    expect(deployment?.before).toContain('api-server:v2.14.0')
    expect(deployment?.after).toContain('api-server:v2.15.1')
    expect(result.summary.update).toBe(3)
  })

  test('normalizes create, update, and delete operations', () => {
    const result = kubernetesPlanDiff(mixedOperationsPlan)

    expect(result.sections.map(({ operation }) => operation)).toEqual([
      'update',
      'create',
      'delete',
    ])
    expect(result.summary).toMatchObject({
      create: 1,
      update: 1,
      delete: 1,
    })
  })

  test('preserves complete create manifests', () => {
    const result = kubernetesPlanDiff(freshInstallPlan)

    expect(result.sections).toHaveLength(3)
    expect(result.sections.every(({ before }) => before === '')).toBe(true)
    expect(result.sections[0]?.after).toContain('kind: Deployment')
    expect(result.sections[1]?.after).toContain('kind: ServiceAccount')
    expect(result.sections[2]?.after).toContain('kind: Service')
    expect(result.summary.create).toBe(3)
  })

  test('keeps planner errors as sections without counting them as changes', () => {
    const result = kubernetesPlanDiff(withErrorsPlan)

    expect(result.sections).toHaveLength(3)
    expect(result.sections[1]).toMatchObject({
      title: 'broken-crd',
      error:
        'resource "gateway.networking.k8s.io/v1/GatewayRoute" not found: ensure the CRD is installed',
      before: '',
      after: '',
    })
    expect(result.sections[2]?.error).toContain('RATE_LIMIT')
    expect(result.summary.update).toBe(1)
  })

  test('marks resources the planner returned no content for', () => {
    const result = kubernetesPlanDiff(mockPlan)

    expect(result.sections[1]).toMatchObject({
      title: 'new-secret',
      before: '',
      after: '',
      error: 'Diff not available from planner',
    })
    expect(kubernetesPlanDiff(deploymentUpgradePlan).sections[2]).toMatchObject(
      {
        before: '',
        after: '',
        error: 'Diff not available from planner',
      }
    )
  })

  test('preserves entry-level errors beside available content', () => {
    const result = kubernetesPlanDiff({
      plan: 'k8s-entry-error',
      op: 'apply',
      k8s_content_diff: [
        {
          _version: 'v1',
          name: 'api',
          namespace: 'default',
          kind: 'Deployment',
          api: 'apps/v1',
          resource: 'deployments',
          op: 'apply',
          type: 3,
          dry_run: false,
          entries: [
            { type: 0, payload: 'kind: Deployment', path: '' },
            { type: 4, payload: 'Unable to compare labels', path: '' },
          ],
        },
      ],
    } as TKubernetesPlan)

    expect(result.sections[0]).toMatchObject({
      before: 'kind: Deployment',
      after: 'kind: Deployment',
      error: 'Unable to compare labels',
    })
    expect(result.summary.update).toBe(1)
  })

  test('searches every resource identity field', () => {
    const [section] = kubernetesPlanDiff(largeScatteredChangesPlan).sections

    expect(section?.searchable).toEqual(
      expect.arrayContaining([
        'ctl-api-auth',
        'apps',
        'Deployment',
        'apps/v1',
        'deployments',
        'changed',
      ])
    )
    expect(section?.before).toContain('replicas: 2')
    expect(section?.after).toContain('replicas: 4')
  })

  test('names the searchable fields in its placeholder', () => {
    expect(kubernetesPlanDiff(mockPlan).searchPlaceholder).toBe(
      'Search by name, resource, type, or namespace'
    )
  })

  test('handles an empty plan', () => {
    const result = kubernetesPlanDiff({
      plan: '',
      op: 'apply',
      k8s_content_diff: [],
    })

    expect(result.sections).toEqual([])
    expect(result.summary).toMatchObject({
      create: 0,
      update: 0,
      delete: 0,
    })
  })
})
