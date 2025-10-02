import { describe, expect, test } from 'vitest'
import { getJobHref, getJobName, getJobExecutionStatus } from './runner-utils'
import type { TRunnerJob } from '@/types'

describe('runner-utils', () => {
  describe('getJobHref', () => {
    test('should generate correct href for build jobs', () => {
      const mockJob: TRunnerJob = {
        group: 'build',
        metadata: {
          app_id: 'app123',
          component_id: 'comp456',
          component_build_id: 'build789',
        },
      } as TRunnerJob

      const href = getJobHref(mockJob)
      expect(href).toBe('apps/app123/components/comp456/builds/build789')
    })

    test('should generate correct href for sandbox jobs', () => {
      const mockJob: TRunnerJob = {
        group: 'sandbox',
        metadata: {
          install_id: 'install123',
          sandbox_run_id: 'sandbox456',
        },
      } as TRunnerJob

      const href = getJobHref(mockJob)
      expect(href).toBe('installs/install123/sandbox/sandbox456')
    })

    test('should generate correct href for sync jobs', () => {
      const mockJob: TRunnerJob = {
        group: 'sync',
        metadata: {
          install_id: 'install123',
          component_id: 'comp456',
          deploy_id: 'deploy789',
        },
      } as TRunnerJob

      const href = getJobHref(mockJob)
      expect(href).toBe(
        'installs/install123/components/comp456/deploys/deploy789'
      )
    })

    test('should generate correct href for deploy jobs', () => {
      const mockJob: TRunnerJob = {
        group: 'deploy',
        metadata: {
          install_id: 'install123',
          component_id: 'comp456',
          deploy_id: 'deploy789',
        },
      } as TRunnerJob

      const href = getJobHref(mockJob)
      expect(href).toBe(
        'installs/install123/components/comp456/deploys/deploy789'
      )
    })

    test('should generate correct href for actions jobs', () => {
      const mockJob: TRunnerJob = {
        group: 'actions',
        metadata: {
          install_id: 'install123',
          action_workflow_id: 'workflow456',
          action_workflow_run_id: 'run789',
        },
      } as TRunnerJob

      const href = getJobHref(mockJob)
      expect(href).toBe('installs/install123/actions/workflow456/run789')
    })

    test('should return empty string for unknown job groups', () => {
      const mockJob: TRunnerJob = {
        group: 'unknown' as any,
        metadata: {},
      } as TRunnerJob

      const href = getJobHref(mockJob)
      expect(href).toBe('')
    })

    test('should handle missing metadata gracefully', () => {
      const mockJob: TRunnerJob = {
        group: 'build',
      } as TRunnerJob

      const href = getJobHref(mockJob)
      expect(href).toBe('apps/undefined/components/undefined/builds/undefined')
    })
  })

  describe('getJobName', () => {
    test('should return component name for build jobs', () => {
      const mockJob: TRunnerJob = {
        group: 'build',
        metadata: {
          component_name: 'api-service',
        },
      } as TRunnerJob

      const name = getJobName(mockJob)
      expect(name).toBe('api-service')
    })

    test('should return component name for sync jobs', () => {
      const mockJob: TRunnerJob = {
        group: 'sync',
        metadata: {
          component_name: 'web-app',
        },
      } as TRunnerJob

      const name = getJobName(mockJob)
      expect(name).toBe('web-app')
    })

    test('should return component name for deploy jobs', () => {
      const mockJob: TRunnerJob = {
        group: 'deploy',
        metadata: {
          component_name: 'database',
        },
      } as TRunnerJob

      const name = getJobName(mockJob)
      expect(name).toBe('database')
    })

    test('should return sandbox run type for sandbox jobs', () => {
      const mockJob: TRunnerJob = {
        group: 'sandbox',
        metadata: {
          sandbox_run_type: 'plan',
        },
      } as TRunnerJob

      const name = getJobName(mockJob)
      expect(name).toBe('plan')
    })

    test('should return action workflow name for actions jobs', () => {
      const mockJob: TRunnerJob = {
        group: 'actions',
        metadata: {
          action_workflow_name: 'Deploy to Production',
        },
      } as TRunnerJob

      const name = getJobName(mockJob)
      expect(name).toBe('Deploy to Production')
    })

    test('should return type for operations jobs', () => {
      const mockJob: TRunnerJob = {
        group: 'operations',
        type: 'cleanup',
      } as TRunnerJob

      const name = getJobName(mockJob)
      expect(name).toBe('cleanup')
    })

    test('should return Unknown for missing metadata', () => {
      const mockJob: TRunnerJob = {
        group: 'build',
        metadata: {},
      } as TRunnerJob

      const name = getJobName(mockJob)
      expect(name).toBe('Unknown')
    })

    test('should return Unknown for unknown job groups', () => {
      const mockJob: TRunnerJob = {
        group: 'unknown' as any,
      } as TRunnerJob

      const name = getJobName(mockJob)
      expect(name).toBe('Unknown')
    })
  })

  describe('getJobExecutionStatus', () => {
    test('should return correct status for build jobs', () => {
      const finishedJob: TRunnerJob = {
        group: 'build',
        status: 'finished',
      } as TRunnerJob

      const failedJob: TRunnerJob = {
        group: 'build',
        status: 'failed',
      } as TRunnerJob

      const inProgressJob: TRunnerJob = {
        group: 'build',
        status: 'in-progress',
      } as TRunnerJob

      expect(getJobExecutionStatus(finishedJob)).toBe(
        'component built successfully'
      )
      expect(getJobExecutionStatus(failedJob)).toBe('component build failed')
      expect(getJobExecutionStatus(inProgressJob)).toBe(
        'component build is being built'
      )
    })

    test('should return correct status for sandbox jobs', () => {
      const finishedJob: TRunnerJob = {
        group: 'sandbox',
        status: 'finished',
      } as TRunnerJob

      const failedJob: TRunnerJob = {
        group: 'sandbox',
        status: 'failed',
      } as TRunnerJob

      const queuedJob: TRunnerJob = {
        group: 'sandbox',
        status: 'queued',
      } as TRunnerJob

      expect(getJobExecutionStatus(finishedJob)).toBe(
        'sandbox provisioned successfully'
      )
      expect(getJobExecutionStatus(failedJob)).toBe(
        'sandbox provisioning failed'
      )
      expect(getJobExecutionStatus(queuedJob)).toBe(
        'sandbox provisioning queued'
      )
    })

    test('should return correct status for sync jobs', () => {
      const finishedJob: TRunnerJob = {
        group: 'sync',
        status: 'finished',
      } as TRunnerJob

      const timedOutJob: TRunnerJob = {
        group: 'sync',
        status: 'timed-out',
      } as TRunnerJob

      expect(getJobExecutionStatus(finishedJob)).toBe(
        'component synced successfully'
      )
      expect(getJobExecutionStatus(timedOutJob)).toBe(
        'component sync timed out'
      )
    })

    test('should return correct status for deploy jobs', () => {
      const finishedJob: TRunnerJob = {
        group: 'deploy',
        status: 'finished',
      } as TRunnerJob

      const inProgressJob: TRunnerJob = {
        group: 'deploy',
        status: 'in-progress',
      } as TRunnerJob

      const cancelledJob: TRunnerJob = {
        group: 'deploy',
        status: 'cancelled',
      } as TRunnerJob

      expect(getJobExecutionStatus(finishedJob)).toBe(
        'component deployed successfully'
      )
      expect(getJobExecutionStatus(inProgressJob)).toBe(
        'component is being deployed'
      )
      expect(getJobExecutionStatus(cancelledJob)).toBe(
        'component deployment canceled'
      )
    })

    test('should return correct status for actions jobs', () => {
      const finishedJob: TRunnerJob = {
        group: 'actions',
        status: 'finished',
      } as TRunnerJob

      const availableJob: TRunnerJob = {
        group: 'actions',
        status: 'available',
      } as TRunnerJob

      expect(getJobExecutionStatus(finishedJob)).toBe(
        'action completed successfully'
      )
      expect(getJobExecutionStatus(availableJob)).toBe('action starting soon')
    })

    test('should return correct status for operations jobs', () => {
      const finishedJob: TRunnerJob = {
        group: 'operations',
        status: 'finished',
      } as TRunnerJob

      const notAttemptedJob: TRunnerJob = {
        group: 'operations',
        status: 'not-attempted',
      } as TRunnerJob

      expect(getJobExecutionStatus(finishedJob)).toBe(
        'operation completed successfully'
      )
      expect(getJobExecutionStatus(notAttemptedJob)).toBe(
        'operation not attempted'
      )
    })

    test('should return Unknown for unknown job status', () => {
      const mockJob: TRunnerJob = {
        group: 'build',
        status: 'unknown_status' as any,
      } as TRunnerJob

      const status = getJobExecutionStatus(mockJob)
      expect(status).toBe('Unknown')
    })

    test('should return Unknown for unknown job group', () => {
      const mockJob: TRunnerJob = {
        group: 'unknown' as any,
        status: 'finished',
      } as TRunnerJob

      const status = getJobExecutionStatus(mockJob)
      expect(status).toBe('Unknown')
    })

    test('should handle missing group gracefully', () => {
      const mockJob: TRunnerJob = {
        status: 'finished',
      } as TRunnerJob

      const status = getJobExecutionStatus(mockJob)
      expect(status).toBe('Unknown')
    })
  })
})
