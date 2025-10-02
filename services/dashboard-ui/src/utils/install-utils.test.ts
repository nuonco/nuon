import { describe, expect, test } from 'vitest'
import {
  getInstallRunnerStatusTitle,
  getInstallSandboxStatusTitle,
  getInstallComponentsStatusTitle,
  getInstallStatusTitle,
} from './install-utils'

describe('install-utils', () => {
  describe('getInstallRunnerStatusTitle', () => {
    test('should return correct runner status titles', () => {
      expect(getInstallRunnerStatusTitle('active')).toBe('Runner is healthy')
      expect(getInstallRunnerStatusTitle('error')).toBe('Runner is unhealthy')
      expect(getInstallRunnerStatusTitle('pending')).toBe('Runner is pending')
      expect(getInstallRunnerStatusTitle('provisioning')).toBe(
        'Runner is provisioning'
      )
      expect(getInstallRunnerStatusTitle('deprovisioning')).toBe(
        'Runner is deprovisioning'
      )
      expect(getInstallRunnerStatusTitle('deprovisioned')).toBe(
        'Runner is deprovisioned'
      )
      expect(getInstallRunnerStatusTitle('reprovisioning')).toBe(
        'Runner is reprovisioning'
      )
      expect(getInstallRunnerStatusTitle('offline')).toBe('Runner is offline')
      expect(getInstallRunnerStatusTitle('awaiting-install-stack-run')).toBe(
        'Runner is awaiting install stack run'
      )
    })

    test('should return unknown status for unrecognized runner statuses', () => {
      expect(getInstallRunnerStatusTitle('unknown_status')).toBe(
        'Runner status is unknown'
      )
      expect(getInstallRunnerStatusTitle('')).toBe('Runner status is unknown')
      expect(getInstallRunnerStatusTitle('invalid')).toBe(
        'Runner status is unknown'
      )
    })

    test('should return known status for known runner status', () => {
      expect(getInstallRunnerStatusTitle('unknown')).toBe(
        'Runner status is unknown'
      )
    })
  })

  describe('getInstallSandboxStatusTitle', () => {
    test('should return correct sandbox status titles', () => {
      expect(getInstallSandboxStatusTitle('active')).toBe(
        'Sandbox is provisioned'
      )
      expect(getInstallSandboxStatusTitle('error')).toBe('Sandbox has an error')
      expect(getInstallSandboxStatusTitle('queued')).toBe('Sandbox is queued')
      expect(getInstallSandboxStatusTitle('provisioning')).toBe(
        'Sandbox is provisioning'
      )
      expect(getInstallSandboxStatusTitle('deprovisioning')).toBe(
        'Sandbox is deprovisioning'
      )
      expect(getInstallSandboxStatusTitle('deprovisioned')).toBe(
        'Sandbox is deprovisioned'
      )
      expect(getInstallSandboxStatusTitle('reprovisioning')).toBe(
        'Sandbox is reprovisioning'
      )
      expect(getInstallSandboxStatusTitle('access_error')).toBe(
        'Sandbox has an access error'
      )
      expect(getInstallSandboxStatusTitle('deleted')).toBe(
        'Sandbox has been deleted'
      )
      expect(getInstallSandboxStatusTitle('delete_failed')).toBe(
        'Sandbox deletion failed'
      )
      expect(getInstallSandboxStatusTitle('empty')).toBe('Sandbox is empty')
    })

    test('should return unknown status for unrecognized sandbox statuses', () => {
      expect(getInstallSandboxStatusTitle('unknown_status')).toBe(
        'Sandbox status is unknown'
      )
      expect(getInstallSandboxStatusTitle('')).toBe('Sandbox status is unknown')
      expect(getInstallSandboxStatusTitle('invalid')).toBe(
        'Sandbox status is unknown'
      )
    })

    test('should return known status for known sandbox status', () => {
      expect(getInstallSandboxStatusTitle('unknown')).toBe(
        'Sandbox status is unknown'
      )
    })
  })

  describe('getInstallComponentsStatusTitle', () => {
    test('should return correct components status titles', () => {
      expect(getInstallComponentsStatusTitle('active')).toBe(
        'Components are deployed'
      )
      expect(getInstallComponentsStatusTitle('inactive')).toBe(
        'Components are inactive'
      )
      expect(getInstallComponentsStatusTitle('error')).toBe(
        'Component has an error'
      )
      expect(getInstallComponentsStatusTitle('noop')).toBe(
        'Deployment had no changes'
      )
      expect(getInstallComponentsStatusTitle('planning')).toBe(
        'Deployment is planning'
      )
      expect(getInstallComponentsStatusTitle('syncing')).toBe(
        'Deployment is syncing'
      )
      expect(getInstallComponentsStatusTitle('executing')).toBe(
        'Deployment is executing'
      )
      expect(getInstallComponentsStatusTitle('cancelled')).toBe(
        'Deployment was cancelled'
      )
      expect(getInstallComponentsStatusTitle('pending')).toBe(
        'Deployment is pending'
      )
      expect(getInstallComponentsStatusTitle('queued')).toBe(
        'Deployment is queued'
      )
      expect(getInstallComponentsStatusTitle('pending-approval')).toBe(
        'Deployment is pending approval'
      )
      expect(getInstallComponentsStatusTitle('approval-denied')).toBe(
        'Deployment approval was denied'
      )
    })

    test('should return unknown status for unrecognized components statuses', () => {
      expect(getInstallComponentsStatusTitle('unknown_status')).toBe(
        'Deployment status is unknown'
      )
      expect(getInstallComponentsStatusTitle('')).toBe(
        'Deployment status is unknown'
      )
      expect(getInstallComponentsStatusTitle('invalid')).toBe(
        'Deployment status is unknown'
      )
    })

    test('should return known status for known components status', () => {
      expect(getInstallComponentsStatusTitle('unknown')).toBe(
        'Deployment status is unknown'
      )
    })
  })

  describe('getInstallStatusTitle', () => {
    test('should route to runner status for runner_status key', () => {
      const result = getInstallStatusTitle('runner_status', 'active')
      expect(result).toBe('Runner is healthy')
    })

    test('should route to sandbox status for sandbox_status key', () => {
      const result = getInstallStatusTitle('sandbox_status', 'active')
      expect(result).toBe('Sandbox is provisioned')
    })

    test('should route to components status for composite_component_status key', () => {
      const result = getInstallStatusTitle(
        'composite_component_status',
        'active'
      )
      expect(result).toBe('Components are deployed')
    })

    test('should return waiting message for unknown status keys', () => {
      const result = getInstallStatusTitle('unknown_key', 'some_status')
      expect(result).toBe('Waiting on status')
    })

    test('should return waiting message for empty status key', () => {
      const result = getInstallStatusTitle('', 'active')
      expect(result).toBe('Waiting on status')
    })

    test('should handle complex runner statuses through routing', () => {
      const result = getInstallStatusTitle(
        'runner_status',
        'awaiting-install-stack-run'
      )
      expect(result).toBe('Runner is awaiting install stack run')
    })

    test('should handle complex sandbox statuses through routing', () => {
      const result = getInstallStatusTitle('sandbox_status', 'access_error')
      expect(result).toBe('Sandbox has an access error')
    })

    test('should handle complex components statuses through routing', () => {
      const result = getInstallStatusTitle(
        'composite_component_status',
        'pending-approval'
      )
      expect(result).toBe('Deployment is pending approval')
    })

    test('should handle unknown statuses through routing', () => {
      const result1 = getInstallStatusTitle('runner_status', 'unknown_status')
      const result2 = getInstallStatusTitle('sandbox_status', 'unknown_status')
      const result3 = getInstallStatusTitle(
        'composite_component_status',
        'unknown_status'
      )

      expect(result1).toBe('Runner status is unknown')
      expect(result2).toBe('Sandbox status is unknown')
      expect(result3).toBe('Deployment status is unknown')
    })
  })
})
