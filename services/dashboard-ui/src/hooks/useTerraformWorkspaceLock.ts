import { usePolling } from '@/hooks/use-polling'
import { useOrg } from '@/hooks/use-org'
import type { TTerraformWorkspaceLock } from '@/types/ctl-api.types'

export function useTerraformWorkspaceLock(workspaceId?: string) {
  const { org } = useOrg()

  const { data: lockStatus, error, isLoading } = usePolling<TTerraformWorkspaceLock | null>({
    path: workspaceId && org?.id
      ? `/api/orgs/${org.id}/terraform-workspaces/${workspaceId}/lock`
      : null,
    shouldPoll: !!workspaceId && !!org?.id,
    pollInterval: 5000, // 5 seconds
    initData: null,
    backoff: {
      enabled: true,
      initialDelay: 1000,
      maxDelay: 30000,
    },
  })

  return {
    lockStatus,
    isLocked: !!lockStatus,
    isLoading,
    error,
  }
}
