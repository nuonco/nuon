import { useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getActions, getComponents, getRunbooks } from '@/lib'
import { RunbookBuilder } from './RunbookBuilder'

export function RunbookBuilderContainer() {
  const { org } = useOrg()
  const { app } = useApp()
  const enabled = !!org?.id && !!app?.id
  const components = useQuery({
    queryKey: ['components', org?.id, app?.id, 'runbook-builder'],
    queryFn: () =>
      getComponents({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled,
  })
  const actions = useQuery({
    queryKey: ['actions', org?.id, app?.id, 'runbook-builder'],
    queryFn: () => getActions({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled,
  })
  const runbooks = useQuery({
    queryKey: ['runbooks', org?.id, app?.id, 'runbook-builder'],
    queryFn: () => getRunbooks({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled,
  })
  return (
    <RunbookBuilder
      components={(components.data?.data ?? []).flatMap((item) =>
        item?.id && item?.name ? [{ id: item.id, name: item.name }] : []
      )}
      actions={(actions.data?.data ?? []).flatMap((item) =>
        item?.id && item?.name ? [{ id: item.id, name: item.name }] : []
      )}
      runbooks={(runbooks.data?.data ?? []).filter(
        (item) => !!item?.id && !!item?.name
      )}
      loading={components.isLoading || actions.isLoading || runbooks.isLoading}
      loadingError={!!components.error || !!actions.error || !!runbooks.error}
    />
  )
}
