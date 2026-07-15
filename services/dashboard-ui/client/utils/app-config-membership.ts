import type { TAppConfig } from '@/types'

export function isComponentInAppConfig(
  appConfig: TAppConfig | undefined,
  componentId: string | undefined
): boolean {
  if (!appConfig || !componentId) return true
  return !!appConfig.component_config_connections?.some(
    (c) => c.component_id === componentId
  )
}

export function isActionInAppConfig(
  appConfig: TAppConfig | undefined,
  actionWorkflowId: string | undefined
): boolean {
  if (!appConfig || !actionWorkflowId) return true
  return !!appConfig.action_workflow_configs?.some(
    (a) => a.action_workflow_id === actionWorkflowId
  )
}
