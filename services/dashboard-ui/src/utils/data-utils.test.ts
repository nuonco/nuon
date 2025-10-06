import { describe, expect, test } from 'vitest'
import { objectToKeyValueArray } from './data-utils'

describe('data-utils', () => {
  describe('objectToKeyValueArray', () => {
    test('should convert object to key-value array with types', () => {
      const obj = {
        name: 'John',
        age: 30,
        active: true,
      }

      const result = objectToKeyValueArray(obj)

      expect(result).toEqual([
        { key: 'name', value: 'John', type: 'string' },
        { key: 'age', value: '30', type: 'number' },
        { key: 'active', value: 'true', type: 'boolean' },
      ])
    })

    test('should handle empty object', () => {
      const result = objectToKeyValueArray({})
      expect(result).toEqual([])
    })

    test('should handle null and undefined values', () => {
      const obj = {
        nullValue: null,
        undefinedValue: undefined,
      }

      const result = objectToKeyValueArray(obj)

      expect(result).toEqual([
        { key: 'nullValue', value: 'null', type: 'null' },
        { key: 'undefinedValue', value: 'undefined', type: 'undefined' },
      ])
    })

    test('should handle array values', () => {
      const obj = {
        numbers: [1, 2, 3],
        strings: ['a', 'b', 'c'],
        mixed: [1, 'two', true, null],
        empty: [],
      }

      const result = objectToKeyValueArray(obj)

      expect(result).toEqual([
        { key: 'numbers', value: '[1, 2, 3]', type: 'array' },
        { key: 'strings', value: '[a, b, c]', type: 'array' },
        { key: 'mixed', value: '[1, two, true, null]', type: 'array' },
        { key: 'empty', value: '[]', type: 'array' },
      ])
    })

    test('should handle object values with JSON formatting', () => {
      const obj = {
        simpleObject: { key: 'value' },
        nestedObject: {
          level1: {
            level2: 'deep value',
          },
        },
      }

      const result = objectToKeyValueArray(obj)

      expect(result).toEqual([
        {
          key: 'simpleObject',
          value: '{\n  "key": "value"\n}',
          type: 'object',
        },
        {
          key: 'nestedObject',
          value: '{\n  "level1": {\n    "level2": "deep value"\n  }\n}',
          type: 'object',
        },
      ])
    })

    test('should handle function values', () => {
      const obj = {
        myFunction: () => 'test',
        namedFunction: function testFunc() {
          return 'named'
        },
      }

      const result = objectToKeyValueArray(obj)

      expect(result).toEqual([
        { key: 'myFunction', value: '[Function]', type: 'function' },
        { key: 'namedFunction', value: '[Function]', type: 'function' },
      ])
    })

    test('should handle objects with circular references', () => {
      const obj: any = { name: 'test' }
      obj.circular = obj // Create circular reference

      const result = objectToKeyValueArray(obj)

      expect(result).toEqual([
        { key: 'name', value: 'test', type: 'string' },
        {
          key: 'circular',
          value: '[Object - Unable to serialize]',
          type: 'object',
        },
      ])
    })

    test('should handle mixed data types', () => {
      const obj = {
        str: 'hello',
        num: 42,
        bool: false,
        nullVal: null,
        undefinedVal: undefined,
        obj: { nested: 'value' },
        arr: [1, 2, 3],
        func: () => 'test',
      }

      const result = objectToKeyValueArray(obj)

      expect(result).toEqual([
        { key: 'str', value: 'hello', type: 'string' },
        { key: 'num', value: '42', type: 'number' },
        { key: 'bool', value: 'false', type: 'boolean' },
        { key: 'nullVal', value: 'null', type: 'null' },
        { key: 'undefinedVal', value: 'undefined', type: 'undefined' },
        { key: 'obj', value: '{\n  "nested": "value"\n}', type: 'object' },
        { key: 'arr', value: '[1, 2, 3]', type: 'array' },
        { key: 'func', value: '[Function]', type: 'function' },
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
        { key: 'key with spaces', value: 'value with spaces', type: 'string' },
        { key: 'key@symbol', value: 'value@symbol', type: 'string' },
        { key: 'key-dash', value: 'value-dash', type: 'string' },
      ])
    })

    test('should handle nested arrays with different types', () => {
      const obj = {
        nestedArrays: [
          [1, 2],
          ['a', 'b'],
          [true, false],
        ],
        objectsInArray: [{ id: 1 }, { id: 2 }],
      }

      const result = objectToKeyValueArray(obj)

      expect(result[0]).toEqual({
        key: 'nestedArrays',
        value: '[[1, 2], [a, b], [true, false]]',
        type: 'array',
      })

      expect(result[1]).toEqual({
        key: 'objectsInArray',
        value: '[{\n  "id": 1\n}, {\n  "id": 2\n}]',
        type: 'array',
      })
    })
  })
})
