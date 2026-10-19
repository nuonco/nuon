import { describe, expect, test } from 'bun:test'
import type { TAction, TComponent } from '@/types'
import { buildDeploymentPlan } from './deployment-plan'

describe('buildDeploymentPlan', () => {
  test('places component and all-component actions around each deploy', () => {
    const components = [
      { id: 'cmp-api', name: 'api' },
      { id: 'cmp-worker', name: 'worker' },
    ] as TComponent[]
    const actions = [
      {
        id: 'act-migrate',
        name: 'migrate',
        configs: [
          {
            triggers: [
              {
                type: 'pre-deploy-component',
                component_id: 'cmp-api',
              },
            ],
          },
        ],
      },
      {
        id: 'act-notify',
        name: 'notify',
        configs: [
          {
            triggers: [{ type: 'post-deploy-all-components' }],
          },
        ],
      },
    ] as TAction[]

    const plan = buildDeploymentPlan(components, actions)

    expect(plan[0].before.map((action) => action.name)).toEqual(['migrate'])
    expect(plan[1].before).toEqual([])
    expect(plan[0].after.map((action) => action.name)).toEqual(['notify'])
    expect(plan[1].after.map((action) => action.name)).toEqual(['notify'])
  })
})
