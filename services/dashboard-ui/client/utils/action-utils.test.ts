import { describe, expect, test } from 'bun:test'
import { hydrateActionRunSteps, sortByIdx } from './action-utils'
import type { TActionConfig, TInstallActionRun } from '@/types'

describe('action-utils', () => {
  describe('hydrateActionRunSteps', () => {
    test('should return empty array when steps is undefined', () => {
      const result = hydrateActionRunSteps({
        steps: undefined as any,
        stepConfigs: [],
      })
      expect(result).toEqual([])
    })

    test('should return original steps when stepConfigs is undefined', () => {
      const steps = [{ step_id: 'test-1' }] as TInstallActionRun['steps']
      const result = hydrateActionRunSteps({
        steps,
        stepConfigs: undefined as any,
      })
      expect(result).toEqual([{ step_id: 'test-1', name: undefined, idx: 0 }])
    })

    test('should hydrate steps with config data', () => {
      const steps = [
        { step_id: 'step-1', status: 'finished' },
        { step_id: 'step-2', status: 'pending' },
      ] as TInstallActionRun['steps']

      const stepConfigs = [
        { id: 'step-1', name: 'Build', idx: 0 },
        { id: 'step-2', name: 'Deploy', idx: 1 },
      ] as TActionConfig['steps']

      const result = hydrateActionRunSteps({ steps, stepConfigs })

      expect(result).toEqual([
        { step_id: 'step-1', status: 'finished', name: 'Build', idx: 0 },
        { step_id: 'step-2', status: 'pending', name: 'Deploy', idx: 1 },
      ])
    })

    test('should handle steps without matching config', () => {
      const steps = [
        { step_id: 'step-1', status: 'finished' },
        { step_id: 'step-unknown', status: 'pending' },
      ] as TInstallActionRun['steps']

      const stepConfigs = [
        { id: 'step-1', name: 'Build', idx: 0 },
      ] as TActionConfig['steps']

      const result = hydrateActionRunSteps({ steps, stepConfigs })

      expect(result).toEqual([
        { step_id: 'step-1', status: 'finished', name: 'Build', idx: 0 },
        { step_id: 'step-unknown', status: 'pending', name: undefined, idx: 1 },
      ])
    })
  })

  describe('sortByIdx', () => {
    test('sorts by idx ascending', () => {
      const items = [{ idx: 3 }, { idx: 1 }, { idx: 2 }]
      expect(sortByIdx(items)).toEqual([{ idx: 1 }, { idx: 2 }, { idx: 3 }])
    })

    test('handles empty array', () => {
      expect(sortByIdx([])).toEqual([])
    })

    test('handles all undefined idx values', () => {
      const items: { name: string; idx?: number }[] = [{ name: 'a' }, { name: 'b' }, { name: 'c' }]
      expect(sortByIdx(items)).toEqual([
        { name: 'a' },
        { name: 'b' },
        { name: 'c' },
      ])
    })

    test('places undefined idx items before defined ones', () => {
      const items: { idx?: number; name?: string }[] = [{ idx: 2 }, { name: 'no-idx' }, { idx: 1 }]
      expect(sortByIdx(items)).toEqual([
        { name: 'no-idx' },
        { idx: 1 },
        { idx: 2 },
      ])
    })

    test('handles mix of undefined and defined idx', () => {
      const items = [
        { idx: 3, name: 'c' },
        { name: 'x' },
        { idx: 1, name: 'a' },
        { name: 'y' },
        { idx: 2, name: 'b' },
      ]
      expect(sortByIdx(items)).toEqual([
        { name: 'x' },
        { name: 'y' },
        { idx: 1, name: 'a' },
        { idx: 2, name: 'b' },
        { idx: 3, name: 'c' },
      ])
    })

    test('does not mutate the original array', () => {
      const items = [{ idx: 2 }, { idx: 1 }]
      const original = [...items]
      sortByIdx(items)
      expect(items).toEqual(original)
    })

    test('handles single item', () => {
      const items = [{ idx: 5 }]
      expect(sortByIdx(items)).toEqual([{ idx: 5 }])
    })
  })
})
