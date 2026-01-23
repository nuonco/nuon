import { describe, expect, test } from 'vitest'
import { getStatusTheme } from './vcs-connection-utils'

describe('vcs-connection-utils', () => {
  describe('getStatusTheme', () => {
    test('should return success theme for active status', () => {
      expect(getStatusTheme('active')).toBe('success')
    })

    test('should return error theme for suspended status', () => {
      expect(getStatusTheme('suspended')).toBe('error')
    })

    test('should return warn theme for unknown status', () => {
      expect(getStatusTheme('unknown')).toBe('warn')
    })

    test('should return neutral theme for any other status', () => {
      // Test default case with various invalid inputs
      expect(getStatusTheme('pending' as any)).toBe('neutral')
      expect(getStatusTheme('inactive' as any)).toBe('neutral')
      expect(getStatusTheme('' as any)).toBe('neutral')
      expect(getStatusTheme('random-status' as any)).toBe('neutral')
    })
  })
})
