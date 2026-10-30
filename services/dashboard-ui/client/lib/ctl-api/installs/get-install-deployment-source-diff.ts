import { api } from '@/lib/api'
import type { TInstallDeploymentSourceDiff } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export type TInstallDeploymentSourceDiffTarget = {
  orgId: string
  installId: string
  workflowId: string
  kind: 'component' | 'sandbox'
  name: string
}

export const getInstallDeploymentSourceDiff = ({
  orgId,
  installId,
  workflowId,
  kind,
  name,
  path,
}: TInstallDeploymentSourceDiffTarget & { path?: string }) =>
  api<TInstallDeploymentSourceDiff>({
    path: `installs/${installId}/deployments/${workflowId}/source-diff${buildQueryParams({ kind, name, path })}`,
    orgId,
    abortTimeout: path ? 30000 : undefined,
  })
