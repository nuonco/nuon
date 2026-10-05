import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useBranchScopedAppConfig } from '@/hooks/use-branch-scoped-app-config'
import { getAppLabels } from '@/lib'

export const useBranchNavCounts = ({
  orgId,
  appId,
  branchId,
}: {
  orgId?: string
  appId?: string
  branchId?: string
}) => {
  const { appConfig, isLoading } = useBranchScopedAppConfig({
    orgId,
    appId,
    branchId,
  })

  const { data: labelsData } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-labels', orgId, appId],
    queryFn: () => getAppLabels({ orgId: orgId!, appId: appId! }),
    enabled: !!orgId && !!appId,
  })

  return {
    inputs: appConfig?.input?.inputs?.length,
    roles: isLoading
      ? undefined
      : (appConfig?.permissions?.aws_iam_roles?.length ?? 0),
    policies: isLoading
      ? undefined
      : (appConfig?.policies?.policies?.length ?? 0),
    labels: labelsData?.labels?.length,
  }
}
