import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import {
  getActions,
  getAppInstalls,
  getComponents,
  getInstallState,
  getRunbooks,
} from '@/lib'
import { RunbookNotebook } from './RunbookNotebook'

const toOptions = (items?: ({ id?: string; name?: string } | undefined)[]) =>
  (items ?? []).flatMap((item) =>
    item?.id && item?.name ? [{ id: item.id, name: item.name }] : []
  )

export function OperationsStudioContainer() {
  const { org } = useOrg()
  const { app } = useApp()
  const enabled = !!org?.id && !!app?.id
  const [previewInstallId, setPreviewInstallId] = useState('')
  const installs = useQuery({
    queryKey: ['installs', org?.id, app?.id, 'operations-studio'],
    queryFn: () =>
      getAppInstalls({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled,
  })
  const runbooks = useQuery({
    queryKey: ['runbooks', org?.id, app?.id, 'operations-studio'],
    queryFn: () => getRunbooks({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled,
  })
  const actions = useQuery({
    queryKey: ['actions', org?.id, app?.id, 'operations-studio'],
    queryFn: () => getActions({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled,
  })
  const components = useQuery({
    queryKey: ['components', org?.id, app?.id, 'operations-studio'],
    queryFn: () =>
      getComponents({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled,
  })
  const installState = useQuery({
    queryKey: ['install-state', org?.id, previewInstallId, 'operations-studio'],
    queryFn: () =>
      getInstallState({ orgId: org!.id, installId: previewInstallId }),
    enabled: !!org?.id && !!previewInstallId,
  })
  return (
    <RunbookNotebook
      appId={app?.id}
      installs={toOptions(installs.data?.data)}
      runbooks={(runbooks.data?.data ?? []).filter(
        (item) => !!item?.id && !!item?.name
      )}
      actions={toOptions(actions.data?.data)}
      components={toOptions(components.data?.data)}
      previewInstallId={previewInstallId}
      previewInstallState={installState.data}
      previewInstallStateLoading={installState.isLoading && !!previewInstallId}
      onPreviewInstallChange={setPreviewInstallId}
      loading={
        components.isLoading || actions.isLoading || runbooks.isLoading
      }
      loadingError={
        !!installs.error ||
        !!installState.error ||
        !!components.error ||
        !!actions.error ||
        !!runbooks.error
      }
    />
  )
}
