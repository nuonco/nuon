import { describe, expect, test } from 'bun:test'
import type { THelmPlan } from '@/types'
import {
  mixedHelmPlan,
  vmagentSingleRemovalPlan,
} from '../fixtures/plan-diffs/helm'
import { helmPlanDiff } from './helm'

describe('helmPlanDiff', () => {
  test('normalizes direct content and summary counts', () => {
    const result = helmPlanDiff(mixedHelmPlan)

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
    expect(result.sections[0]?.before).toContain(
      'image: example.com/payments:1.2.0'
    )
    expect(result.sections[0]?.after).toContain(
      'image: example.com/payments:1.3.0'
    )
  })

  test('builds before and after content from entry arrays', () => {
    const [section] = helmPlanDiff(vmagentSingleRemovalPlan).sections

    expect(section?.before).toContain('--legacy-endpoint=:8429')
    expect(section?.after).not.toContain('--legacy-endpoint=:8429')
    expect(section?.before).toContain('--logger-format=json')
    expect(section?.after).toContain('--logger-format=json')
  })

  test('preserves modified and error entries from v2 payloads', () => {
    const result = helmPlanDiff({
      op: 'upgrade',
      plan: [
        'default, payments, ConfigMap (v1) to be changed',
        'Plan: 0 to add, 1 to change, 0 to destroy',
      ].join('\n'),
      helm_content_diff: [
        {
          api: 'v1',
          kind: 'ConfigMap',
          name: 'payments',
          namespace: 'default',
          entries: [
            {
              type: 3,
              original: { LOG_LEVEL: 'info' },
              applied: { LOG_LEVEL: 'debug' },
            },
            { type: 4, payload: 'Unable to compare annotations' },
          ],
        },
      ],
    } as unknown as THelmPlan)

    expect(result.sections[0]).toMatchObject({
      before: 'LOG_LEVEL: info',
      after: 'LOG_LEVEL: debug',
      error: 'Unable to compare annotations',
    })
  })

  test('strips ANSI sequences and searches all resource metadata', () => {
    const [section] = helmPlanDiff(vmagentSingleRemovalPlan).sections

    expect(section?.title).toBe('metrics-agent')
    expect(section?.searchable).toEqual(
      expect.arrayContaining([
        'observability',
        'metrics-agent',
        'Deployment',
        'apps/v1',
      ])
    )
  })

  test('names the searchable fields in its placeholder', () => {
    expect(helmPlanDiff(vmagentSingleRemovalPlan).searchPlaceholder).toBe(
      'Search by release, resource, or namespace'
    )
  })

  test('marks missing planner content without inventing a diff', () => {
    const result = helmPlanDiff({
      op: 'install',
      plan: [
        'default, payments, ConfigMap (v1) to be added',
        'Plan: 1 to add, 0 to change, 0 to destroy',
      ].join('\n'),
      helm_content_diff: [],
    } as THelmPlan)

    expect(result.sections[0]).toMatchObject({
      before: '',
      after: '',
      error: 'Diff not available from planner',
    })
  })

  test('falls back to normalized section counts without a summary line', () => {
    const result = helmPlanDiff({
      ...mixedHelmPlan,
      plan: mixedHelmPlan.plan.split('\n').slice(0, -1).join('\n'),
    })

    expect(result.summary).toMatchObject({
      create: 1,
      update: 1,
      delete: 1,
    })
  })
})
