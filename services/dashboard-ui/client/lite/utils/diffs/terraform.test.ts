import { describe, expect, test } from 'bun:test'
import {
  azureNoOpWithCosmeticDriftPlan,
  driftDetectedPlan,
  driftWithChangesAndOutputsPlan,
  eksClusterCreatePlan,
  noOpAndReadResourcesPlan,
  rbacArrayNoisePlan,
  rdsReplacePlan,
  withPlan,
} from '@/lib/fixtures/plan-diffs/terraform'
import type { TTerraformPlan } from '@/types'
import {
  TERRAFORM_DEFAULT_DIFF_OPERATIONS,
  terraformPlanDiff,
} from './terraform'

describe('terraformPlanDiff', () => {
  test('splits create and update resources into one group', () => {
    const { resources, drift, outputs } = terraformPlanDiff(withPlan)

    expect(drift.sections).toEqual([])
    expect(outputs.sections).toEqual([])
    expect(
      resources.sections.map(({ operation, title }) => [operation, title])
    ).toEqual([
      ['create', 'aws_s3_bucket.app_assets'],
      ['update', 'aws_instance.web'],
    ])
    expect(resources.summary).toMatchObject({
      create: 1,
      update: 1,
      delete: 0,
    })
    expect(resources.sections[0]?.after).toContain('bucket = "my-app-assets"')
    expect(resources.sections[1]?.before).toContain(
      'instance_type = "t3.micro"'
    )
    expect(resources.sections[1]?.after).toContain('instance_type = "t3.small"')
  })

  test('collapses create and delete actions into a replace', () => {
    const { resources } = terraformPlanDiff(rdsReplacePlan)

    expect(resources.sections.map(({ operation }) => operation)).toContain(
      'replace'
    )
    expect(resources.summary.replace).toBeGreaterThan(0)
  })

  test('masks values known after apply', () => {
    const { resources } = terraformPlanDiff(eksClusterCreatePlan)
    const cluster = resources.sections.find(
      (section) => section.title === 'aws_eks_cluster.main'
    )

    expect(cluster?.after).toContain('(known after apply)')
    expect(cluster?.after).toContain('version = "1.29"')
  })

  test('omits cosmetic null and empty drift while keeping no-op resources', () => {
    const { drift, resources, outputs } = terraformPlanDiff(
      azureNoOpWithCosmeticDriftPlan
    )

    expect(drift.sections).toEqual([])
    expect(
      resources.sections.every(({ operation }) => operation === 'no-op')
    ).toBe(true)
    expect(
      outputs.sections.every(({ operation }) => operation === 'no-op')
    ).toBe(true)
  })

  test('explains read and no-op resources instead of diffing them', () => {
    const { resources, outputs } = terraformPlanDiff(noOpAndReadResourcesPlan)
    const read = resources.sections.find(
      ({ operation }) => operation === 'read'
    )
    const noOp = resources.sections.find(
      ({ operation }) => operation === 'no-op'
    )

    expect(read).toMatchObject({
      before: '',
      after: '',
      note: 'Terraform will refresh this resource from the provider.',
    })
    expect(noOp).toMatchObject({
      before: '',
      after: '',
      note: 'No changes to this resource.',
    })
    expect(read?.error).toBeUndefined()
    expect(noOp?.error).toBeUndefined()
    expect(outputs.sections[0]?.note).toBe(
      'Terraform will refresh this resource from the provider.'
    )
  })

  test('leaves read and no-op out of the default filter selection', () => {
    expect([...TERRAFORM_DEFAULT_DIFF_OPERATIONS]).toEqual([
      'create',
      'update',
      'replace',
      'delete',
    ])
  })

  test('keeps real drift next to planned resource changes', () => {
    const { drift, resources } = terraformPlanDiff(driftDetectedPlan)

    expect(drift.sections).toHaveLength(1)
    expect(drift.sections[0]?.title).toBe('aws_autoscaling_group.web')
    expect(drift.sections[0]?.before).toContain('desired_capacity = 3')
    expect(drift.sections[0]?.after).toContain('desired_capacity = 5')
    expect(resources.sections.map(({ title }) => title)).toEqual([
      'aws_autoscaling_group.web',
      'aws_launch_template.web',
    ])
  })

  test('serializes outputs including unknown values', () => {
    const { outputs } = terraformPlanDiff(driftWithChangesAndOutputsPlan)

    expect(
      outputs.sections.map(({ title, operation }) => [title, operation])
    ).toEqual([
      ['asg_desired_capacity', 'update'],
      ['launch_template_version', 'update'],
      ['cpu_alarm_arn', 'create'],
    ])
    expect(outputs.sections[2]?.after).toBe('(known after apply)')
    expect(outputs.searchPlaceholder).toBe('Search outputs by name')
  })

  test('still shows array noise as an update', () => {
    const { resources } = terraformPlanDiff(rbacArrayNoisePlan)

    expect(resources.sections).toHaveLength(1)
    expect(resources.sections[0]?.operation).toBe('update')
    expect(resources.sections[0]?.before).toContain('resource_names = null')
    expect(resources.sections[0]?.after).toContain('resource_names = []')
  })

  test('searches address, resource, name, and module', () => {
    const { resources } = terraformPlanDiff(eksClusterCreatePlan)
    const cluster = resources.sections[0]

    expect(cluster?.searchable).toEqual(
      expect.arrayContaining([
        'aws_eks_cluster.main',
        'main',
        'aws_eks_cluster',
        'module.eks',
        'create',
      ])
    )
  })

  test('handles an empty plan', () => {
    const result = terraformPlanDiff({
      resource_changes: [],
    } as TTerraformPlan)

    expect(result.drift.sections).toEqual([])
    expect(result.resources.sections).toEqual([])
    expect(result.outputs.sections).toEqual([])
  })
})
