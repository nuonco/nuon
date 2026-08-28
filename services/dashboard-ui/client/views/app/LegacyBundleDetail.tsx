import { useMutation, useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { BundleContentsTable } from '@/components/apps/bundles/BundleContentsTable'
import { DownloadBundle } from '@/components/apps/bundles/DownloadBundle/DownloadBundle'
import { BackLink } from '@/components/common/BackLink'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import {
  createCustomerManagedBundleDownloadGrant,
  getCustomerManagedBundle,
} from '@/lib'
import { formatBytes } from '@/utils/string-utils'

const PENDING_STATUSES = ['queued', 'publishing']

export const LegacyBundleDetail = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { bundleId } = useParams()
  const { data: bundle, isLoading } = useQuery({
    queryKey: ['customer-managed-bundle', org?.id, app?.id, bundleId],
    queryFn: () =>
      getCustomerManagedBundle({
        appId: app!.id,
        bundleId: bundleId!,
        orgId: org!.id,
      }),
    enabled: !!org?.id && !!app?.id && !!bundleId,
    refetchInterval: (query) =>
      PENDING_STATUSES.includes(query.state.data?.status ?? '') ? 5000 : false,
  })
  const { mutate: download, isPending: isDownloadPending } = useMutation({
    mutationFn: () =>
      createCustomerManagedBundleDownloadGrant({
        appId: app!.id,
        bundleId: bundleId!,
        orgId: org!.id,
      }),
    onSuccess: (grant) => {
      if (grant.url) window.location.assign(grant.url)
    },
  })

  const breadcrumbs = (
    <Breadcrumbs
      breadcrumbs={[
        { path: `/${org?.id}`, text: org?.name },
        { path: `/${org?.id}/apps`, text: 'Apps' },
        { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
        {
          path: `/${org?.id}/apps/${app?.id}/bundles/${bundleId}`,
          text: 'Bundle',
        },
      ]}
    />
  )

  if (isLoading)
    return (
      <PageSection>
        {breadcrumbs}
        <Loading />
      </PageSection>
    )
  if (!bundle) {
    return (
      <PageSection>
        {breadcrumbs}
        <BackLink />
        <Text>Bundle not found.</Text>
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
          <Text variant="h3" weight="strong">
            Published bundle
          </Text>
          <ID clickToCopyProps={{ copyValue: bundle.id }}>{bundle.id}</ID>
        </HeadingGroup>
        {bundle.status === 'active' ? (
          <DownloadBundle
            isPending={isDownloadPending}
            onClick={() => download()}
          />
        ) : null}
      </div>
      <div className="flex flex-wrap gap-8">
        <LabeledStatus
          label="Status"
          statusProps={{ status: bundle.status ?? 'unknown' }}
        />
        <LabeledValue label="Platform">
          <Text>{bundle.target_platform || '—'}</Text>
        </LabeledValue>
        <LabeledValue label="Size">
          <Text>{bundle.size ? formatBytes(bundle.size) : '—'}</Text>
        </LabeledValue>
        <LabeledValue label="Created">
          {bundle.created_at ? (
            <Time variant="body" time={bundle.created_at} />
          ) : (
            <Text>—</Text>
          )}
        </LabeledValue>
      </div>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Bundle contents
        </Text>
        <Text variant="subtext" theme="neutral">
          Legacy package contents retained for registered install snapshots.
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
