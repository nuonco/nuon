import { useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { DownloadBundleButton } from '@/components/apps/bundles/DownloadBundle'
import { ReleaseFilesContainer } from '@/components/apps/bundles/ReleaseFiles'
import { BackLink } from '@/components/common/BackLink'
import { Badge } from '@/components/common/Badge'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Loading } from '@/components/common/Loading'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Select } from '@/components/common/form/Select'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import {
  getAppRelease,
  getAppReleases,
  getAppInstalls,
  proposeAppRelease,
} from '@/lib'
import type { TAPIError, TInstall, TReleasePackage } from '@/types'
import { formatBytes } from '@/utils/string-utils'

const PENDING_STATUSES = ['preparing', 'queued', 'publishing']

const isEligibleInstall = (install: TInstall) =>
  install.management_policy?.connectivity === 'connected' &&
  install.management_policy?.release_selection === 'vendor_proposed' &&
  install.management_policy?.approval_authority === 'customer'

const ProposeReleaseModal = ({
  releaseId,
  ...props
}: IModal & { releaseId: string }) => {
  const { org } = useOrg()
  const { app } = useApp()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const [installId, setInstallId] = useState('')
  const installs = useQuery({
    queryKey: ['app-installs', org?.id, app?.id, 'release-proposal'],
    queryFn: () =>
      getAppInstalls({ appId: app!.id, orgId: org!.id, limit: 100 }),
    enabled: !!org?.id && !!app?.id,
  })
  const eligible = (installs.data?.data ?? []).filter(isEligibleInstall)
  const proposal = useMutation({
    mutationFn: () =>
      proposeAppRelease({ installId, orgId: org!.id, releaseId }),
    onSuccess: () => {
      addToast(
        <Toast heading="Release proposed" theme="success">
          <Text>
            The customer can now review the release changes and deployment plans
            in their portal.
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
  })
  const error = proposal.error as TAPIError | null

  return (
    <Modal
      heading={
        <Text variant="h3" weight="strong">
          Propose release
        </Text>
      }
      primaryActionTrigger={{
        children: proposal.isPending ? 'Proposing release' : 'Propose release',
        disabled: !installId || proposal.isPending,
        onClick: () => proposal.mutate(),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4 mb-6">
        <Text>
          The customer will see this immutable release immediately. Nuon will
          generate deployment plans, and nothing is applied until the customer
          approves them.
        </Text>
        {error ? (
          <Text theme="error">
            {error.error ?? 'Unable to propose the release.'}
          </Text>
        ) : null}
        <Select
          labelProps={{ labelText: 'Customer-managed install' }}
          placeholder={
            installs.isLoading ? 'Loading installs' : 'Select an install'
          }
          options={eligible.map((install) => ({
            value: install.id!,
            label: install.name ?? install.id!,
          }))}
          value={installId}
          onChange={setInstallId}
          searchable
        />
        {!installs.isLoading && eligible.length === 0 ? (
          <Text variant="subtext" theme="neutral">
            No connected customer-managed installs are eligible for this
            release.
          </Text>
        ) : null}
      </div>
    </Modal>
  )
}

const PortableBundles = ({ packages }: { packages: TReleasePackage[] }) => (
  <div className="flex flex-col gap-3 mt-6">
    {packages.map((pkg) => (
      <div
        key={pkg.id}
        className="border rounded-md border-cool-grey-300 dark:border-dark-grey-500 p-4 flex items-start justify-between gap-4"
      >
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-2">
            <ID clickToCopyProps={{ copyValue: pkg.id! }}>{pkg.id}</ID>
            <Badge variant="code">{pkg.target_platform}</Badge>
            <Badge>{pkg.format}</Badge>
          </div>
          <Text variant="subtext" theme="neutral">
            {pkg.archive_size
              ? formatBytes(pkg.archive_size)
              : 'Archive pending'}
            {pkg.archive_checksum ? ` · ${pkg.archive_checksum}` : ''}
          </Text>
          {pkg.status_description ? (
            <Text variant="subtext" theme="neutral">
              {pkg.status_description}
            </Text>
          ) : null}
        </div>
        <div className="flex items-center gap-3">
          <LabeledStatus
            label="Status"
            statusProps={{ status: pkg.status ?? 'unknown' }}
          />
          {pkg.status === 'active' ? (
            <DownloadBundleButton bundle={pkg} />
          ) : null}
        </div>
      </div>
    ))}
  </div>
)

export const BundleDetail = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { releaseId } = useParams()
  const { addModal } = useSurfaces()

  const {
    data: release,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ['app-release', org?.id, app?.id, releaseId],
    queryFn: () =>
      getAppRelease({ appId: app!.id, orgId: org!.id, releaseId: releaseId! }),
    enabled: !!org?.id && !!app?.id && !!releaseId,
    refetchInterval: (query) =>
      PENDING_STATUSES.includes(query.state.data?.status ?? '') ||
      query.state.data?.packages?.some((pkg) =>
        PENDING_STATUSES.includes(pkg.status ?? '')
      )
        ? 5000
        : false,
  })
  const { data: releasesResult, isLoading: isHistoryLoading } = useQuery({
    queryKey: ['app-releases', org?.id, app?.id, 'comparison-history'],
    queryFn: () =>
      getAppReleases({ appId: app!.id, orgId: org!.id, limit: 100 }),
    enabled: !!org?.id && !!app?.id && !!releaseId,
  })
  const releases = releasesResult?.data ?? []
  const releaseIndex = releases.findIndex(({ id }) => id === releaseId)
  const previousReleaseId =
    releaseIndex >= 0 ? releases[releaseIndex + 1]?.id : undefined
  const { data: previousRelease, isLoading: isPreviousLoading } = useQuery({
    queryKey: ['app-release', org?.id, app?.id, previousReleaseId],
    queryFn: () =>
      getAppRelease({
        appId: app!.id,
        orgId: org!.id,
        releaseId: previousReleaseId!,
      }),
    enabled: !!org?.id && !!app?.id && !!previousReleaseId,
  })

  const breadcrumbs = (
    <Breadcrumbs
      breadcrumbs={[
        { path: `/${org?.id}`, text: org?.name },
        { path: `/${org?.id}/apps`, text: 'Apps' },
        { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
        { path: `/${org?.id}/apps/${app?.id}/releases`, text: 'Releases' },
        {
          path: `/${org?.id}/apps/${app?.id}/releases/${releaseId}`,
          text: 'Release',
        },
      ]}
    />
  )

  if (isLoading) {
    return (
      <PageSection>
        {breadcrumbs}
        <Loading />
      </PageSection>
    )
  }
  if (isError || !release) {
    return (
      <PageSection>
        {breadcrumbs}
        <BackLink />
        <Text variant="body" theme="neutral">
          Release not found.
        </Text>
      </PageSection>
    )
  }

  return (
    <PageSection>
      <PageTitle title={`Release | ${app?.name}`} />
      {breadcrumbs}
      <HeadingGroup>
        <BackLink className="mb-4" />
        <Text variant="h3" weight="strong">
          Immutable release
        </Text>
        <ID clickToCopyProps={{ copyValue: release.id! }}>{release.id}</ID>
        {release.status_description ? (
          <Text variant="subtext" theme="neutral">
            {release.status_description}
          </Text>
        ) : null}
      </HeadingGroup>

      {org?.features?.['customer-managed-installs'] &&
      release.status === 'ready' &&
      releaseIndex === 0 ? (
        <div className="flex justify-end">
          <Button
            variant="primary"
            onClick={() =>
              addModal(<ProposeReleaseModal releaseId={release.id!} />)
            }
          >
            <Icon variant="PackageIcon" />
            Propose release
          </Button>
        </div>
      ) : null}

      <div className="flex flex-wrap gap-8">
        <LabeledStatus
          label="Status"
          statusProps={{ status: release.status ?? 'unknown' }}
        />
        <LabeledValue label="Semantic digest">
          <ID clickToCopyProps={{ copyValue: release.semantic_digest! }}>
            {release.semantic_digest}
          </ID>
        </LabeledValue>
        <LabeledValue label="Created">
          {release.created_at ? (
            <Time variant="body" time={release.created_at} />
          ) : (
            <Text>—</Text>
          )}
        </LabeledValue>
      </div>

      <Tabs
        tabs={{
          files:
            isHistoryLoading || (previousReleaseId && isPreviousLoading) ? (
              <div className="flex justify-center py-12">
                <Loading />
              </div>
            ) : (
              <ReleaseFilesContainer
                appId={app!.id}
                orgId={org!.id}
                release={release}
                previousRelease={previousRelease}
              />
            ),
          packages: <PortableBundles packages={release.packages ?? []} />,
        }}
      />
    </PageSection>
  )
}
