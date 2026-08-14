import type { TDeploy } from '@/types'

// Matches the composite error type the API emits for a Helm release left
// mid-operation (services/ctl-api/internal/app/runners/errparse/helm).
export const HELM_PENDING_OPERATION_ERROR_TYPE = 'helm.pending_operation'

/**
 * stuckHelmReleaseStatus returns the pending Helm status when a deploy failed
 * because the release was left mid-operation, and undefined otherwise.
 *
 * It keys off the deploy's composite error rather than a stored release status:
 * the composite error is proof the runner actually observed the stuck release,
 * and it is replaced whenever a newer deploy runs, so it cannot keep claiming a
 * release is stuck after it has been recovered.
 */
export const stuckHelmReleaseStatus = (deploy?: TDeploy): string | undefined => {
  const compositeError = deploy?.composite_error
  if (compositeError?.type !== HELM_PENDING_OPERATION_ERROR_TYPE) {
    return undefined
  }

  const status = (compositeError?.data as { status?: string } | undefined)?.status
  return status || 'a pending state'
}
