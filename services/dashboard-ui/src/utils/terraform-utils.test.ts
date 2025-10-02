import { describe, expect, test } from 'vitest'
import {
  cleanString,
  isStringYaml,
  isStringJson,
  isTerraformEscapedYaml,
  isTerraformArrayWithYaml,
  detectValueFormat,
  isOutputAfterUnknown,
  parseTerraformPlan,
} from './terraform-utils'
import type { TTerraformPlan } from '@/types'

describe('terraform-utils', () => {
  describe('cleanString', () => {
    test('should remove surrounding quotes', () => {
      expect(cleanString('"hello world"')).toBe('hello world')
    })

    test('should replace escaped newlines with real newlines', () => {
      expect(cleanString('line1\\nline2')).toBe('line1\nline2')
    })

    test('should handle null and undefined', () => {
      expect(cleanString(null as any)).toBe('')
      expect(cleanString(undefined as any)).toBe('')
    })

    test('should handle string without quotes', () => {
      expect(cleanString('test string')).toBe('test string')
    })

    test('should handle complex escaped string', () => {
      expect(cleanString('"Hello\\nWorld\\nTest"')).toBe('Hello\nWorld\nTest')
    })

    test('should only remove quotes if both start and end with quotes', () => {
      expect(cleanString('"hello world')).toBe('"hello world')
      expect(cleanString('hello world"')).toBe('hello world"')
    })
  })

  describe('isStringYaml', () => {
    test('should return true for valid YAML', () => {
      expect(isStringYaml('key: value\nother: data')).toBe(true)
      expect(isStringYaml('- item1\n- item2')).toBe(true)
    })

    test('should return false for invalid YAML', () => {
      expect(isStringYaml('not yaml')).toBe(false)
      expect(isStringYaml('{ invalid yaml')).toBe(false)
    })

    test('should return false for non-object YAML', () => {
      expect(isStringYaml('just a string')).toBe(false)
      expect(isStringYaml('123')).toBe(false)
    })
  })

  describe('isStringJson', () => {
    test('should return true for valid JSON objects', () => {
      expect(isStringJson('{"key": "value"}')).toBe(true)
      expect(isStringJson('{"nested": {"data": "value"}}')).toBe(true)
    })

    test('should return false for invalid JSON', () => {
      expect(isStringJson('not json')).toBe(false)
      expect(isStringJson('{invalid json}')).toBe(false)
    })

    test('should return false for non-object JSON', () => {
      expect(isStringJson('"just a string"')).toBe(false)
      expect(isStringJson('123')).toBe(false)
    })
  })

  describe('isTerraformEscapedYaml', () => {
    test('should detect escaped YAML content', () => {
      const input =
        '\\"apiVersion\\": \\"v1\\"\\n\\"kind\\": \\"ConfigMap\\"\\n\\"metadata\\":\\n  \\"name\\": \\"test\\"'
      const result = isTerraformEscapedYaml(input)

      expect(result.isTerraformYaml).toBe(true)
      expect(result.yamlContent).toBe(
        '"apiVersion": "v1"\n"kind": "ConfigMap"\n"metadata":\n  "name": "test"'
      )
    })

    test('should handle strings with escaped quotes and YAML structure', () => {
      const input =
        '\\"key\\": \\"value\\"\\n\\"other\\": \\"data\\"\\n  \\"nested\\": \\"item\\"'
      const result = isTerraformEscapedYaml(input)

      expect(result.isTerraformYaml).toBe(true)
    })

    test('should return false for non-YAML content', () => {
      const result = isTerraformEscapedYaml('just plain text')
      expect(result.isTerraformYaml).toBe(false)
    })

    test('should return false for strings without escaped content', () => {
      const result = isTerraformEscapedYaml('key: value')
      expect(result.isTerraformYaml).toBe(false)
    })
  })

  describe('isTerraformArrayWithYaml', () => {
    test('should detect array with YAML content', () => {
      const input = '["apiVersion: v1\\nkind: Pod\\nmetadata:\\n  name: test"]'
      const result = isTerraformArrayWithYaml(input)

      expect(result.isArrayYaml).toBe(true)
      expect(result.yamlContent).toContain('apiVersion: v1')
    })

    test('should handle arrays with quoted YAML', () => {
      const input = '["\\"key\\": \\"value\\"\\n\\"list\\":\\n  - item"]'
      const result = isTerraformArrayWithYaml(input)

      expect(result.isArrayYaml).toBe(true)
    })

    test('should return false for non-arrays', () => {
      const result = isTerraformArrayWithYaml('{"not": "array"}')
      expect(result.isArrayYaml).toBe(false)
    })

    test('should return false for arrays without YAML', () => {
      const result = isTerraformArrayWithYaml('["just text"]')
      expect(result.isArrayYaml).toBe(false)
    })

    test('should return false for empty arrays', () => {
      const result = isTerraformArrayWithYaml('[]')
      expect(result.isArrayYaml).toBe(false)
    })
  })

  describe('detectValueFormat', () => {
    test('should detect Terraform escaped YAML', () => {
      const input = 'apiVersion: v1\\nkind: ConfigMap\\ndata:\\n  key: value'
      const result = detectValueFormat(input)

      expect(result.language).toBe('yaml')
      expect(result.showLineNumbers).toBe(true)
      expect(result.displayValue).toContain('apiVersion: v1\nkind: ConfigMap')
    })

    test('should detect JSON and format it', () => {
      const input = '{"key":"value","nested":{"data":"test"}}'
      const result = detectValueFormat(input)

      expect(result.language).toBe('json')
      expect(result.showLineNumbers).toBe(true)
      expect(result.displayValue).toContain('  ') // Should be formatted with indentation
    })

    test('should detect YAML', () => {
      const input = 'key: value\nlist:\n  - item1\n  - item2'
      const result = detectValueFormat(input)

      expect(result.language).toBe('yaml')
      expect(result.showLineNumbers).toBe(true)
      expect(result.displayValue).toBe(input)
    })

    test('should detect array with YAML', () => {
      const input = '["key: value\\nother: data"]'
      const result = detectValueFormat(input)

      expect(result.language).toBe('yaml')
      expect(result.showLineNumbers).toBe(true)
    })

    test('should default to shell format for plain text', () => {
      const input = 'just plain text'
      const result = detectValueFormat(input)

      expect(result.language).toBe('sh')
      expect(result.showLineNumbers).toBe(false)
      expect(result.displayValue).toBe('just plain text')
    })

    test('should clean quoted strings', () => {
      const input = '"hello world"'
      const result = detectValueFormat(input)

      expect(result.displayValue).toBe('hello world')
    })
  })

  describe('isOutputAfterUnknown', () => {
    test('should return true for non-empty objects', () => {
      expect(isOutputAfterUnknown({ key: 'value' })).toBe(true)
      expect(isOutputAfterUnknown({ nested: { data: true } })).toBe(true)
    })

    test('should return false for empty objects', () => {
      expect(isOutputAfterUnknown({})).toBe(false)
    })

    test('should return false for non-objects', () => {
      expect(isOutputAfterUnknown(null)).toBe(false)
      expect(isOutputAfterUnknown(undefined)).toBe(false)
      expect(isOutputAfterUnknown('string')).toBe(false)
      expect(isOutputAfterUnknown(123)).toBe(false)
      expect(isOutputAfterUnknown([])).toBe(false)
    })
  })

  describe('parseTerraformPlan', () => {
    test('should parse a complete Terraform plan with resources and outputs', () => {
      const mockPlan: TTerraformPlan = {
        resource_changes: [
          {
            address: 'aws_instance.example',
            module_address: 'module.compute',
            type: 'aws_instance',
            name: 'example',
            change: {
              actions: ['create'],
              before: null,
              after: { instance_type: 't3.micro' },
              after_unknown: { id: true },
            },
          },
          {
            address: 'aws_s3_bucket.data',
            type: 'aws_s3_bucket',
            name: 'data',
            change: {
              actions: ['update'],
              before: { versioning: false },
              after: { versioning: true },
            },
          },
          {
            address: 'aws_security_group.old',
            type: 'aws_security_group',
            name: 'old',
            change: {
              actions: ['delete'],
              before: { name: 'old-sg' },
              after: null,
            },
          },
        ],
        output_changes: {
          instance_ip: {
            actions: ['create'],
            before: null,
            after: null,
            after_unknown: { ip: true },
          },
          bucket_name: {
            actions: ['update'],
            before: 'old-bucket',
            after: 'new-bucket',
          },
        },
      }

      const result = parseTerraformPlan(mockPlan)

      // Check structure
      expect(result).toHaveProperty('resources')
      expect(result).toHaveProperty('outputs')
      expect(result.resources).toHaveProperty('summary')
      expect(result.resources).toHaveProperty('changes')
      expect(result.outputs).toHaveProperty('summary')
      expect(result.outputs).toHaveProperty('changes')

      // Check resource summary
      expect(result.resources.summary.create).toBe(1)
      expect(result.resources.summary.update).toBe(1)
      expect(result.resources.summary.delete).toBe(1)

      // Check resource changes
      expect(result.resources.changes).toHaveLength(3)

      expect(result.resources.changes[0]).toEqual({
        address: 'aws_instance.example',
        module: 'module.compute',
        resource: 'aws_instance',
        name: 'example',
        action: 'create',
        before: null,
        after: { instance_type: 't3.micro', id: 'Known after apply' },
      })

      // Check output summary
      expect(result.outputs.summary.create).toBe(1)
      expect(result.outputs.summary.update).toBe(1)

      // Check output changes
      expect(result.outputs.changes).toHaveLength(2)

      expect(result.outputs.changes[0]).toMatchObject({
        output: 'instance_ip',
        action: 'create',
        before: null,
        after: { ip: 'Known after apply' },
      })
    })

    test('should handle read-only operations', () => {
      const mockPlan: TTerraformPlan = {
        resource_changes: [
          {
            address: 'data.aws_ami.example',
            type: 'aws_ami',
            name: 'example',
            change: {
              actions: ['read'],
              before: null,
              after: { id: 'ami-12345' },
            },
          },
        ],
      }

      const result = parseTerraformPlan(mockPlan)

      expect(result.resources.summary.read).toBe(1)
      expect(result.resources.changes[0].action).toBe('read')
    })

    test('should handle replace operations', () => {
      const mockPlan: TTerraformPlan = {
        resource_changes: [
          {
            address: 'aws_instance.example',
            type: 'aws_instance',
            name: 'example',
            change: {
              actions: ['replace'],
              before: { instance_type: 't2.micro' },
              after: { instance_type: 't3.micro' },
            },
          },
        ],
      }

      const result = parseTerraformPlan(mockPlan)

      expect(result.resources.summary.replace).toBe(1)
      expect(result.resources.summary.delete).toBe(1) // Replace counts as delete + create
      expect(result.resources.summary.create).toBe(1)
      expect(result.resources.changes[0].action).toBe('replace')
    })

    test('should handle empty plan', () => {
      const mockPlan: TTerraformPlan = {
        resource_changes: [],
      }

      const result = parseTerraformPlan(mockPlan)

      expect(result.resources.changes).toEqual([])
      expect(result.outputs.changes).toEqual([])

      // All operation counts should be 0
      expect(result.resources.summary.create).toBe(0)
      expect(result.resources.summary.update).toBe(0)
      expect(result.resources.summary.delete).toBe(0)
      expect(result.resources.summary.replace).toBe(0)
      expect(result.resources.summary.read).toBe(0)
    })

    test('should merge after_unknown values correctly', () => {
      const mockPlan: TTerraformPlan = {
        resource_changes: [
          {
            address: 'aws_instance.example',
            type: 'aws_instance',
            name: 'example',
            change: {
              actions: ['create'],
              before: null,
              after: { instance_type: 't3.micro', tags: { Name: 'test' } },
              after_unknown: {
                id: true,
                tags: { Environment: true },
                nested: { deep: { value: true } },
              },
            },
          },
        ],
      }

      const result = parseTerraformPlan(mockPlan)
      const change = result.resources.changes[0]

      expect(change.after).toEqual({
        instance_type: 't3.micro',
        tags: {
          Name: 'test',
          Environment: 'Known after apply',
        },
        id: 'Known after apply',
        nested: {
          deep: {
            value: 'Known after apply',
          },
        },
      })
    })

    test('should handle multiple actions on single resource', () => {
      const mockPlan: TTerraformPlan = {
        resource_changes: [
          {
            address: 'aws_instance.example',
            type: 'aws_instance',
            name: 'example',
            change: {
              actions: ['delete', 'create'], // Replace scenario with multiple actions
              before: { instance_type: 't2.micro' },
              after: { instance_type: 't3.micro' },
            },
          },
        ],
      }

      const result = parseTerraformPlan(mockPlan)

      expect(result.resources.changes).toHaveLength(2) // One for each action
      expect(result.resources.summary.delete).toBe(1)
      expect(result.resources.summary.create).toBe(1)
    })

    test('should handle missing output_changes', () => {
      const mockPlan: TTerraformPlan = {
        resource_changes: [
          {
            address: 'aws_instance.example',
            type: 'aws_instance',
            name: 'example',
            change: {
              actions: ['create'],
              after: { instance_type: 't3.micro' },
            },
          },
        ],
        // No output_changes property
      }

      const result = parseTerraformPlan(mockPlan)

      expect(result.outputs.changes).toEqual([])
      expect(result.outputs.summary.create).toBe(0)
    })

    test('should handle null module_address', () => {
      const mockPlan: TTerraformPlan = {
        resource_changes: [
          {
            address: 'aws_instance.example',
            module_address: null,
            type: 'aws_instance',
            name: 'example',
            change: {
              actions: ['create'],
              after: { instance_type: 't3.micro' },
            },
          },
        ],
      }

      const result = parseTerraformPlan(mockPlan)

      expect(result.resources.changes[0].module).toBeNull()
    })
  })
})
