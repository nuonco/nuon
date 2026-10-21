export const workflowPageSuffix = (workflowId: string, nested: boolean) =>
  nested ? `/workflows/${workflowId}/workflow` : `/workflows/${workflowId}`

export const isWorkflowPath = (pathname: string, workflowId?: string) =>
  !!workflowId &&
  (pathname.includes(`/workflows/${workflowId}`) ||
    pathname.includes(`/deployments/${workflowId}`))
