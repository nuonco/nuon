import { describe, expect, test } from 'vitest'
import { getApprovalType, getApprovalResponseType } from './approval-utils'

describe('approval-utils', () => {
  describe('getApprovalType', () => {
    test('should return correct approval type strings', () => {
      expect(getApprovalType('approve-all')).toBe('all changes approved')
      expect(getApprovalType('terraform_plan')).toBe('terraform')
      expect(getApprovalType('kubernetes_manifest_approval')).toBe('kubernetes')
      expect(getApprovalType('helm_approval')).toBe('helm')
      expect(getApprovalType('noop')).toBe('no-op')
    })

    test('should handle all defined approval types', () => {
      // Test all the approval types from the APPROVAL_TYPE record
      const approvalTypes = [
        'approve-all',
        'terraform_plan',
        'kubernetes_manifest_approval',
        'helm_approval',
        'noop',
      ] as const

      approvalTypes.forEach((type) => {
        const result = getApprovalType(type)
        expect(typeof result).toBe('string')
        expect(result.length).toBeGreaterThan(0)
      })
    })
  })

  describe('getApprovalResponseType', () => {
    test('should return correct response type strings', () => {
      expect(getApprovalResponseType('approve')).toBe('approved')
      expect(getApprovalResponseType('auto-approve')).toBe('auto-approved')
      expect(getApprovalResponseType('deny')).toBe('denied')
      expect(getApprovalResponseType('retry')).toBe('retired')
      expect(getApprovalResponseType('skip')).toBe('skipped')
    })

    test('should handle all defined response types', () => {
      // Test all the response types from the RESPONSE_TYPE record
      const responseTypes = [
        'approve',
        'auto-approve',
        'deny',
        'retry',
        'skip',
      ] as const

      responseTypes.forEach((type) => {
        const result = getApprovalResponseType(type)
        expect(typeof result).toBe('string')
        expect(result.length).toBeGreaterThan(0)
      })
    })

    test('should return specific mappings correctly', () => {
      // Test specific mappings that might be confusing
      expect(getApprovalResponseType('retry')).toBe('retired') // Note: "retired" not "retried"
      expect(getApprovalResponseType('auto-approve')).toBe('auto-approved')
    })
  })

  describe('type mappings validation', () => {
    test('approval types should map to expected values', () => {
      const expectedMappings = {
        'approve-all': 'all changes approved',
        terraform_plan: 'terraform',
        kubernetes_manifest_approval: 'kubernetes',
        helm_approval: 'helm',
        noop: 'no-op',
      }

      Object.entries(expectedMappings).forEach(([input, expected]) => {
        expect(getApprovalType(input as any)).toBe(expected)
      })
    })

    test('response types should map to expected values', () => {
      const expectedMappings = {
        approve: 'approved',
        'auto-approve': 'auto-approved',
        deny: 'denied',
        retry: 'retired', // Note the difference
        skip: 'skipped',
      }

      Object.entries(expectedMappings).forEach(([input, expected]) => {
        expect(getApprovalResponseType(input as any)).toBe(expected)
      })
    })
  })
})
