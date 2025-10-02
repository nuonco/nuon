import { describe, expect, test } from 'vitest'
import {
  getWorkflowBadge,
  getStepBadge,
  getStepButtons,
  getStepBanner,
} from './workflow-utils'
import type { TWorkflow, TWorkflowStep } from '@/types'

describe('workflow-utils', () => {
  describe('getWorkflowBadge', () => {
    test('should return correct badge for success workflow', () => {
      const workflow: TWorkflow = {
        status: {
          status: 'success',
        },
      } as TWorkflow

      const badge = getWorkflowBadge(workflow)
      expect(badge).toEqual({
        children: 'Completed',
        theme: 'success',
      })
    })

    test('should return correct badge for error workflow', () => {
      const workflow: TWorkflow = {
        status: {
          status: 'error',
        },
      } as TWorkflow

      const badge = getWorkflowBadge(workflow)
      expect(badge).toEqual({
        children: 'Failed',
        theme: 'error',
      })
    })

    test('should return correct badge for approval-awaiting workflow', () => {
      const workflow: TWorkflow = {
        status: {
          status: 'approval-awaiting',
        },
      } as TWorkflow

      const badge = getWorkflowBadge(workflow)
      expect(badge).toEqual({
        children: 'Awaiting approval',
        theme: 'warn',
      })
    })

    test('should return empty object for unknown workflow status', () => {
      const workflow: TWorkflow = {
        status: {
          status: 'unknown_status' as any,
        },
      } as TWorkflow

      const badge = getWorkflowBadge(workflow)
      expect(badge).toEqual({})
    })

    test('should return empty object for workflow without status', () => {
      const workflow: TWorkflow = {} as TWorkflow

      const badge = getWorkflowBadge(workflow)
      expect(badge).toEqual({})
    })
  })

  describe('getStepBadge', () => {
    test('should return retried badge when step is retried', () => {
      const step: TWorkflowStep = {
        retried: true,
      } as TWorkflowStep

      const badge = getStepBadge(step)
      expect(badge).toEqual({
        children: 'Retried',
        theme: 'info',
      })
    })

    test('should return skipped badge when execution_type is skipped', () => {
      const step: TWorkflowStep = {
        execution_type: 'skipped',
        retried: false,
      } as TWorkflowStep

      const badge = getStepBadge(step)
      expect(badge).toEqual({
        children: 'Skipped',
      })
    })

    test('should return correct badge for approved step', () => {
      const step: TWorkflowStep = {
        status: {
          status: 'approved',
        },
        retried: false,
      } as TWorkflowStep

      const badge = getStepBadge(step)
      expect(badge).toEqual({
        children: 'Plan approved',
        theme: 'success',
      })
    })

    test('should return correct badge for approval-denied step', () => {
      const step: TWorkflowStep = {
        status: {
          status: 'approval-denied',
        },
        retried: false,
      } as TWorkflowStep

      const badge = getStepBadge(step)
      expect(badge).toEqual({
        children: 'Plan denied',
        theme: 'warn',
      })
    })

    test('should return empty object for step without status', () => {
      const step: TWorkflowStep = {
        retried: false,
      } as TWorkflowStep

      const badge = getStepBadge(step)
      expect(badge).toEqual({})
    })
  })

  describe('getStepButtons', () => {
    test('should show approval and cancel buttons for approval-awaiting step', () => {
      const step: TWorkflowStep = {
        status: {
          status: 'approval-awaiting',
        },
      } as TWorkflowStep

      const buttons = getStepButtons(step)
      expect(buttons).toEqual({
        approval: true,
        cancel: true,
        retry: false,
      })
    })

    test('should show retry button for error step with retryable flag', () => {
      const step: TWorkflowStep = {
        status: {
          status: 'error',
        },
        retryable: true,
        retried: false,
      } as TWorkflowStep

      const buttons = getStepButtons(step)
      expect(buttons).toEqual({
        approval: false,
        cancel: false,
        retry: true,
      })
    })

    test('should not show retry button for error step without retryable flag', () => {
      const step: TWorkflowStep = {
        status: {
          status: 'error',
        },
        retryable: false,
      } as TWorkflowStep

      const buttons = getStepButtons(step)
      expect(buttons).toEqual({
        approval: false,
        cancel: false,
        retry: false,
      })
    })

    test('should not show retry button for error step that was already retried', () => {
      const step: TWorkflowStep = {
        status: {
          status: 'error',
        },
        retryable: true,
        retried: true,
      } as TWorkflowStep

      const buttons = getStepButtons(step)
      expect(buttons).toEqual({
        approval: false,
        cancel: false,
        retry: false,
      })
    })

    test('should show cancel button for in-progress step', () => {
      const step: TWorkflowStep = {
        status: {
          status: 'in-progress',
        },
      } as TWorkflowStep

      const buttons = getStepButtons(step)
      expect(buttons).toEqual({
        approval: false,
        cancel: true,
        retry: false,
      })
    })

    test('should not show any buttons for success step', () => {
      const step: TWorkflowStep = {
        status: {
          status: 'success',
        },
      } as TWorkflowStep

      const buttons = getStepButtons(step)
      expect(buttons).toEqual({
        approval: false,
        cancel: false,
        retry: false,
      })
    })

    test('should handle step without status', () => {
      const step: TWorkflowStep = {} as TWorkflowStep

      const buttons = getStepButtons(step)
      expect(buttons).toEqual({
        approval: false,
        cancel: false,
        retry: false,
      })
    })
  })

  describe('getStepBanner', () => {
    test('should return error banner for error step', () => {
      const step: TWorkflowStep = {
        status: {
          status: 'error',
          status_human_description: 'Connection failed',
        },
      } as TWorkflowStep

      const banner = getStepBanner(step)
      expect(banner).toEqual({
        copy: 'Step encountered an error: Connection failed',
        theme: 'error',
        title: 'Step failed',
      })
    })

    test('should return cancelled banner for cancelled step', () => {
      const step: TWorkflowStep = {
        status: {
          status: 'cancelled',
          status_human_description: 'User cancelled',
        },
      } as TWorkflowStep

      const banner = getStepBanner(step)
      expect(banner).toEqual({
        copy: 'Step was cancelled: User cancelled',
        theme: 'warn',
        title: 'Step cancelled',
      })
    })

    test('should return discarded banner for discarded step', () => {
      const step: TWorkflowStep = {
        status: {
          status: 'discarded',
          status_human_description: 'Step discarded',
        },
      } as TWorkflowStep

      const banner = getStepBanner(step)
      expect(banner).toEqual({
        copy: 'Step was discarded: Step discarded',
        theme: 'default',
        title: 'Step discarded',
      })
    })

    test('should return user-skipped banner with email', () => {
      const step: TWorkflowStep = {
        status: {
          status: 'user-skipped',
          status_human_description: 'Manually skipped',
        },
        created_by: {
          email: 'user@example.com',
        },
      } as TWorkflowStep

      const banner = getStepBanner(step)
      expect(banner).toEqual({
        copy: 'Step was skipped by user@example.com: Manually skipped',
        theme: 'default',
        title: 'Step skipped',
      })
    })

    test('should return skipped banner for skipped execution type', () => {
      const step: TWorkflowStep = {
        status: {
          status: 'success',
        },
        execution_type: 'skipped',
      } as TWorkflowStep

      const banner = getStepBanner(step)
      expect(banner).toEqual({
        copy: 'Step was skipped due to being a plan only workflow',
        theme: 'default',
        title: 'Step skipped',
      })
    })

    test('should return retried banner for retried step', () => {
      const step: TWorkflowStep = {
        status: {
          status: 'success',
          status_human_description: 'Completed after retry',
        },
        retryable: true,
        retried: true,
        created_by: {
          email: 'user@example.com',
        },
      } as TWorkflowStep

      const banner = getStepBanner(step)
      expect(banner).toEqual({
        copy: 'Step was retried by user@example.com: Completed after retry',
        theme: 'info',
        title: 'Step retried',
      })
    })

    test('should return undefined for normal step states', () => {
      const step: TWorkflowStep = {
        status: {
          status: 'success',
        },
      } as TWorkflowStep

      const banner = getStepBanner(step)
      expect(banner).toBeUndefined()
    })

    test('should return undefined for step without status', () => {
      const step: TWorkflowStep = {} as TWorkflowStep

      const banner = getStepBanner(step)
      expect(banner).toBeUndefined()
    })
  })
})
