import { useQuery } from '@tanstack/react-query'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { InstallStacksTable, InstallStacksTableSkeleton } from '@/components/stacks/InstallStacksTable'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { getAppConfig, getAppConfigs } from '@/lib'
import { hasNewerAppConfig, hasStackConfigChanged } from '@/utils/app-utils'

export const Stacks = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: configResult, isLoading: isLoadingConfig } = useQuery({
    queryKey: [
      'app-config',
      org?.id,
      install?.app_id,
      install?.app_config_id,
      'recurse',
    ],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: install.app_config_id,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_config_id,
  })

  const config = configResult

  const { data: latestConfigs } = useQuery({
    queryKey: ['app-configs', org?.id, install?.app_id, 'latest'],
    queryFn: () =>
      getAppConfigs({ orgId: org.id, appId: install.app_id, limit: 1, offset: 0 }),
    enabled: !!org?.id && !!install?.app_id,
  })
  const latestConfigSummary = latestConfigs?.[0]
  const newerAppConfig = hasNewerAppConfig(latestConfigSummary, install)

  const { data: latestFullConfig } = useQuery({
    queryKey: ['app-config', org?.id, install?.app_id, latestConfigSummary?.id, 'recurse'],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: latestConfigSummary!.id!,
        recurse: true,
      }),
    enabled: newerAppConfig && !!latestConfigSummary?.id,
  })

  const stackChanged = hasStackConfigChanged(config, latestFullConfig)

  return (
    <PageSection>
      <PageTitle title={`Stacks | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/stacks`,
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

      {stackChanged && (
        <Banner theme="info">
          <div className="flex items-center gap-8">
            <div className="flex flex-col">
              <Text weight="strong">New stack config available</Text>
              <Text variant="subtext" theme="neutral">
                A newer stack config (v{latestFullConfig?.version}) is available. This install
                is using v{config?.version}.
              </Text>
            </div>
            <Button
              className="ml-auto"
              href={`/${org.id}/apps/${install.app_id}`}
              variant="secondary"
            >
              View latest config
            </Button>
          </div>
        </Banner>
      )}

      {isLoadingConfig ? (
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
      ) : config ? (
        <Card>
          <Text weight="strong">Current stack config</Text>
          <div className="grid grid-cols-6 gap-3">
            <LabeledValue label="App config version">
              {config?.version?.toString()}
            </LabeledValue>
            <LabeledValue label="Stack type">
              {config?.stack?.type}
            </LabeledValue>
            <LabeledValue label="Stack name">
              {config?.stack?.name}
            </LabeledValue>
            {config?.stack?.runner_nested_template_url ? (
              <LabeledValue
                className="col-span-6"
                label="Runner nested template URL"
              >
                <Text variant="subtext">
                  <Link
                    href={config.stack.runner_nested_template_url}
                    isExternal
                  >
                    {config.stack.runner_nested_template_url}
                  </Link>
                </Text>
              </LabeledValue>
            ) : null}
            {config?.stack?.vpc_nested_template_url ? (
              <LabeledValue
                className="col-span-6"
                label="VPC nested template URL"
              >
                <Text variant="subtext">
                  <Link href={config.stack.vpc_nested_template_url} isExternal>
                    {config.stack.vpc_nested_template_url}
                  </Link>
                </Text>
              </LabeledValue>
            ) : null}
          </div>
        </Card>
      ) : (
        <Card>
          <EmptyState
            emptyTitle="No stack config"
            emptyMessage="Unable to load this install stack config"
            variant="table"
          />
        </Card>
      )}

      <div className="flex flex-col gap-4">
        <Text weight="strong">Install stack versions</Text>
        <InstallStacksTable shouldPoll />
      </div>
    </PageSection>
  )
}
