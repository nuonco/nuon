export default {
  title: 'Approvals/PlanDiffs/KubernetesDiff',
}

import { KubernetesDiff } from './KubernetesDiff'

const mockPlan = {
  plan: 'k8s-plan-1',
  op: 'apply',
  k8s_content_diff: [
    {
      _version: 'v1',
      name: 'my-configmap',
      namespace: 'default',
      kind: 'ConfigMap',
      api: 'v1',
      resource: 'configmaps',
      op: 'changed',
      type: 3,
      dry_run: false,
      entries: [
        {
          path: 'data.DATABASE_URL',
          original: 'postgres://old-host:5432/db',
          applied: 'postgres://new-host:5432/db',
          type: 2,
          payload: '',
        },
      ],
    },
    {
      _version: 'v1',
      name: 'new-secret',
      namespace: 'staging',
      kind: 'Secret',
      api: 'v1',
      resource: 'secrets',
      op: 'added',
      type: 1,
      dry_run: false,
    },
  ],
} as any

export const Default = () => <KubernetesDiff plan={mockPlan} />
