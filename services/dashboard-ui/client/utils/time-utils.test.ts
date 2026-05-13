// @ts-nocheck

import { describe, expect, test, beforeAll, afterAll, spyOn } from 'bun:test'
import { isRecentTimestamp } from './time-utils'
import { DateTime } from 'luxon'

describe('time-utils', () => {
  describe('isRecentTimestamp', () => {
    let nowSpy: ReturnType<typeof spyOn>

    beforeAll(() => {
      nowSpy = spyOn(DateTime, 'now').mockReturnValue(
        DateTime.fromISO('2023-01-01T10:00:00.000Z')
      )
    })

    afterAll(() => {
      nowSpy.mockRestore()
    })

    test('should return true for timestamp within default 60s window', () => {
      const timestamp = '2023-01-01T09:59:10.000Z' // 50 seconds ago
      expect(isRecentTimestamp(timestamp)).toBe(true)
    })

    test('should return false for timestamp older than default 60s window', () => {
      const timestamp = '2023-01-01T09:58:50.000Z' // 70 seconds ago
      expect(isRecentTimestamp(timestamp)).toBe(false)
    })

    test('should return false for timestamp exactly at maxAgeSeconds', () => {
      const timestamp = '2023-01-01T09:59:00.000Z' // exactly 60 seconds ago
      expect(isRecentTimestamp(timestamp)).toBe(false)
    })

    test('should respect custom maxAgeSeconds', () => {
      const timestamp = '2023-01-01T09:59:50.000Z' // 10 seconds ago
      expect(isRecentTimestamp(timestamp, 15)).toBe(true)
      expect(isRecentTimestamp(timestamp, 5)).toBe(false)
    })

    test('should return false for future timestamps', () => {
      const timestamp = '2023-01-01T10:00:10.000Z' // 10 seconds in the future
      expect(isRecentTimestamp(timestamp)).toBe(false)
    })

    test('should return true for current timestamp', () => {
      const timestamp = '2023-01-01T10:00:00.000Z' // exactly now
      expect(isRecentTimestamp(timestamp)).toBe(true)
    })

    test('should return false for undefined', () => {
      expect(isRecentTimestamp(undefined)).toBe(false)
    })

    test('should handle invalid ISO strings gracefully', () => {
      const result = isRecentTimestamp('invalid-date')
      expect(typeof result).toBe('boolean')
    })
  })
})
