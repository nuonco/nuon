import type { TDeploy } from '@/types'

export const HELM_PENDING_OPERATION_ERROR_TYPE = 'helm.pending_operation'

export const stuckHelmReleaseStatus = (deploy?: TDeploy): string | undefined => {
  const compositeError = deploy?.composite_error
  if (compositeError?.type !== HELM_PENDING_OPERATION_ERROR_TYPE) {
    return undefined
  }

  const status = (compositeError?.data as { status?: string } | undefined)?.status
  return status || 'a pending state'
}
