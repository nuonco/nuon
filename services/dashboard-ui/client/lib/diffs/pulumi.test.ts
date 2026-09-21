import { describe, expect, test } from 'bun:test'
import {
  azureCosmeticUpdatesPlan,
  databaseReplacePlan,
  ecsServiceUpdatePlan,
  mixedInfraChangesPlan,
  s3BucketCreatePlan,
  withDiagnosticsPlan,
} from '@/lib/fixtures/plan-diffs/pulumi'
import { PULUMI_DEFAULT_DIFF_OPERATIONS, pulumiPlanDiff } from './pulumi'

describe('pulumiPlanDiff', () => {
  test('normalizes creates and serializes inputs deterministically', () => {
    const result = pulumiPlanDiff(s3BucketCreatePlan)

    expect(result.sections).toHaveLength(3)
    expect(
      result.sections.every(({ operation }) => operation === 'create')
    ).toBe(true)
    expect(result.summary.create).toBe(3)
    expect(result.sections[0]?.after).toContain(
      '"bucket": "acme-artifacts-prod"'
    )
    expect(result.sections[0]?.after.indexOf('"bucket"')).toBeLessThan(
      result.sections[0]?.after.indexOf('"forceDestroy"') ?? 0
    )
  })

  test('preserves old and new inputs for updates', () => {
    const result = pulumiPlanDiff(ecsServiceUpdatePlan)
    const service = result.sections.find(
      ({ title }) => title === 'aws:ecs/service:Service'
    )

    expect(service).toMatchObject({
      operation: 'update',
      description: 'api-service · desiredCount',
    })
    expect(service?.before).toContain('"desiredCount": 2')
    expect(service?.after).toContain('"desiredCount": 4')
  })

  test('normalizes replacement lifecycle actions', () => {
    const result = pulumiPlanDiff(databaseReplacePlan)

    expect(result.sections.map(({ operation }) => operation)).toEqual([
      'replace',
      'replace',
      'replace',
    ])
    expect(result.summary.replace).toBe(3)
  })

  test('describes read and unchanged resources without a diff body', () => {
    const result = pulumiPlanDiff(mixedInfraChangesPlan)
    const read = result.sections.find(({ operation }) => operation === 'read')
    const noOp = result.sections.find(({ operation }) => operation === 'no-op')

    expect(read).toMatchObject({
      before: '',
      after: '',
      note: 'Pulumi will read this resource from the provider.',
    })
    expect(noOp).toMatchObject({
      before: '',
      after: '',
      note: 'No changes to this resource.',
    })
  })

  test('uses the same default operations as the legacy filter', () => {
    expect([...PULUMI_DEFAULT_DIFF_OPERATIONS]).toEqual([
      'create',
      'update',
      'replace',
      'delete',
    ])
  })

  test('keeps cosmetic updates visible', () => {
    const result = pulumiPlanDiff(azureCosmeticUpdatesPlan)

    expect(
      result.sections.filter(({ operation }) => operation === 'update')
    ).toHaveLength(3)
    expect(result.sections[0]?.before).toContain('"tags": null')
    expect(result.sections[0]?.after).toContain('"tags": {}')
  })

  test('preserves diagnostics outside resource sections', () => {
    const result = pulumiPlanDiff(withDiagnosticsPlan)

    expect(result.diagnostics).toHaveLength(2)
    expect(
      result.diagnostics?.every(({ severity }) => severity === 'warning')
    ).toBe(true)
  })

  test('uses detailed diffs when inputs are unavailable', () => {
    const result = pulumiPlanDiff(withDiagnosticsPlan)
    const api = result.sections[0]

    expect(api?.after).toContain('"description"')
    expect(api?.after).toContain('"inputDiff": true')
  })

  test('handles an empty plan', () => {
    const result = pulumiPlanDiff()

    expect(result.sections).toEqual([])
    expect(result.diagnostics).toBeUndefined()
    expect(result.summary).toEqual({
      create: 0,
      update: 0,
      replace: 0,
      delete: 0,
      read: 0,
      'no-op': 0,
    })
  })
})
