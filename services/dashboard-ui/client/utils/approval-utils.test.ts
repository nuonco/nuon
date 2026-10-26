import { describe, expect, test } from 'bun:test'
import {
  approvalResponder,
  describeApprovalResponse,
  getApprovalType,
  getApprovalResponseType,
} from './approval-utils'

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
      expect(getApprovalResponseType('retry')).toBe('retried')
      expect(getApprovalResponseType('skip')).toBe('skipped')
    })

    test('should handle all defined response types', () => {
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
      expect(getApprovalResponseType('retry')).toBe('retried')
      expect(getApprovalResponseType('auto-approve')).toBe('auto-approved')
    })
  })

  describe('describeApprovalResponse', () => {
    test('names the person who approved a helm plan', () => {
      expect(
        describeApprovalResponse({
          approvalType: 'helm_approval',
          responseType: 'approve',
          approver: 'ada@example.com',
        })
      ).toBe('This helm plan was approved by ada@example.com.')
    })

    test('keeps the anonymous copy when nobody is recorded', () => {
      expect(
        describeApprovalResponse({
          approvalType: 'helm_approval',
          responseType: 'approve',
        })
      ).toBe('This helm plan was approved.')
    })

    test('does not attribute an auto-approval to an account', () => {
      expect(
        describeApprovalResponse({
          approvalType: 'terraform_plan',
          responseType: 'auto-approve',
          approver: 'runner@example.com',
        })
      ).toBe('This terraform plan was auto-approved.')
    })
  })

  describe('approvalResponder', () => {
    test('prefers email, then name, and skips auto-approve', () => {
      expect(
        approvalResponder({
          type: 'approve',
          created_by: { email: 'ada@example.com', name: 'Ada' },
        })
      ).toBe('ada@example.com')
      expect(
        approvalResponder({
          type: 'deny',
          created_by: { name: 'Ada' },
        })
      ).toBe('Ada')
      expect(
        approvalResponder({
          type: 'auto-approve',
          created_by: { email: 'runner@example.com' },
        })
      ).toBeUndefined()
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
        retry: 'retried',
        skip: 'skipped',
      }

      Object.entries(expectedMappings).forEach(([input, expected]) => {
        expect(getApprovalResponseType(input as any)).toBe(expected)
      })
    })
  })
})
