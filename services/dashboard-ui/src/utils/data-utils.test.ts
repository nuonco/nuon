import { describe, expect, test } from 'vitest'
import { objectToKeyValueArray } from './data-utils'

describe('data-utils', () => {
  describe('objectToKeyValueArray', () => {
    test('should convert object to key-value array', () => {
      const obj = {
        name: 'John',
        age: 30,
        active: true,
      }

      const result = objectToKeyValueArray(obj)

      expect(result).toEqual([
        { key: 'name', value: 'John' },
        { key: 'age', value: '30' },
        { key: 'active', value: 'true' },
      ])
    })

    test('should handle empty object', () => {
      const result = objectToKeyValueArray({})
      expect(result).toEqual([])
    })

    test('should convert all values to strings', () => {
      const obj = {
        str: 'hello',
        num: 42,
        bool: false,
        null: null,
        undefined: undefined,
        obj: { nested: 'value' },
        arr: [1, 2, 3],
      }

      const result = objectToKeyValueArray(obj)

      expect(result).toEqual([
        { key: 'str', value: 'hello' },
        { key: 'num', value: '42' },
        { key: 'bool', value: 'false' },
        { key: 'null', value: 'null' },
        { key: 'undefined', value: 'undefined' },
        { key: 'obj', value: '[object Object]' },
        { key: 'arr', value: '1,2,3' },
      ])
    })

    test('should handle special characters in keys and values', () => {
      const obj = {
        'key with spaces': 'value with spaces',
        'key@symbol': 'value@symbol',
        'key-dash': 'value-dash',
      }

      const result = objectToKeyValueArray(obj)

      expect(result).toEqual([
        { key: 'key with spaces', value: 'value with spaces' },
        { key: 'key@symbol', value: 'value@symbol' },
        { key: 'key-dash', value: 'value-dash' },
      ])
    })
  })
})
