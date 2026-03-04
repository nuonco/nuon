import { describe, expect, test } from 'vitest'
import { normalizeAppInputGroups } from './app-utils'
import type { TAppConfig } from '@/types'

describe('app-utils', () => {
  describe('normalizeAppInputGroups', () => {
    test('should map inputs to their corresponding groups', () => {
      const groups: TAppConfig['input']['input_groups'] = [
        {
          id: 'group-1',
          name: 'Database Settings',
          description: 'Database configuration',
          index: 1,
        },
        {
          id: 'group-2', 
          name: 'Network Settings',
          description: 'Network configuration',
          index: 2,
        },
      ]

      const inputs: TAppConfig['input']['inputs'] = [
        {
          id: 'input-1',
          name: 'db_host',
          type: 'string',
          group_id: 'group-1',
          required: true,
        },
        {
          id: 'input-2',
          name: 'db_port',
          type: 'number', 
          group_id: 'group-1',
          required: false,
        },
        {
          id: 'input-3',
          name: 'vpc_cidr',
          type: 'string',
          group_id: 'group-2',
          required: true,
        },
      ]

      const result = normalizeAppInputGroups(groups, inputs)

      expect(result).toEqual([
        {
          id: 'group-1',
          name: 'Database Settings',
          description: 'Database configuration',
          index: 1,
          app_inputs: [
            {
              id: 'input-1',
              name: 'db_host',
              type: 'string',
              group_id: 'group-1',
              required: true,
            },
            {
              id: 'input-2',
              name: 'db_port',
              type: 'number',
              group_id: 'group-1', 
              required: false,
            },
          ],
        },
        {
          id: 'group-2',
          name: 'Network Settings',
          description: 'Network configuration',
          index: 2,
          app_inputs: [
            {
              id: 'input-3',
              name: 'vpc_cidr',
              type: 'string',
              group_id: 'group-2',
              required: true,
            },
          ],
        },
      ])
    })

    test('should handle groups with no matching inputs', () => {
      const groups: TAppConfig['input']['input_groups'] = [
        {
          id: 'group-1',
          name: 'Empty Group',
          description: 'Group with no inputs',
          index: 1,
        },
      ]

      const inputs: TAppConfig['input']['inputs'] = [
        {
          id: 'input-1',
          name: 'orphaned_input',
          type: 'string',
          group_id: 'non-existent-group',
          required: false,
        },
      ]

      const result = normalizeAppInputGroups(groups, inputs)

      expect(result).toEqual([
        {
          id: 'group-1',
          name: 'Empty Group',
          description: 'Group with no inputs',
          index: 1,
          app_inputs: [],
        },
      ])
    })

    test('should handle empty groups array', () => {
      const groups: TAppConfig['input']['input_groups'] = []
      const inputs: TAppConfig['input']['inputs'] = [
        {
          id: 'input-1',
          name: 'test_input',
          type: 'string',
          group_id: 'some-group',
          required: true,
        },
      ]

      const result = normalizeAppInputGroups(groups, inputs)

      expect(result).toEqual([])
    })

    test('should handle empty inputs array', () => {
      const groups: TAppConfig['input']['input_groups'] = [
        {
          id: 'group-1',
          name: 'Test Group',
          description: 'Test description',
          index: 1,
        },
      ]
      const inputs: TAppConfig['input']['inputs'] = []

      const result = normalizeAppInputGroups(groups, inputs)

      expect(result).toEqual([
        {
          id: 'group-1',
          name: 'Test Group',
          description: 'Test description',
          index: 1,
          app_inputs: [],
        },
      ])
    })

    test('should handle both empty arrays', () => {
      const groups: TAppConfig['input']['input_groups'] = []
      const inputs: TAppConfig['input']['inputs'] = []

      const result = normalizeAppInputGroups(groups, inputs)

      expect(result).toEqual([])
    })

    test('should preserve all group properties', () => {
      const groups: TAppConfig['input']['input_groups'] = [
        {
          id: 'group-1',
          name: 'Advanced Group',
          description: 'Advanced configuration options',
          index: 10,
          // Additional properties that might exist
          custom_field: 'custom_value',
        } as any,
      ]

      const inputs: TAppConfig['input']['inputs'] = []

      const result = normalizeAppInputGroups(groups, inputs)

      expect(result[0]).toEqual({
        id: 'group-1',
        name: 'Advanced Group', 
        description: 'Advanced configuration options',
        index: 10,
        custom_field: 'custom_value',
        app_inputs: [],
      })
    })
  })
})
