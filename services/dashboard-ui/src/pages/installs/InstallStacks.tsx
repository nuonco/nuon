import { useParams } from 'react-router-dom'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { usePolling } from '@/hooks/use-polling'
import { useQuery } from '@/hooks/use-query'
import { BackToTop } from '@/components/common/BackToTop'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { InstallStacksTable as Table, InstallStacksTableSkeleton } from '@/components/stacks/InstallStacksTable'
import type { TAppConfig, TInstallStack } from '@/types'

const StackConfig = ({ install, orgId }: { install: any; orgId: string }) => {
  const { data: config, error, isLoading } = useQuery<TAppConfig>({
    path: `/api/ctl-api/v1/apps/${install?.app_id}/configs/${install?.app_config_id}?recurse=true`,
  })

  if (isLoading) {
    return (
      <Card>
        <Skeleton width="135px" height="24px" />
        <div className="grid grid-cols-6 gap-3">
          <LabeledValue label={<Skeleton height="17px" width="100px" />}>
            <Skeleton height="17px" width="20px" />
          </LabeledValue>
          <LabeledValue label={<Skeleton height="17px" width="60px" />}>
            <Skeleton height="17px" width="110px" />
          </LabeledValue>
          <LabeledValue label={<Skeleton height="17px" width="65px" />}>
            <Skeleton height="17px" width="180px" />
          </LabeledValue>
        </div>
      </Card>
    )
  }

  if (!config && error) {
    return (
      <EmptyState
        emptyTitle="Could not load stack config"
        emptyMessage="Unable to load the stack config for this install"
      />
    )
  }

  return (
    <Card>
      <Text weight="strong">Current stack config</Text>

      <div className="grid grid-cols-6 gap-3">
        <LabeledValue label="App config version">
          {config?.version?.toString()}
        </LabeledValue>

        <LabeledValue label="Stack type">{config?.stack?.type}</LabeledValue>

        <LabeledValue label="Stack name">{config?.stack?.name}</LabeledValue>

        {config?.stack?.runner_nested_template_url ? (
          <LabeledValue
            className="col-span-6"
            label="Runner nested template URL"
          >
            <Text variant="subtext">
              <Link href={config?.stack?.runner_nested_template_url} isExternal>
                {config?.stack?.runner_nested_template_url}
              </Link>
            </Text>
          </LabeledValue>
        ) : null}

        {config?.stack?.vpc_nested_template_url ? (
          <LabeledValue className="col-span-6" label="VPC nested template URL">
            <Text variant="subtext">
              <Link href={config?.stack?.vpc_nested_template_url} isExternal>
                {config?.stack?.vpc_nested_template_url}
              </Link>
            </Text>
          </LabeledValue>
        ) : null}
      </div>
    </Card>
  )
}

const InstallStacksTableWrapper = ({ installId, orgId }: { installId: string; orgId: string }) => {
  const { data: stack, error, isLoading, headers } = usePolling<TInstallStack>({
    path: `/api/ctl-api/v1/installs/${installId}/stack`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  const pagination = {
    limit: Number(headers?.['x-nuon-page-limit'] ?? 10),
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  if (isLoading && !stack) {
    return <InstallStacksTableSkeleton />
  }

  if (error) {
    return (
      <Banner theme="error">
        Can&apos;t load install stacks: {error?.error}
      </Banner>
    )
  }

  if (!stack) {
    return <InstallStacksTableSkeleton />
  }

  return <Table stack={stack} pagination={pagination} shouldPoll />
}

export default function InstallStacks() {
  const { orgId, installId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  const containerId = 'stack-page'
  return (
    <PageSection id={containerId} className="!pb-24" isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name || '',
          },
          {
            path: `/${orgId}/installs`,
            text: 'Installs',
          },
          {
            path: `/${orgId}/installs/${installId}`,
            text: install?.name || '',
          },
          {
            path: `/${orgId}/installs/${installId}/stacks`,
            text: 'Stacks',
          },
        ]}
      />

      <HeadingGroup>
        <Text variant="base" weight="strong">
          Install stacks
        </Text>
        <Text variant="subtext" theme="neutral">
          View your install stack config and versions below.
        </Text>
      </HeadingGroup>

      <StackConfig orgId={orgId || ''} install={install} />

      <div className="flex flex-col gap-4">
        <Text weight="strong">Install stack versions</Text>
        <InstallStacksTableWrapper installId={installId || ''} orgId={orgId || ''} />
      </div>
      <BackToTop containerId={containerId} />
    </PageSection>
  )
}