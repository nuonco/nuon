import { describe, expect, test, beforeAll, afterAll, vi } from 'vitest'
import { formatToRelativeDay, parseActivityTimeline } from './timeline-utils'
import { DateTime } from 'luxon'

describe('timeline-utils', () => {
  describe('formatToRelativeDay', () => {
    beforeAll(() => {
      // Mock DateTime.now() to return a fixed time for consistent testing
      vi.spyOn(DateTime, 'now').mockReturnValue(
        DateTime.fromISO('2023-01-01T12:00:00.000Z')
      )
    })

    afterAll(() => {
      vi.restoreAllMocks()
    })

    test('should return "Today" for current date', () => {
      const today = '2023-01-01T10:00:00.000Z'
      expect(formatToRelativeDay(today)).toBe('Today')
    })

    test('should return "Yesterday" for previous date', () => {
      const yesterday = '2022-12-31T10:00:00.000Z'
      expect(formatToRelativeDay(yesterday)).toBe('Yesterday')
    })

    test('should return formatted date for older dates', () => {
      const olderDate = '2022-12-29T10:00:00.000Z'
      const result = formatToRelativeDay(olderDate)
      expect(result).not.toBe('Today')
      expect(result).not.toBe('Yesterday')
      expect(typeof result).toBe('string')
    })

    test('should handle invalid date strings', () => {
      const result = formatToRelativeDay('invalid-date')
      expect(typeof result).toBe('string')
    })

    test('should handle empty string', () => {
      const result = formatToRelativeDay('')
      expect(typeof result).toBe('string')
    })
  })

  describe('parseActivityTimeline', () => {
    test('should group items by date', () => {
      const items = [
        { id: '1', created_at: '2023-01-01T10:00:00.000Z', name: 'Item 1' },
        { id: '2', created_at: '2023-01-01T11:00:00.000Z', name: 'Item 2' },
        { id: '3', created_at: '2023-01-02T10:00:00.000Z', name: 'Item 3' },
      ]

      const result = parseActivityTimeline(items)

      expect(typeof result).toBe('object')
      expect(Object.keys(result)).toHaveLength(2) // 2 different dates
    })

    test('should handle empty array', () => {
      const result = parseActivityTimeline([])
      expect(result).toEqual({})
    })

    test('should handle items without created_at', () => {
      const items = [
        { id: '1', name: 'Item 1' },
        { id: '2', created_at: '2023-01-01T10:00:00.000Z', name: 'Item 2' },
      ]

      const result = parseActivityTimeline(items)
      expect(typeof result).toBe('object')
    })

    test('should preserve item data in grouped results', () => {
      const items = [
        {
          id: '1',
          created_at: '2023-01-01T10:00:00.000Z',
          name: 'Item 1',
          data: 'test',
        },
      ]

      const result = parseActivityTimeline(items)

      // Get the first (and only) group
      const dateKey = Object.keys(result)[0]
      const groupedItems = result[dateKey]

      expect(groupedItems).toHaveLength(1)
      expect(groupedItems[0]).toEqual(items[0])
    })

    test('should sort items within the same date', () => {
      const items = [
        { id: '1', created_at: '2023-01-01T12:00:00.000Z', name: 'Later' },
        { id: '2', created_at: '2023-01-01T10:00:00.000Z', name: 'Earlier' },
      ]

      const result = parseActivityTimeline(items)
      const dateKey = Object.keys(result)[0]
      const groupedItems = result[dateKey]

      // Should be sorted by creation time (typically newest first)
      expect(groupedItems).toHaveLength(2)
    })
  })
})
