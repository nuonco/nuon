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
import { getAppConfig } from '@/lib'

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

  return (
    <PageSection isScrollable>
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
