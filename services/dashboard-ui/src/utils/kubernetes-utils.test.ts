import { describe, expect, test } from 'vitest'
import { parseKubernetesPlan } from './kubernetes-utils'
import type { TKubernetesPlan } from '@/types'

describe('kubernetes-utils', () => {
  describe('parseKubernetesPlan', () => {
    test('should parse a valid Kubernetes plan with different operations', () => {
      const mockPlan: TKubernetesPlan = [
        {
          op: 'apply',
          namespace: 'default',
          name: 'myapp-deployment',
          group_version_kind: {
            Kind: 'Deployment',
            Version: 'apps/v1',
            Group: 'apps',
          },
          before: null,
          after: { spec: { replicas: 3 } },
        },
        {
          op: 'apply',
          namespace: 'default',
          name: 'myapp-service',
          group_version_kind: {
            Kind: 'Service',
            Version: 'v1',
            Group: '',
          },
          before: { spec: { ports: [{ port: 8080 }] } },
          after: { spec: { ports: [{ port: 9090 }] } },
        },
        {
          op: 'delete',
          namespace: 'default',
          name: 'myapp-configmap',
          group_version_kind: {
            Kind: 'ConfigMap',
            Version: 'v1',
            Group: '',
          },
          before: { data: { config: 'old' } },
          after: null,
        },
      ]

      const result = parseKubernetesPlan(mockPlan)

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

      // Check first change (apply with no before = added)
      expect(result.changes[0]).toEqual({
        namespace: 'default',
        name: 'myapp-deployment',
        resource: 'Deployment',
        resourceType: 'apps/v1',
        action: 'added',
        before: null,
        after: { spec: { replicas: 3 } },
      })

      // Check second change (apply with before and after = changed)
      expect(result.changes[1]).toEqual({
        namespace: 'default',
        name: 'myapp-service',
        resource: 'Service',
        resourceType: 'v1',
        action: 'changed',
        before: { spec: { ports: [{ port: 8080 }] } },
        after: { spec: { ports: [{ port: 9090 }] } },
      })

      // Check third change (delete = destroyed)
      expect(result.changes[2]).toEqual({
        namespace: 'default',
        name: 'myapp-configmap',
        resource: 'ConfigMap',
        resourceType: 'v1',
        action: 'destroyed',
        before: { data: { config: 'old' } },
        after: null,
      })
    })

    test('should handle empty plan', () => {
      const mockPlan: TKubernetesPlan = []

      const result = parseKubernetesPlan(mockPlan)

      expect(result.changes).toEqual([])
      expect(result.summary).toEqual({
        add: 0,
        change: 0,
        destroy: 0,
      })
    })

    test('should handle apply operation with before but no after (destroyed)', () => {
      const mockPlan: TKubernetesPlan = [
        {
          op: 'apply',
          namespace: 'default',
          name: 'myapp-pod',
          group_version_kind: {
            Kind: 'Pod',
            Version: 'v1',
            Group: '',
          },
          before: { spec: { containers: [] } },
          after: null,
        },
      ]

      const result = parseKubernetesPlan(mockPlan)

      expect(result.changes).toHaveLength(1)
      expect(result.changes[0].action).toBe('destroyed')
      expect(result.summary).toEqual({
        add: 0,
        change: 0,
        destroy: 1,
      })
    })

    test('should handle delete operation', () => {
      const mockPlan: TKubernetesPlan = [
        {
          op: 'delete',
          namespace: 'kube-system',
          name: 'system-pod',
          group_version_kind: {
            Kind: 'Pod',
            Version: 'v1',
            Group: '',
          },
          before: { spec: { containers: [] } },
          after: null,
        },
      ]

      const result = parseKubernetesPlan(mockPlan)

      expect(result.changes).toHaveLength(1)
      expect(result.changes[0]).toEqual({
        namespace: 'kube-system',
        name: 'system-pod',
        resource: 'Pod',
        resourceType: 'v1',
        action: 'destroyed',
        before: { spec: { containers: [] } },
        after: null,
      })
      expect(result.summary.destroy).toBe(1)
    })

    test('should handle unknown operation as fallback', () => {
      const mockPlan: TKubernetesPlan = [
        {
          op: 'unknown_op' as any,
          namespace: 'default',
          name: 'test-resource',
          group_version_kind: {
            Kind: 'CustomResource',
            Version: 'v1alpha1',
            Group: 'custom.io',
          },
          before: null,
          after: { spec: {} },
        },
      ]

      const result = parseKubernetesPlan(mockPlan)

      expect(result.changes).toHaveLength(1)
      expect(result.changes[0].action).toBe('unknown_op')
      expect(result.summary).toEqual({
        add: 0,
        change: 0,
        destroy: 0,
      })
    })

    test('should count different operation types correctly', () => {
      const mockPlan: TKubernetesPlan = [
        // 2 adds
        {
          op: 'apply',
          namespace: 'default',
          name: 'pod1',
          group_version_kind: { Kind: 'Pod', Version: 'v1', Group: '' },
          before: null,
          after: { spec: {} },
        },
        {
          op: 'apply',
          namespace: 'default',
          name: 'pod2',
          group_version_kind: { Kind: 'Pod', Version: 'v1', Group: '' },
          before: null,
          after: { spec: {} },
        },
        // 1 change
        {
          op: 'apply',
          namespace: 'default',
          name: 'svc1',
          group_version_kind: { Kind: 'Service', Version: 'v1', Group: '' },
          before: { spec: { port: 80 } },
          after: { spec: { port: 8080 } },
        },
        // 1 destroy (delete op)
        {
          op: 'delete',
          namespace: 'default',
          name: 'cm1',
          group_version_kind: { Kind: 'ConfigMap', Version: 'v1', Group: '' },
          before: { data: {} },
          after: null,
        },
      ]

      const result = parseKubernetesPlan(mockPlan)

      expect(result.summary.add).toBe(2)
      expect(result.summary.change).toBe(1)
      expect(result.summary.destroy).toBe(1)
      expect(result.changes).toHaveLength(4)
    })

    test('should handle resources with different group versions', () => {
      const mockPlan: TKubernetesPlan = [
        {
          op: 'apply',
          namespace: 'default',
          name: 'my-deployment',
          group_version_kind: {
            Kind: 'Deployment',
            Version: 'apps/v1',
            Group: 'apps',
          },
          before: null,
          after: { spec: { replicas: 1 } },
        },
        {
          op: 'apply',
          namespace: 'default',
          name: 'my-crd',
          group_version_kind: {
            Kind: 'CustomResource',
            Version: 'v1beta1',
            Group: 'example.com',
          },
          before: null,
          after: { spec: { custom: 'value' } },
        },
      ]

      const result = parseKubernetesPlan(mockPlan)

      expect(result.changes[0].resource).toBe('Deployment')
      expect(result.changes[0].resourceType).toBe('apps/v1')

      expect(result.changes[1].resource).toBe('CustomResource')
      expect(result.changes[1].resourceType).toBe('v1beta1')
    })

    test('should handle apply with neither before nor after', () => {
      const mockPlan: TKubernetesPlan = [
        {
          op: 'apply',
          namespace: 'default',
          name: 'test-resource',
          group_version_kind: {
            Kind: 'Pod',
            Version: 'v1',
            Group: '',
          },
          before: null,
          after: null,
        },
      ]

      const result = parseKubernetesPlan(mockPlan)

      expect(result.changes).toHaveLength(1)
      expect(result.changes[0].action).toBe('apply') // Falls back to op
      expect(result.summary).toEqual({
        add: 0,
        change: 0,
        destroy: 0,
      })
    })

    test('should preserve all resource metadata', () => {
      const mockPlan: TKubernetesPlan = [
        {
          op: 'apply',
          namespace: 'production',
          name: 'webapp-deployment',
          group_version_kind: {
            Kind: 'Deployment',
            Version: 'apps/v1',
            Group: 'apps',
          },
          before: { metadata: { labels: { env: 'staging' } } },
          after: { metadata: { labels: { env: 'production' } } },
        },
      ]

      const result = parseKubernetesPlan(mockPlan)

      expect(result.changes[0]).toEqual({
        namespace: 'production',
        name: 'webapp-deployment',
        resource: 'Deployment',
        resourceType: 'apps/v1',
        action: 'changed',
        before: { metadata: { labels: { env: 'staging' } } },
        after: { metadata: { labels: { env: 'production' } } },
      })
    })
  })
})
