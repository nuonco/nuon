import { describe, expect, test, beforeAll, afterAll, vi } from 'vitest'
import { isLessThan15SecondsOld } from './time-utils'
import { DateTime } from 'luxon'

describe('time-utils', () => {
  describe('isLessThan15SecondsOld', () => {
    beforeAll(() => {
      // Mock DateTime.now() to return a fixed time for consistent testing
      vi.spyOn(DateTime, 'now').mockReturnValue(
        DateTime.fromISO('2023-01-01T10:00:00.000Z')
      )
    })

    afterAll(() => {
      vi.restoreAllMocks()
    })

    test('should return true for timestamp less than 15 seconds old', () => {
      const timestamp = '2023-01-01T09:59:50.000Z' // 10 seconds ago
      expect(isLessThan15SecondsOld(timestamp)).toBe(true)
    })

    test('should return false for timestamp exactly 15 seconds old', () => {
      const timestamp = '2023-01-01T09:59:45.000Z' // exactly 15 seconds ago
      expect(isLessThan15SecondsOld(timestamp)).toBe(false)
    })

    test('should return false for timestamp more than 15 seconds old', () => {
      const timestamp = '2023-01-01T09:59:40.000Z' // 20 seconds ago
      expect(isLessThan15SecondsOld(timestamp)).toBe(false)
    })

    test('should return false for future timestamps', () => {
      const timestamp = '2023-01-01T10:00:10.000Z' // 10 seconds in the future
      expect(isLessThan15SecondsOld(timestamp)).toBe(false)
    })

    test('should return true for current timestamp', () => {
      const timestamp = '2023-01-01T10:00:00.000Z' // exactly now
      expect(isLessThan15SecondsOld(timestamp)).toBe(true)
    })

    test('should handle invalid ISO strings gracefully', () => {
      // This test assumes DateTime.fromISO handles invalid strings by returning invalid DateTime
      const result = isLessThan15SecondsOld('invalid-date')
      expect(typeof result).toBe('boolean')
    })
  })
})
