import type {
  TAction,
  TActionConfigTrigger,
  TComponent,
} from '@/types'

export interface IDeploymentPlanRow {
  component: TComponent
  before: TAction[]
  after: TAction[]
}

const actionTriggers = (action: TAction): TActionConfigTrigger[] =>
  action.configs?.at(-1)?.triggers ?? []

const actionMatches = (
  action: TAction,
  component: TComponent,
  types: string[]
) =>
  actionTriggers(action).some((trigger) => {
    if (!types.includes(trigger.type ?? '')) return false
    if (trigger.type?.includes('-all-components')) return true
    return trigger.component_id === component.id
  })

export const buildDeploymentPlan = (
  components: TComponent[],
  actions: TAction[]
): IDeploymentPlanRow[] =>
  components.map((component) => ({
    component,
    before: actions.filter((action) =>
      actionMatches(action, component, [
        'pre-deploy-component',
        'pre-deploy-all-components',
      ])
    ),
    after: actions.filter((action) =>
      actionMatches(action, component, [
        'post-deploy-component',
        'post-deploy-all-components',
      ])
    ),
  }))
