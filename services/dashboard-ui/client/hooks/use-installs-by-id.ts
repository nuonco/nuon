import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getAppInstalls } from '@/lib'
import type { TInstall } from '@/types'

export function useInstallsById(appId?: string): Record<string, TInstall> {
  const { org } = useOrg()
  const orgId = org?.id ?? ''

  const { data } = useQuery({
    queryKey: ['app-installs', orgId, appId],
    queryFn: () => getAppInstalls({ appId: appId!, orgId, limit: 100 }),
    enabled: !!orgId && !!appId,
    refetchInterval: 10000,
  })

  return useMemo(
    () =>
      (data?.data ?? []).reduce<Record<string, TInstall>>((acc, install) => {
        acc[install.id] = install
        return acc
      }, {}),
    [data]
  )
}
