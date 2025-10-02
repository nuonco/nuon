import { describe, expect, test } from 'vitest'
import { parseHelmPlan } from './helm-utils'
import type { THelmPlan } from '@/types'

describe('helm-utils', () => {
  describe('parseHelmPlan', () => {
    test('should parse a valid Helm plan with changes and summary', () => {
      const mockPlan: THelmPlan = {
        plan: `
default, myapp-release, ConfigMap (v1) to be created
default, myapp-release, Deployment (apps/v1) to be modified
default, myapp-release, Service (v1) to be destroyed

Plan: 1 to add, 1 to change, 1 to destroy
        `,
        helm_content_diff: [
          {
            kind: 'ConfigMap',
            name: 'myapp-release',
            namespace: 'default',
            before: null,
            after: { data: { config: 'value' } },
          },
          {
            kind: 'Deployment',
            name: 'myapp-release',
            namespace: 'default',
            before: { replicas: 2 },
            after: { replicas: 3 },
          },
          {
            kind: 'Service',
            name: 'myapp-release',
            namespace: 'default',
            before: { port: 80 },
            after: null,
          },
        ],
      } as THelmPlan

      const result = parseHelmPlan(mockPlan)

      expect(result).toHaveProperty('changes')
      expect(result).toHaveProperty('summary')
      expect(Array.isArray(result.changes)).toBe(true)
      expect(result.changes).toHaveLength(3)

      // Check summary
      expect(result.summary).toEqual({
        add: 1,
        change: 1,
        destroy: 1,
      })

      // Check changes structure
      expect(result.changes[0]).toEqual({
        workspace: 'default',
        release: 'myapp-release',
        resource: 'ConfigMap',
        resourceType: 'v1',
        action: 'created',
        before: null,
        after: { data: { config: 'value' } },
      })

      expect(result.changes[1]).toEqual({
        workspace: 'default',
        release: 'myapp-release',
        resource: 'Deployment',
        resourceType: 'apps/v1',
        action: 'modified',
        before: { replicas: 2 },
        after: { replicas: 3 },
      })

      expect(result.changes[2]).toEqual({
        workspace: 'default',
        release: 'myapp-release',
        resource: 'Service',
        resourceType: 'v1',
        action: 'destroyed',
        before: { port: 80 },
        after: null,
      })
    })

    test('should handle empty plan', () => {
      const mockPlan: THelmPlan = {
        plan: '',
        helm_content_diff: [],
      } as THelmPlan

      const result = parseHelmPlan(mockPlan)

      expect(result.changes).toEqual([])
      expect(result.summary).toEqual({
        add: 0,
        change: 0,
        destroy: 0,
      })
    })

    test('should parse plan without diffs', () => {
      const mockPlan: THelmPlan = {
        plan: `
default, myapp-release, ConfigMap (v1) to be created
Plan: 1 to add, 0 to change, 0 to destroy
        `,
        helm_content_diff: null,
      } as THelmPlan

      const result = parseHelmPlan(mockPlan)

      expect(result.changes).toHaveLength(1)
      expect(result.changes[0]).toEqual({
        workspace: 'default',
        release: 'myapp-release',
        resource: 'ConfigMap',
        resourceType: 'v1',
        action: 'created',
        before: undefined,
        after: undefined,
      })

      expect(result.summary).toEqual({
        add: 1,
        change: 0,
        destroy: 0,
      })
    })

    test('should handle plan with no matching diffs', () => {
      const mockPlan: THelmPlan = {
        plan: `
default, myapp-release, ConfigMap (v1) to be created
Plan: 1 to add, 0 to change, 0 to destroy
        `,
        helm_content_diff: [
          {
            kind: 'Deployment', // Different kind
            name: 'myapp-release',
            namespace: 'default',
            before: null,
            after: { replicas: 1 },
          },
        ],
      } as THelmPlan

      const result = parseHelmPlan(mockPlan)

      expect(result.changes).toHaveLength(1)
      expect(result.changes[0]).toEqual({
        workspace: 'default',
        release: 'myapp-release',
        resource: 'ConfigMap',
        resourceType: 'v1',
        action: 'created',
        before: undefined, // No matching diff found
        after: undefined,
      })
    })

    test('should handle ANSI color codes in plan text', () => {
      const mockPlan: THelmPlan = {
        plan: `
\u001b[32mdefault, myapp-release, ConfigMap (v1) to be created\u001b[0m
\u001b[33mPlan: 1 to add, 0 to change, 0 to destroy\u001b[0m
        `,
        helm_content_diff: [],
      } as THelmPlan

      const result = parseHelmPlan(mockPlan)

      expect(result.changes).toHaveLength(1)
      expect(result.changes[0].workspace).toBe('default')
      expect(result.summary.add).toBe(1)
    })

    test('should handle multiple namespaces and releases', () => {
      const mockPlan: THelmPlan = {
        plan: `
default, app1-release, ConfigMap (v1) to be created
kube-system, app2-release, Secret (v1) to be modified
production, app3-release, Deployment (apps/v1) to be destroyed

Plan: 1 to add, 1 to change, 1 to destroy
        `,
        helm_content_diff: [],
      } as THelmPlan

      const result = parseHelmPlan(mockPlan)

      expect(result.changes).toHaveLength(3)

      expect(result.changes[0].workspace).toBe('default')
      expect(result.changes[0].release).toBe('app1-release')
      expect(result.changes[0].action).toBe('created')

      expect(result.changes[1].workspace).toBe('kube-system')
      expect(result.changes[1].release).toBe('app2-release')
      expect(result.changes[1].action).toBe('modified')

      expect(result.changes[2].workspace).toBe('production')
      expect(result.changes[2].release).toBe('app3-release')
      expect(result.changes[2].action).toBe('destroyed')
    })

    test('should handle plan with only summary line', () => {
      const mockPlan: THelmPlan = {
        plan: 'Plan: 5 to add, 3 to change, 2 to destroy',
        helm_content_diff: [],
      } as THelmPlan

      const result = parseHelmPlan(mockPlan)

      expect(result.changes).toEqual([])
      expect(result.summary).toEqual({
        add: 5,
        change: 3,
        destroy: 2,
      })
    })

    test('should handle plan with changes but no summary', () => {
      const mockPlan: THelmPlan = {
        plan: 'default, myapp-release, ConfigMap (v1) to be created',
        helm_content_diff: [],
      } as THelmPlan

      const result = parseHelmPlan(mockPlan)

      expect(result.changes).toHaveLength(1)
      expect(result.summary).toEqual({
        add: 0,
        change: 0,
        destroy: 0,
      })
    })

    test('should handle malformed lines gracefully', () => {
      const mockPlan: THelmPlan = {
        plan: `
This is not a valid helm plan line
default, myapp-release, ConfigMap (v1) to be created
Another invalid line
Plan: 1 to add, 0 to change, 0 to destroy
        `,
        helm_content_diff: [],
      } as THelmPlan

      const result = parseHelmPlan(mockPlan)

      // Should only parse the valid line
      expect(result.changes).toHaveLength(1)
      expect(result.changes[0].resource).toBe('ConfigMap')
      expect(result.summary.add).toBe(1)
    })
  })
})
