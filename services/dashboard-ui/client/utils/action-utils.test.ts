import { describe, expect, test } from 'vitest'
import { hydrateActionRunSteps } from './action-utils'
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
      expect(result).toEqual(steps)
    })

    test('should hydrate steps with config data', () => {
      const steps = [
        { step_id: 'step-1', status: 'completed' },
        { step_id: 'step-2', status: 'running' },
      ] as TInstallActionRun['steps']

      const stepConfigs = [
        { id: 'step-1', name: 'Build', idx: 0 },
        { id: 'step-2', name: 'Deploy', idx: 1 },
      ] as TActionConfig['steps']

      const result = hydrateActionRunSteps({ steps, stepConfigs })

      expect(result).toEqual([
        { step_id: 'step-1', status: 'completed', name: 'Build', idx: 0 },
        { step_id: 'step-2', status: 'running', name: 'Deploy', idx: 1 },
      ])
    })

    test('should handle steps without matching config', () => {
      const steps = [
        { step_id: 'step-1', status: 'completed' },
        { step_id: 'step-unknown', status: 'running' },
      ] as TInstallActionRun['steps']

      const stepConfigs = [
        { id: 'step-1', name: 'Build', idx: 0 },
      ] as TActionConfig['steps']

      const result = hydrateActionRunSteps({ steps, stepConfigs })

      expect(result).toEqual([
        { step_id: 'step-1', status: 'completed', name: 'Build', idx: 0 },
        { step_id: 'step-unknown', status: 'running' },
      ])
    })
  })
})
