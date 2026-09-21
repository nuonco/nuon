import { describe, expect, test } from 'bun:test'
import {
  appConfigAllSections,
  appConfigMockSections,
} from '@/lib/fixtures/plan-diffs/app-config'
import type { TAppConfigDiffSection } from '@/types'
import { appConfigPlanDiff } from './app-config'

describe('appConfigPlanDiff', () => {
  test('preserves the supplied legacy summary', () => {
    const result = appConfigPlanDiff(appConfigMockSections, {
      added: 2,
      removed: 1,
      changed: 3,
    })

    expect(result.summary).toEqual({
      create: 2,
      update: 3,
      replace: 0,
      delete: 1,
      read: 0,
      'no-op': 0,
    })
  })

  test('maps grouped entities and their fields into TOML diffs', () => {
    const result = appConfigPlanDiff(appConfigMockSections)
    const redis = result.sections.find(({ title }) => title === 'redis')

    expect(redis).toMatchObject({
      operation: 'create',
      description: 'Components · helm_chart',
      filename: 'redis.toml',
      language: 'toml',
    })
    expect(redis?.after).toContain("type = 'helm_chart'")
    expect(redis?.before).toBe('')
  })

  test('keeps embedded files as independently searchable sections', () => {
    const result = appConfigPlanDiff(appConfigMockSections)
    const values = result.sections.find(
      ({ title }) => title === './values/prod.yaml'
    )
    const script = result.sections.find(
      ({ title }) => title === 'inline_contents'
    )

    expect(values).toMatchObject({
      description: 'Components · ctl-api',
      operation: 'update',
      language: 'yaml',
      filename: './values/prod.yaml',
    })
    expect(values?.before).toContain('replicas: 1')
    expect(values?.after).toContain('replicas: 3')
    expect(script).toMatchObject({
      description: 'Actions · healthcheck',
      language: 'shellscript',
    })
  })

  test('uses complete section content instead of duplicating child entities', () => {
    const result = appConfigPlanDiff(appConfigAllSections)
    const inputs = result.sections.filter(({ searchable }) =>
      searchable.includes('inputs')
    )
    const secrets = result.sections.filter(({ searchable }) =>
      searchable.includes('secrets')
    )

    expect(inputs).toHaveLength(1)
    expect(inputs[0]).toMatchObject({
      title: 'Install inputs',
      operation: 'create',
      language: 'toml',
      filename: 'inputs.toml',
    })
    expect(inputs[0]?.after).toContain('[[input]]')
    expect(secrets).toHaveLength(1)
    expect(secrets[0]?.after).toContain('[[secret]]')
  })

  test('preserves section-level runner, stack, sandbox, and permission TOML', () => {
    const result = appConfigPlanDiff(appConfigAllSections)

    expect(
      result.sections.find(({ title }) => title === 'Runner')?.after
    ).toContain('runner_type = "gpu"')
    expect(
      result.sections.find(({ title }) => title === 'Stack')?.after
    ).toContain('type = "eks-v2"')
    expect(
      result.sections.find(({ title }) => title === 'Sandbox')?.after
    ).toContain('terraform_version = "1.6.0"')
    expect(
      result.sections.find(({ title }) => title === 'Permissions')?.after
    ).toContain('[provision_role]')
  })

  test('marks contentless fields as unavailable', () => {
    const source: TAppConfigDiffSection[] = [
      {
        name: 'Runner',
        sectionKey: 'runner',
        additions: 0,
        removals: 0,
        changed: 1,
        grouped: false,
        entities: [],
        fields: [{ key: 'type', op: 'change', diff: '' }],
      },
    ]

    expect(appConfigPlanDiff(source).sections[0]?.error).toBe(
      'Diff not available from planner'
    )
  })

  test('handles an empty config diff', () => {
    const result = appConfigPlanDiff([])

    expect(result.sections).toEqual([])
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
