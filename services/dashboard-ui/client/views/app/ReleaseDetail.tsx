import { useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { ReleaseFilesContainer } from '@/components/apps/bundles/ReleaseFiles'
import { Hash } from '@/components/common/Hash'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppRelease, getPreviousAppRelease } from '@/lib'

export const ReleaseDetail = () => {
  const { releaseId } = useParams()
  const { app } = useApp()
  const { org } = useOrg()
  const {
    data: release,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['app-release', org?.id, app?.id, releaseId],
    queryFn: () =>
      getAppRelease({
        appId: app!.id,
        orgId: org!.id,
        releaseId: releaseId!,
      }),
    enabled: !!org?.id && !!app?.id && !!releaseId,
    placeholderData: keepPreviousData,
  })
  const { data: previousRelease } = useQuery({
    queryKey: ['previous-app-release', org?.id, app?.id, releaseId],
    queryFn: () =>
      getPreviousAppRelease({
        appId: app!.id,
        orgId: org!.id,
        releaseId: releaseId!,
      }),
    enabled: !!org?.id && !!app?.id && !!releaseId,
    placeholderData: keepPreviousData,
  })

  return (
    <>
      <PageTitle segments={[release?.id ?? 'Release', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/releases`, text: 'Releases' },
          {
            path: `/${org?.id}/apps/${app?.id}/releases/${releaseId}`,
            text: release?.id ?? 'Release',
          },
        ]}
      />
      <DetailPage
        header={
          <DetailHeader
            title="Release"
            id={release?.id ?? releaseId}
            loading={isLoading}
            status={<Status status={release?.status} loading={isLoading} />}
            description={release?.status_description}
            metadata={
              <>
                <LabeledValue label="App config" loading={isLoading}>
                  {release?.app_config_id}
                </LabeledValue>
                <LabeledValue label="Created" loading={isLoading}>
                  {release?.created_at ? (
                    <Time time={release.created_at} format="relative" />
                  ) : null}
                </LabeledValue>
                <LabeledValue label="Digest" loading={isLoading}>
                  <Hash hash={release?.semantic_digest ?? ''} />
                </LabeledValue>
                <LabeledValue label="Files" loading={isLoading}>
                  {release?.source_files?.length.toString()}
                </LabeledValue>
              </>
            }
          />
        }
      >
        {error ? (
          <Text theme="error">
            Release failed to load. Try refreshing the page.
          </Text>
        ) : release && org?.id && app?.id ? (
          <ReleaseFilesContainer
            appId={app.id}
            orgId={org.id}
            previousRelease={previousRelease}
            release={release}
          />
        ) : null}
      </DetailPage>
    </>
  )
}
