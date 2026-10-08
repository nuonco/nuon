import { api } from '@/lib/api'
import type { TInstallDeploymentRecord } from '@/types'

export const getInstallDeployment = ({
  installId,
  orgId,
  workflowId,
}: {
  installId: string
  orgId: string
  workflowId: string
}) =>
  api<TInstallDeploymentRecord>({
    path: `installs/${installId}/deployments/${workflowId}`,
    orgId,
  })
