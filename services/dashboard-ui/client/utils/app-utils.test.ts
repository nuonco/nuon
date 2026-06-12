import { describe, expect, test } from 'bun:test'
import { hasNewerAppConfig, hasStackConfigChanged, normalizeAppInputGroups } from './app-utils'
import type { TAppConfig, TInstall } from '@/types'

describe('app-utils', () => {
  describe('hasNewerAppConfig', () => {
    test('returns true when latest config id differs from install app_config_id', () => {
      const latestConfig = { id: 'config-2' } as TAppConfig
      const install = { app_config_id: 'config-1' } as TInstall
      expect(hasNewerAppConfig(latestConfig, install)).toBe(true)
    })

    test('returns false when ids match', () => {
      const latestConfig = { id: 'config-1' } as TAppConfig
      const install = { app_config_id: 'config-1' } as TInstall
      expect(hasNewerAppConfig(latestConfig, install)).toBe(false)
    })

    test('returns false when latestConfig is undefined', () => {
      const install = { app_config_id: 'config-1' } as TInstall
      expect(hasNewerAppConfig(undefined, install)).toBe(false)
    })

    test('returns false when install is undefined', () => {
      const latestConfig = { id: 'config-2' } as TAppConfig
      expect(hasNewerAppConfig(latestConfig, undefined)).toBe(false)
    })

    test('returns false when install has no app_config_id', () => {
      const latestConfig = { id: 'config-2' } as TAppConfig
      const install = {} as TInstall
      expect(hasNewerAppConfig(latestConfig, install)).toBe(false)
    })

    test('returns false when both are undefined', () => {
      expect(hasNewerAppConfig(undefined, undefined)).toBe(false)
    })
  })

  describe('hasStackConfigChanged', () => {
    const baseStack = {
      type: 'aws-cloudformation',
      name: 'my-stack',
      runner_nested_template_url: 'https://example.com/runner.yaml',
      vpc_nested_template_url: 'https://example.com/vpc.yaml',
    }

    test('returns false when stacks are identical', () => {
      const a = { stack: { ...baseStack } } as TAppConfig
      const b = { stack: { ...baseStack } } as TAppConfig
      expect(hasStackConfigChanged(a, b)).toBe(false)
    })

    test('returns true when stack type differs', () => {
      const a = { stack: { ...baseStack } } as TAppConfig
      const b = { stack: { ...baseStack, type: 'gcp-terraform' } } as TAppConfig
      expect(hasStackConfigChanged(a, b)).toBe(true)
    })

    test('returns true when stack name differs', () => {
      const a = { stack: { ...baseStack } } as TAppConfig
      const b = { stack: { ...baseStack, name: 'other-stack' } } as TAppConfig
      expect(hasStackConfigChanged(a, b)).toBe(true)
    })

    test('returns true when runner_nested_template_url differs', () => {
      const a = { stack: { ...baseStack } } as TAppConfig
      const b = { stack: { ...baseStack, runner_nested_template_url: 'https://example.com/new.yaml' } } as TAppConfig
      expect(hasStackConfigChanged(a, b)).toBe(true)
    })

    test('returns true when vpc_nested_template_url differs', () => {
      const a = { stack: { ...baseStack } } as TAppConfig
      const b = { stack: { ...baseStack, vpc_nested_template_url: 'https://example.com/new-vpc.yaml' } } as TAppConfig
      expect(hasStackConfigChanged(a, b)).toBe(true)
    })

    test('returns false when currentConfig is undefined', () => {
      const b = { stack: { ...baseStack } } as TAppConfig
      expect(hasStackConfigChanged(undefined, b)).toBe(false)
    })

    test('returns false when latestConfig is undefined', () => {
      const a = { stack: { ...baseStack } } as TAppConfig
      expect(hasStackConfigChanged(a, undefined)).toBe(false)
    })

    test('returns false when currentConfig has no stack', () => {
      const a = {} as TAppConfig
      const b = { stack: { ...baseStack } } as TAppConfig
      expect(hasStackConfigChanged(a, b)).toBe(false)
    })

    test('returns false when both are undefined', () => {
      expect(hasStackConfigChanged(undefined, undefined)).toBe(false)
    })
  })

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

    test('should omit groups with no matching inputs', () => {
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

      expect(result).toEqual([])
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

      expect(result).toEqual([])
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

      const inputs: TAppConfig['input']['inputs'] = [
        {
          id: 'input-1',
          name: 'test_input',
          type: 'string',
          group_id: 'group-1',
          required: true,
        },
      ]

      const result = normalizeAppInputGroups(groups, inputs)

      expect(result[0]).toEqual(expect.objectContaining({
        id: 'group-1',
        name: 'Advanced Group',
        description: 'Advanced configuration options',
        index: 10,
        custom_field: 'custom_value',
      }))
      expect(result[0].app_inputs).toHaveLength(1)
    })
  })
})
