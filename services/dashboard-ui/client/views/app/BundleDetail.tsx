import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { BundleContentsTable } from '@/components/apps/bundles/BundleContentsTable'
import { DownloadBundleButton } from '@/components/apps/bundles/DownloadBundle'
import { RegisterAirgapInstallButton } from '@/components/apps/bundles/RegisterAirgapInstall'
import { BackLink } from '@/components/common/BackLink'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAirgapBundle } from '@/lib'
import { formatBytes } from '@/utils/string-utils'

const PENDING_STATUSES = ['queued', 'publishing']

export const BundleDetail = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const params = useParams()
  const bundleId = params.bundleId as string

  const {
    data: bundle,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ['airgap-bundle', org?.id, app?.id, bundleId],
    queryFn: () =>
      getAirgapBundle({ appId: app!.id, bundleId, orgId: org!.id }),
    enabled: !!org?.id && !!app?.id && !!bundleId,
    refetchInterval: (query) =>
      PENDING_STATUSES.includes(query.state.data?.status ?? '') ? 5000 : false,
  })

  const breadcrumbs = (
    <Breadcrumbs
      breadcrumbs={[
        { path: `/${org?.id}`, text: org?.name },
        { path: `/${org?.id}/apps`, text: 'Apps' },
        { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
        { path: `/${org?.id}/apps/${app?.id}/bundles`, text: 'Bundles' },
        {
          path: `/${org?.id}/apps/${app?.id}/bundles/${bundleId}`,
          text: 'Bundle',
        },
      ]}
    />
  )

  if (isLoading) {
    return (
      <PageSection>
        {breadcrumbs}
        <Text variant="body" theme="neutral">
          Loading bundle...
        </Text>
      </PageSection>
    )
  }

  if (isError || !bundle) {
    return (
      <PageSection>
        {breadcrumbs}
        <BackLink />
        <Text variant="body" theme="neutral">
          Bundle not found.
        </Text>
      </PageSection>
    )
  }

  return (
    <PageSection>
      <PageTitle title={`Bundle | ${app?.name}`} />
      {breadcrumbs}

      <div className="flex items-start justify-between">
        <HeadingGroup>
          <BackLink className="mb-4" />
          <Text variant="base" weight="strong">
            Bundle
          </Text>
          <ID clickToCopyProps={{ copyValue: bundle.id }}>{bundle.id}</ID>
          {bundle.status_description ? (
            <Text variant="subtext" theme="neutral">
              {bundle.status_description}
            </Text>
          ) : null}
        </HeadingGroup>

        {bundle.status === 'active' ? (
          <div className="flex gap-1">
            <RegisterAirgapInstallButton bundle={bundle} />
            <DownloadBundleButton bundle={bundle} />
          </div>
        ) : null}
      </div>

      <div className="flex flex-wrap gap-8">
        <LabeledStatus
          label="Status"
          statusProps={{ status: bundle.status ?? 'unknown' }}
          tooltipProps={{
            tipContent: bundle.status_description,
            tipContentClassName: 'w-fit',
          }}
        />
        <LabeledValue label="Platform">
          <Text variant="body">{bundle.target_platform || '—'}</Text>
        </LabeledValue>
        <LabeledValue label="Size">
          <Text variant="body">
            {bundle.size ? formatBytes(bundle.size) : '—'}
          </Text>
        </LabeledValue>
        <LabeledValue label="Created">
          {bundle.created_at ? (
            <Time variant="body" time={bundle.created_at} />
          ) : (
            <Text variant="body">—</Text>
          )}
        </LabeledValue>
        {bundle.transport_checksum ? (
          <LabeledValue label="Checksum">
            <ID clickToCopyProps={{ copyValue: bundle.transport_checksum }}>
              {bundle.transport_checksum.slice(0, 12)}…
            </ID>
          </LabeledValue>
        ) : null}
      </div>

      <HeadingGroup>
        <Text variant="base" weight="strong">
          Contents
        </Text>
        <Text variant="subtext" theme="neutral">
          Artifacts packaged into this bundle.
        </Text>
      </HeadingGroup>

      <BundleContentsTable
        artifacts={bundle.artifacts ?? []}
        orgId={org?.id}
        appId={app?.id}
      />
    </PageSection>
  )
}
