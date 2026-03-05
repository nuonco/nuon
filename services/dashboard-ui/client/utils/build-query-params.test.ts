import { describe, expect, test } from 'vitest'
import { buildQueryParams } from './build-query-params'

describe('build-query-params', () => {
  describe('buildQueryParams', () => {
    test('should return empty string for empty object', () => {
      expect(buildQueryParams({})).toBe('')
    })

    test('should filter out null and undefined values', () => {
      const params = {
        name: 'test',
        value: null,
        other: undefined,
        count: 0,
      }
      expect(buildQueryParams(params)).toBe('?name=test&count=0')
    })

    test('should handle string values', () => {
      const params = { search: 'hello world', type: 'user' }
      expect(buildQueryParams(params)).toBe('?search=hello+world&type=user')
    })

    test('should handle number values', () => {
      const params = { page: 1, limit: 10 }
      expect(buildQueryParams(params)).toBe('?page=1&limit=10')
    })

    test('should handle boolean values', () => {
      const params = { active: true, deleted: false }
      expect(buildQueryParams(params)).toBe('?active=true&deleted=false')
    })

    test('should handle array values', () => {
      const params = { tags: ['dev', 'prod'] }
      expect(buildQueryParams(params)).toBe('?tags=dev%2Cprod')
    })

    test('should URL encode special characters', () => {
      const params = { query: 'hello & world', symbol: '@#$%' }
      expect(buildQueryParams(params)).toBe(
        '?query=hello+%26+world&symbol=%40%23%24%25'
      )
    })

    test('should return empty string when all values are null/undefined', () => {
      const params = { a: null, b: undefined }
      expect(buildQueryParams(params)).toBe('')
    })
  })
})
