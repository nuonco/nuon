import { useParams, useSearchParams } from 'react-router-dom'
import { usePolling } from '@/hooks/use-polling'
import { useQuery } from '@/hooks/use-query'
import { BackToTop } from '@/components/common/BackToTop'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { ManagementDropdown } from '@/components/sandbox/management/ManagementDropdown'
import { SandboxRunsTimeline } from '@/components/sandbox/SandboxRunsTimeline'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import type { TInstallSandboxRun, TDriftedObject, TAppConfig } from '@/types'

// Old layout stuff
import { Loading, Section } from '@/components'
import { DriftedBanner } from '@/components/old/DriftedBanner'
import { AppSandboxConfig, AppSandboxVariables, Notice } from '@/components'
import { ValuesFileModal } from '@/components/old/InstallSandbox'

const LIMIT = 10

const RunsError = ({
  message = 'We encountered an issue loading your sandbox runs. Please try refreshing the page.',
  title = 'Unable to load runs',
}: {
  message?: string
  title?: string
}) => {
  return (
    <EmptyState variant="history" emptyMessage={message} emptyTitle={title} />
  )
}

const Runs = ({
  installId,
  orgId,
  offset,
}: {
  installId: string
  orgId: string
  offset: string
}) => {
  const {
    data: runs,
    error,
    isLoading,
    headers,
  } = usePolling<TInstallSandboxRun[]>({
    path: `/api/ctl-api/v1/installs/${installId}/sandbox/runs?limit=${LIMIT}&offset=${offset}`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  const pagination = {
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  if (error) {
    return <RunsError />
  }

  if (isLoading && !runs) {
    return <TimelineSkeleton eventCount={10} />
  }

  if (!runs?.length) {
    return (
      <RunsError
        title="No sandbox runs yet"
        message="Once the install is provisioned your sandbox runs will appear here."
      />
    )
  }

  return <SandboxRunsTimeline initRuns={runs} pagination={pagination} shouldPoll />
}

const SandboxConfig = ({
  appId,
  appConfigId,
  orgId,
}: {
  appId: string
  appConfigId: string
  orgId: string
}) => {
  const { data, error, isLoading } = useQuery<TAppConfig>({
    path: `/api/ctl-api/v1/apps/${appId}/configs/${appConfigId}?recurse=true`,
  })

  if (error) {
    return <Notice>{error.message || 'Unable to load sandbox config'}</Notice>
  }

  if (isLoading && !data) {
    return <Loading loadingText="Loading sandbox config..." variant="stack" />
  }

  if (!data?.sandbox) {
    return <Notice>No sandbox configuration found</Notice>
  }

  return (
    <>
      <AppSandboxConfig sandboxConfig={data.sandbox} />
      {data.sandbox.variables && (
        <AppSandboxVariables
          variables={data.sandbox.variables}
          isNotTruncated
        />
      )}
      {data.sandbox.variables_files && (
        <ValuesFileModal valuesFiles={data.sandbox.variables_files} />
      )}
    </>
  )
}

export default function InstallSandbox() {
  const { installId, orgId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()

  const offset = searchParams.get('offset') || '0'

  const { data: driftedObjects, error: driftedError } = usePolling<
    TDriftedObject[]
  >({
    path: `/api/ctl-api/v1/installs/${installId}/drifted-objects`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  if (!installId || !orgId) {
    return null
  }

  const latestSandboxRun = install?.install_sandbox_runs?.at(0)
  const driftedObject = driftedObjects?.find(
    (drifted) =>
      drifted?.['target_type'] === 'install_sandbox_run' &&
      drifted?.['target_id'] === latestSandboxRun?.id
  )

  return (
    <PageSection isScrollable className="!p-0">
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/installs`,
            text: 'Installs',
          },
          {
            path: `/${orgId}/installs/${installId}`,
            text: install?.name,
          },
          {
            path: `/${orgId}/installs/${installId}/sandbox`,
            text: 'Sandbox',
          },
        ]}
      />

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-y md:divide-x">
        <div className="md:col-span-8 divide-y flex-auto flex flex-col">
          {driftedObject ? (
            <Section className="!border-b-0 !pb-0">
              <DriftedBanner drifted={driftedObject} />
            </Section>
          ) : null}
          <Section
            actions={
              <Text variant="subtext">
                <Link href={`/${orgId}/apps/${install.app_id}`}>
                  Details
                  <Icon variant="CaretRightIcon" />
                </Link>
              </Text>
            }
            className="flex-initial"
            heading="Config"
            childrenClassName="flex flex-col gap-4"
          >
            <SandboxConfig
              appId={install?.app_id}
              appConfigId={install?.app_config_id}
              orgId={orgId}
            />
          </Section>

          <Section
            heading="Terraform workspace"
            className="flex-initial"
            childrenClassName="flex flex-col gap-4"
          >
            {install?.sandbox?.terraform_workspace ? (
              <div className="flex flex-col gap-2">
                <Text variant="subtext">Workspace ID: {install.sandbox.terraform_workspace.id}</Text>
                <Text variant="subtext">Name: {install.sandbox.terraform_workspace.name}</Text>
              </div>
            ) : (
              <Loading
                loadingText="Loading latest Terraform workspace..."
                variant="stack"
              />
            )}
          </Section>
        </div>

        <div className="divide-y flex flex-col md:col-span-4">
          <Section heading="Sandbox controls" className="flex-initial">
            <div className="flex items-center gap-4 flex-wrap">
              <ManagementDropdown />
            </div>
          </Section>
          <Section heading="Sandbox history">
            <Runs installId={installId} orgId={orgId} offset={offset} />
          </Section>
        </div>
      </div>

      <BackToTop />
    </PageSection>
  )
}