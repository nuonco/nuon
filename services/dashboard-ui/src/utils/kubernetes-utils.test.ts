import { describe, expect, test } from 'vitest'
import { parseKubernetesPlan } from './kubernetes-utils'
import type { TKubernetesPlan, TKubernetesPlanItem } from '@/types'

describe('kubernetes-utils', () => {
  describe('parseKubernetesPlan', () => {
    test('should parse a valid Kubernetes plan with different operations', () => {
      const mockPlan: TKubernetesPlan = {
        plan: '',
        op: 'apply',
        k8s_content_diff: [
          {
            _version: '2',
            op: 'apply',
            type: 2, // added
            dry_run: true,
            namespace: 'default',
            name: 'myapp-deployment',
            kind: 'Deployment',
            api: 'apps/v1',
            resource: 'deployments',
            entries: [
                {
                    type: 2,
                    payload: 'spec:\n  replicas: 3',
                    path: '',
                    original: '',
                    applied: ''
                }
            ]
          } as TKubernetesPlanItem,
          {
            _version: '2',
            op: 'apply',
            type: 3, // changed
            dry_run: true,
            namespace: 'default',
            name: 'myapp-service',
            kind: 'Service',
            api: 'v1',
            resource: 'services',
            entries: [
                {
                    type: 1,
                    payload: 'spec:\n  ports:\n  - port: 8080',
                    path: '',
                    original: '',
                    applied: ''
                },
                {
                    type: 2,
                    payload: 'spec:\n  ports:\n  - port: 9090',
                    path: '',
                    original: '',
                    applied: ''
                }
            ]
          } as TKubernetesPlanItem,
          {
            _version: '2',
            op: 'delete', // op delete takes precedence
            type: 1, // destroyed
            dry_run: true,
            namespace: 'default',
            name: 'myapp-configmap',
            kind: 'ConfigMap',
            api: 'v1',
            resource: 'configmaps',
            entries: [
                {
                    type: 1,
                    payload: 'data:\n  config: old',
                    path: '',
                    original: '',
                    applied: ''
                }
            ]
          } as TKubernetesPlanItem,
        ]
      }

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

      // Check first change (apply with type 2 = added)
      expect(result.changes[0]).toEqual({
        namespace: 'default',
        name: 'myapp-deployment',
        resource: 'Deployment',
        resourceType: 'apps/v1',
        action: 'added',
        before: null,
        after: 'spec:\n  replicas: 3',
      })

      // Check second change (apply with type 3 = changed)
      expect(result.changes[1]).toEqual({
        namespace: 'default',
        name: 'myapp-service',
        resource: 'Service',
        resourceType: 'v1',
        action: 'changed',
        before: 'spec:\n  ports:\n  - port: 8080',
        after: 'spec:\n  ports:\n  - port: 9090',
      })

      // Check third change (delete op = destroyed)
      expect(result.changes[2]).toEqual({
        namespace: 'default',
        name: 'myapp-configmap',
        resource: 'ConfigMap',
        resourceType: 'v1',
        action: 'destroyed',
        before: 'data:\n  config: old',
        after: null,
      })
    })

    test('should handle empty plan', () => {
      const mockPlan: TKubernetesPlan = {
        plan: '',
        op: 'apply',
        k8s_content_diff: []
      }

      const result = parseKubernetesPlan(mockPlan)

      expect(result.changes).toEqual([])
      expect(result.summary).toEqual({
        add: 0,
        change: 0,
        destroy: 0,
      })
    })

    test('should handle apply operation with type 1 (destroyed)', () => {
      const mockPlan: TKubernetesPlan = {
        plan: '',
        op: 'apply',
        k8s_content_diff: [
          {
            _version: '2',
            op: 'apply',
            type: 1, // destroyed
            dry_run: true,
            namespace: 'default',
            name: 'myapp-pod',
            kind: 'Pod',
            api: 'v1',
            resource: 'pods',
            entries: [
                {
                    type: 1,
                    payload: 'spec:\n  containers: []',
                    path: '',
                    original: '',
                    applied: ''
                }
            ]
          } as TKubernetesPlanItem
        ]
      }

      const result = parseKubernetesPlan(mockPlan)

      expect(result.changes).toHaveLength(1)
      expect(result.changes[0].action).toBe('destroyed')
      expect(result.summary).toEqual({
        add: 0,
        change: 0,
        destroy: 1,
      })
    })
  })
})
