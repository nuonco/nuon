import { useState } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import {
  getActions,
  getAppInstalls,
  getComponents,
  getInstallState,
  getRunbooks,
} from '@/lib'
import { ReadmeStudio } from './ReadmeStudio'

const toOptions = (items?: ({ id?: string; name?: string } | undefined)[]) =>
  (items ?? []).flatMap((item) =>
    item?.id && item?.name ? [{ id: item.id, name: item.name }] : []
  )

export function ReadmeStudioContainer() {
  const { org } = useOrg()
  const { app } = useApp()
  const enabled = !!org?.id && !!app?.id
  const [previewInstallId, setPreviewInstallId] = useState('')
  const installs = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['installs', org?.id, app?.id, 'readme-studio'],
    queryFn: () =>
      getAppInstalls({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled,
  })
  const runbooks = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['runbooks', org?.id, app?.id, 'readme-studio'],
    queryFn: () => getRunbooks({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled,
  })
  const actions = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['actions', org?.id, app?.id, 'readme-studio'],
    queryFn: () => getActions({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled,
  })
  const components = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['components', org?.id, app?.id, 'readme-studio'],
    queryFn: () =>
      getComponents({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled,
  })
  const installState = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-state', org?.id, previewInstallId, 'readme-studio'],
    queryFn: () =>
      getInstallState({ orgId: org!.id, installId: previewInstallId }),
    enabled: !!org?.id && !!previewInstallId,
  })
  return (
    <ReadmeStudio
      appId={app?.id}
      installs={toOptions(installs.data?.data)}
      runbooks={toOptions(runbooks.data?.data)}
      actions={toOptions(actions.data?.data)}
      components={toOptions(components.data?.data)}
      previewInstallId={previewInstallId}
      previewInstallState={installState.data}
      previewInstallStateLoading={installState.isLoading && !!previewInstallId}
      onPreviewInstallChange={setPreviewInstallId}
      loadingError={!!installs.error || !!installState.error}
    />
  )
}
