import { describe, expect, test } from 'bun:test'
import {
  mixedHelmPlan,
  redisClusterRollbackPlan,
  vmagentSingleRemovalPlan,
} from '@/lib/fixtures/plan-diffs/helm'
import type { THelmPlan } from '@/types'
import { helmPlanDiff } from './helm'

const occurrences = (text: string, needle: string) =>
  text.split(needle).length - 1

describe('helmPlanDiff', () => {
  test('normalizes direct content and summary counts', () => {
    const result = helmPlanDiff(mixedHelmPlan)

    expect(result.sections.map(({ operation }) => operation)).toEqual([
      'update',
      'create',
      'delete',
    ])
    expect(result.sections.map(({ title }) => title)).toEqual([
      'my-app',
      'my-app-svc',
      'my-app-cache',
    ])
    expect(result.summary).toMatchObject({
      create: 1,
      update: 1,
      delete: 1,
    })
    expect(result.sections[0]?.before).toContain('image: my-app:1.2.0')
    expect(result.sections[0]?.after).toContain('image: my-app:1.3.0')
    expect(result.sections[1]?.before).toBe('')
    expect(result.sections[1]?.after).toContain('kind: Service')
    expect(result.sections[2]?.before).toContain('image: redis:6.2')
    expect(result.sections[2]?.after).toBe('')
  })

  test('builds before and after content from entry arrays', () => {
    const [section] = helmPlanDiff(vmagentSingleRemovalPlan).sections

    expect(section?.operation).toBe('update')
    expect(section?.before).toContain('image: victoriametrics/vmagent:v1.132.0')
    expect(section?.after).toContain('image: victoriametrics/vmagent:v1.132.0')
    expect(
      occurrences(section?.before ?? '', '--remoteWrite.bearerToken=')
    ).toBe(2)
    expect(
      occurrences(section?.after ?? '', '--remoteWrite.bearerToken=')
    ).toBe(1)
    expect(section?.before).toContain('--remoteWrite.tmpDataPath=/tmpData')
    expect(section?.after).toContain('--remoteWrite.tmpDataPath=/tmpData')
  })

  test('preserves rollback direction and operation counts', () => {
    const result = helmPlanDiff(redisClusterRollbackPlan)

    expect(result.description).toBe('Operation: Rollback')
    expect(result.sections.map(({ operation }) => operation)).toEqual([
      'update',
      'delete',
      'update',
    ])
    expect(result.summary).toMatchObject({
      create: 0,
      update: 2,
      delete: 1,
    })
    expect(result.sections[0]?.before).toContain('image: redis:7.2.4-alpine')
    expect(result.sections[0]?.after).toContain('image: redis:7.2.3-alpine')
    expect(result.sections[1]?.before).toContain('kind: Service')
    expect(result.sections[1]?.after).toBe('')
    expect(result.sections[2]?.before).toContain('sentinel monitor mymaster')
    expect(result.sections[2]?.after).not.toContain('sentinel')
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

    expect(section?.title).toBe('vmagent')
    expect(section?.description).toBe('Deployment · apps · observability')
    expect(section?.searchable).toEqual(
      expect.arrayContaining(['observability', 'vmagent', 'Deployment', 'apps'])
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
