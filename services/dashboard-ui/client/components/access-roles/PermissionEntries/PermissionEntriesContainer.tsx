import { useMemo } from 'react'
import { useQueries, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getAppBranches, getApps, getInstalls } from '@/lib'
import type { TPermissionEntry } from '@/types'
import { PermissionEntries } from './PermissionEntries'

const RESOURCE_LIMIT = 100

export const PermissionEntriesContainer = ({
  value,
  onChange,
  disabled,
}: {
  value: TPermissionEntry[]
  onChange: (next: TPermissionEntry[]) => void
  disabled?: boolean
}) => {
  const { org } = useOrg()

  const { data: apps } = useQuery({
    queryKey: ['permission-entry-apps', org?.id],
    queryFn: () => getApps({ orgId: org!.id, limit: RESOURCE_LIMIT }),
    enabled: !!org?.id,
    staleTime: 5 * 60 * 1000,
  })

  const { data: installs } = useQuery({
    queryKey: ['permission-entry-installs', org?.id],
    queryFn: () => getInstalls({ orgId: org!.id, limit: RESOURCE_LIMIT }),
    enabled: !!org?.id,
    staleTime: 5 * 60 * 1000,
  })

  const appList = useMemo(
    () => (apps?.data ?? []).filter((app) => !!app?.id),
    [apps?.data]
  )

  // Branches are only listable per app, so a flat list costs one request per
  // app. Hold off until an entry actually targets branches.
  const needsBranches = value.some((e) => e.resource_type === 'app_branch')

  const branchQueries = useQueries({
    queries: appList.map((app) => ({
      queryKey: ['permission-entry-branches', org?.id, app.id],
      queryFn: () =>
        getAppBranches({
          orgId: org!.id,
          appId: app.id!,
          limit: RESOURCE_LIMIT,
        }),
      enabled: !!org?.id && needsBranches,
      staleTime: 5 * 60 * 1000,
    })),
  })

  const branchOptions = useMemo(
    () =>
      appList.flatMap((app, index) =>
        (branchQueries[index]?.data?.data ?? [])
          .filter((branch) => !!branch?.id)
          // Branch names are only unique within an app, so the app name has to
          // ride along to tell two "main" branches apart.
          .map((branch) => ({
            value: branch.id!,
            label: `${app.name || app.id} / ${branch.name || branch.id}`,
          }))
      ),
    [appList, branchQueries]
  )

  return (
    <PermissionEntries
      value={value}
      onChange={onChange}
      appOptions={toOptions(appList)}
      installOptions={toOptions(installs?.data)}
      branchOptions={branchOptions}
      branchesLoading={
        needsBranches && branchQueries.some((query) => query.isLoading)
      }
      orgId={org?.id ?? ''}
      disabled={disabled}
    />
  )
}

function toOptions(items: { id?: string; name?: string }[] | undefined) {
  return (items ?? [])
    .filter((item) => !!item?.id)
    .map((item) => ({ value: item.id!, label: item.name || item.id! }))
}
