import { createContext } from 'react'

export const WorkflowInstallContext = createContext<string | undefined>(
  undefined
)

export const workflowPageSuffix = (
  workflowId: string,
  nested: boolean,
  stepId?: string
) => {
  const path = nested
    ? `/workflows/${workflowId}/workflow`
    : `/workflows/${workflowId}`
  if (!stepId) return path
  return `${path}?panel=${encodeURIComponent(stepId)}`
}

export const isWorkflowPath = (pathname: string, workflowId?: string) =>
  !!workflowId &&
  (pathname.includes(`/workflows/${workflowId}`) ||
    pathname.includes(`/deployments/${workflowId}`))
