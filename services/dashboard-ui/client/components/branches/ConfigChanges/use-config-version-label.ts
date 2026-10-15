import { useQuery } from '@tanstack/react-query'
import { getAppConfig } from '@/lib'

export const useConfigVersionLabel = ({
  orgId,
  appId,
  headConfigId,
  baseConfigId,
}: {
  orgId?: string
  appId?: string
  headConfigId?: string
  baseConfigId?: string
}) => {
  const versionQuery = (configId?: string) => ({
    queryKey: ['app-config-version', orgId, appId, configId],
    queryFn: () =>
      getAppConfig({
        orgId: orgId!,
        appId: appId!,
        appConfigId: configId!,
      }),
    enabled: !!orgId && !!appId && !!configId,
    select: (config: { version?: number }) => config.version,
  })
  const { data: nextVersion } = useQuery(versionQuery(headConfigId))
  const { data: previousVersion } = useQuery(versionQuery(baseConfigId))

  if (nextVersion == null) return undefined
  return previousVersion == null || previousVersion === nextVersion
    ? `v${nextVersion}`
    : `v${previousVersion} → v${nextVersion}`
}
